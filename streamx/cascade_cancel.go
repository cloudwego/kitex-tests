// Copyright 2026 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package streamx

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt/streamcall"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/codes"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/status"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi/testpbcancelservice"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/testcancelservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
)

const (
	cascadeCancelDirect          = "direct"
	cascadeCancelWithCancel      = "with_cancel"
	cascadeCancelWithTimeout     = "with_timeout"
	cascadeCancelWithCancelCause = "with_cancel_cause"
	cascadeCancelRemoteStatus    = "remote_status_canceled"
)

type cascadeCancelTestCase struct {
	downstreamReady chan struct{}
	proxyReady      chan struct{}
	result          chan error
}

type cascadeCancelTestCases struct {
	mu    sync.RWMutex
	cases map[string]*cascadeCancelTestCase
}

func newCascadeCancelTestCases() *cascadeCancelTestCases {
	return &cascadeCancelTestCases{cases: make(map[string]*cascadeCancelTestCase)}
}

func (c *cascadeCancelTestCases) add(name string) *cascadeCancelTestCase {
	tc := &cascadeCancelTestCase{
		downstreamReady: make(chan struct{}, 1),
		proxyReady:      make(chan struct{}, 1),
		result:          make(chan error, 1),
	}
	c.mu.Lock()
	c.cases[name] = tc
	c.mu.Unlock()
	return tc
}

func (c *cascadeCancelTestCases) get(name string) (*cascadeCancelTestCase, bool) {
	c.mu.RLock()
	tc, ok := c.cases[name]
	c.mu.RUnlock()
	return tc, ok
}

func wrapCascadeCancelContext(ctx context.Context, mode string) (context.Context, func()) {
	switch mode {
	case cascadeCancelDirect, cascadeCancelRemoteStatus:
		return ctx, func() {}
	case cascadeCancelWithCancel:
		return context.WithCancel(ctx)
	case cascadeCancelWithTimeout:
		return context.WithTimeout(ctx, 30*time.Second)
	case cascadeCancelWithCancelCause:
		wrapped, cancel := context.WithCancelCause(ctx)
		return wrapped, func() { cancel(nil) }
	default:
		return ctx, func() {}
	}
}

func waitCascadeCancelResult(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for cascade cancel result")
		return nil
	}
}

func assertCascadeCancelStatus(t *testing.T, err error, wantCascade bool) {
	t.Helper()
	test.Assert(t, err != nil)
	st, ok := status.FromError(err)
	test.Assert(t, ok, err)
	test.Assert(t, st.Code() == codes.Canceled, st.Code())
	test.Assert(t, st.IsCascadeCancel() == wantCascade, st)
}

type cascadeCancelClient[Req, Res any] interface {
	CancelBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[Req, Res], err error)
}

func commonTestCascadeCancel[Req, Res any](
	t *testing.T,
	testCases *cascadeCancelTestCases,
	cli cascadeCancelClient[Req, Res],
	newReq func(string) *Req,
) {
	for _, mode := range []string{
		cascadeCancelDirect,
		cascadeCancelWithCancel,
		cascadeCancelWithTimeout,
		cascadeCancelWithCancelCause,
	} {
		t.Run(mode, func(t *testing.T) {
			tc := testCases.add(mode)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			upstream, err := cli.CancelBidi(ctx)
			test.Assert(t, err == nil, err)
			err = upstream.Send(upstream.Context(), newReq(mode))
			test.Assert(t, err == nil, err)

			select {
			case <-tc.proxyReady:
			case err = <-tc.result:
				t.Fatalf("proxy failed before downstream was ready: %v", err)
			case <-time.After(1 * time.Second):
				t.Fatal("timeout waiting for downstream stream")
			}

			// A actively cancels the A -> B stream after B -> C is established.
			cancel()
			_, err = upstream.Recv(upstream.Context())
			assertCascadeCancelStatus(t, err, false)
			assertCascadeCancelStatus(t, waitCascadeCancelResult(t, tc.result), true)
		})
	}

	t.Run(cascadeCancelRemoteStatus, func(t *testing.T) {
		tc := testCases.add(cascadeCancelRemoteStatus)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		upstream, err := cli.CancelBidi(ctx)
		test.Assert(t, err == nil, err)
		err = upstream.Send(upstream.Context(), newReq(cascadeCancelRemoteStatus))
		test.Assert(t, err == nil, err)

		// C returns a gRPC status whose numeric code is 1. B must retain the
		// status code without mistaking a remote business error for a cascade.
		assertCascadeCancelStatus(t, waitCascadeCancelResult(t, tc.result), false)
	})
}

func runCascadeCancelServer(svr server.Server) {
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
}

type pbCascadeCancelDownstream struct {
	pbapi.TestPbCancelService
	testCases *cascadeCancelTestCases
}

func (h *pbCascadeCancelDownstream) CancelPbBidi(ctx context.Context, stream pbapi.TestPbCancelService_CancelPbBidiServer) error {
	req, err := stream.Recv(ctx)
	if err != nil {
		return err
	}
	tc, ok := h.testCases.get(req.Message)
	if !ok {
		return fmt.Errorf("unknown cascade cancel test case %q", req.Message)
	}
	tc.downstreamReady <- struct{}{}
	if req.Message == cascadeCancelRemoteStatus {
		return status.Err(codes.Canceled, "downstream returned gRPC status code 1")
	}
	<-ctx.Done()
	return nil
}

type pbCascadeCancelProxy struct {
	pbapi.TestPbCancelService
	testCases  *cascadeCancelTestCases
	downstream testpbcancelservice.Client
}

func (h *pbCascadeCancelProxy) CancelPbBidi(ctx context.Context, stream pbapi.TestPbCancelService_CancelPbBidiServer) error {
	req, err := stream.Recv(ctx)
	if err != nil {
		return err
	}
	tc, ok := h.testCases.get(req.Message)
	if !ok {
		return fmt.Errorf("unknown cascade cancel test case %q", req.Message)
	}

	downstreamCtx, cancel := wrapCascadeCancelContext(ctx, req.Message)
	defer cancel()
	downstream, err := h.downstream.CancelPbBidi(downstreamCtx)
	if err == nil {
		err = downstream.Send(downstreamCtx, req)
	}
	if err != nil {
		tc.result <- err
		return nil
	}
	if req.Message != cascadeCancelRemoteStatus {
		<-tc.downstreamReady
		tc.proxyReady <- struct{}{}
	}
	_, err = downstream.Recv(downstreamCtx)
	tc.result <- err
	return nil
}

// TestPbCascadeCancel verifies the A -> B -> C cascade cancel behavior with
// StreamX's protobuf-generated streaming API.
func TestPbCascadeCancel(t *testing.T) {
	testCases := newCascadeCancelTestCases()

	downstreamLn := serverutils.Listen()
	downstreamAddr := downstreamLn.Addr().String()
	downstreamSvr := testpbcancelservice.NewServer(
		&pbCascadeCancelDownstream{testCases: testCases},
		server.WithListener(downstreamLn),
		server.WithExitWaitTime(10*time.Millisecond),
	)
	runCascadeCancelServer(downstreamSvr)
	defer downstreamSvr.Stop()

	downstreamCli := testpbcancelservice.MustNewClient("cascade-cancel-downstream",
		client.WithHostPorts(downstreamAddr),
		client.WithTransportProtocol(transport.GRPCStreaming),
	)
	proxyLn := serverutils.Listen()
	proxyAddr := proxyLn.Addr().String()
	proxySvr := testpbcancelservice.NewServer(
		&pbCascadeCancelProxy{testCases: testCases, downstream: downstreamCli},
		server.WithListener(proxyLn),
		server.WithExitWaitTime(10*time.Millisecond),
	)
	runCascadeCancelServer(proxySvr)
	defer proxySvr.Stop()

	cli := testpbcancelservice.MustNewClient("cascade-cancel-proxy",
		client.WithHostPorts(proxyAddr),
		client.WithTransportProtocol(transport.GRPCStreaming),
	)
	commonTestCascadeCancel[pbapi.MockReq, pbapi.MockResp](t, testCases, pbCancelClient{cli}, func(mode string) *pbapi.MockReq {
		return &pbapi.MockReq{Message: mode}
	})
}

type thriftCascadeCancelDownstream struct {
	tenant.TestCancelService
	testCases *cascadeCancelTestCases
}

func (h *thriftCascadeCancelDownstream) CancelBidi(ctx context.Context, stream tenant.TestCancelService_CancelBidiServer) error {
	req, err := stream.Recv(ctx)
	if err != nil {
		return err
	}
	tc, ok := h.testCases.get(req.Msg)
	if !ok {
		return fmt.Errorf("unknown cascade cancel test case %q", req.Msg)
	}
	tc.downstreamReady <- struct{}{}
	if req.Msg == cascadeCancelRemoteStatus {
		return status.Err(codes.Canceled, "downstream returned gRPC status code 1")
	}
	<-ctx.Done()
	return nil
}

type thriftCascadeCancelProxy struct {
	tenant.TestCancelService
	testCases  *cascadeCancelTestCases
	downstream testcancelservice.Client
}

func (h *thriftCascadeCancelProxy) CancelBidi(ctx context.Context, stream tenant.TestCancelService_CancelBidiServer) error {
	req, err := stream.Recv(ctx)
	if err != nil {
		return err
	}
	tc, ok := h.testCases.get(req.Msg)
	if !ok {
		return fmt.Errorf("unknown cascade cancel test case %q", req.Msg)
	}

	downstreamCtx, cancel := wrapCascadeCancelContext(ctx, req.Msg)
	defer cancel()
	downstream, err := h.downstream.CancelBidi(downstreamCtx)
	if err == nil {
		err = downstream.Send(downstreamCtx, req)
	}
	if err != nil {
		tc.result <- err
		return nil
	}
	if req.Msg != cascadeCancelRemoteStatus {
		<-tc.downstreamReady
		tc.proxyReady <- struct{}{}
	}
	_, err = downstream.Recv(downstreamCtx)
	tc.result <- err
	return nil
}

// TestThriftCascadeCancel verifies the A -> B -> C cascade cancel behavior
// when StreamX's thrift-generated streaming API uses the gRPC transport.
func TestThriftCascadeCancel(t *testing.T) {
	testCases := newCascadeCancelTestCases()

	downstreamLn := serverutils.Listen()
	downstreamAddr := downstreamLn.Addr().String()
	downstreamSvr := testcancelservice.NewServer(
		&thriftCascadeCancelDownstream{testCases: testCases},
		server.WithListener(downstreamLn),
		server.WithExitWaitTime(10*time.Millisecond),
	)
	runCascadeCancelServer(downstreamSvr)
	defer downstreamSvr.Stop()

	downstreamCli := testcancelservice.MustNewClient("cascade-cancel-downstream",
		client.WithHostPorts(downstreamAddr),
		client.WithTransportProtocol(transport.GRPCStreaming),
	)
	proxyLn := serverutils.Listen()
	proxyAddr := proxyLn.Addr().String()
	proxySvr := testcancelservice.NewServer(
		&thriftCascadeCancelProxy{testCases: testCases, downstream: downstreamCli},
		server.WithListener(proxyLn),
		server.WithExitWaitTime(10*time.Millisecond),
	)
	runCascadeCancelServer(proxySvr)
	defer proxySvr.Stop()

	cli := testcancelservice.MustNewClient("cascade-cancel-proxy",
		client.WithHostPorts(proxyAddr),
		client.WithTransportProtocol(transport.GRPCStreaming),
	)
	commonTestCascadeCancel[tenant.EchoRequest, tenant.EchoResponse](t, testCases, thriftCancelClient{cli}, func(mode string) *tenant.EchoRequest {
		return &tenant.EchoRequest{Msg: mode}
	})
}
