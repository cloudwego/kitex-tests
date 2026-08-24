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

package thrift_streaming

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/codes"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/status"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb"
	grpcpbservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb/pbservice"
)

const (
	legacyCascadeCancelDirect          = "direct"
	legacyCascadeCancelWithCancel      = "with_cancel"
	legacyCascadeCancelWithTimeout     = "with_timeout"
	legacyCascadeCancelWithCancelCause = "with_cancel_cause"
	legacyCascadeCancelRemoteStatus    = "remote_status_canceled"
)

type legacyCascadeCancelTestCase struct {
	downstreamReady chan struct{}
	proxyReady      chan struct{}
	result          chan error
}

type legacyCascadeCancelTestCases struct {
	mu    sync.RWMutex
	cases map[string]*legacyCascadeCancelTestCase
}

func newLegacyCascadeCancelTestCases() *legacyCascadeCancelTestCases {
	return &legacyCascadeCancelTestCases{cases: make(map[string]*legacyCascadeCancelTestCase)}
}

func (c *legacyCascadeCancelTestCases) add(name string) *legacyCascadeCancelTestCase {
	tc := &legacyCascadeCancelTestCase{
		downstreamReady: make(chan struct{}, 1),
		proxyReady:      make(chan struct{}, 1),
		result:          make(chan error, 1),
	}
	c.mu.Lock()
	c.cases[name] = tc
	c.mu.Unlock()
	return tc
}

func (c *legacyCascadeCancelTestCases) get(name string) (*legacyCascadeCancelTestCase, bool) {
	c.mu.RLock()
	tc, ok := c.cases[name]
	c.mu.RUnlock()
	return tc, ok
}

func wrapLegacyCascadeCancelContext(ctx context.Context, mode string) (context.Context, func()) {
	switch mode {
	case legacyCascadeCancelDirect, legacyCascadeCancelRemoteStatus:
		return ctx, func() {}
	case legacyCascadeCancelWithCancel:
		return context.WithCancel(ctx)
	case legacyCascadeCancelWithTimeout:
		return context.WithTimeout(ctx, 30*time.Second)
	case legacyCascadeCancelWithCancelCause:
		wrapped, cancel := context.WithCancelCause(ctx)
		return wrapped, func() { cancel(nil) }
	default:
		return ctx, func() {}
	}
}

type legacyCascadeCancelDownstream struct {
	grpc_pb.PBService
	testCases *legacyCascadeCancelTestCases
}

func (h *legacyCascadeCancelDownstream) Echo(stream grpc_pb.PBService_EchoServer) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	tc, ok := h.testCases.get(req.Message)
	if !ok {
		return fmt.Errorf("unknown cascade cancel test case %q", req.Message)
	}
	tc.downstreamReady <- struct{}{}
	if req.Message == legacyCascadeCancelRemoteStatus {
		return status.Err(codes.Canceled, "downstream returned gRPC status code 1")
	}
	<-stream.Context().Done()
	return nil
}

type legacyCascadeCancelProxy struct {
	grpc_pb.PBService
	testCases  *legacyCascadeCancelTestCases
	downstream grpcpbservice.Client
}

func (h *legacyCascadeCancelProxy) Echo(stream grpc_pb.PBService_EchoServer) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	tc, ok := h.testCases.get(req.Message)
	if !ok {
		return fmt.Errorf("unknown cascade cancel test case %q", req.Message)
	}

	downstreamCtx, cancel := wrapLegacyCascadeCancelContext(stream.Context(), req.Message)
	defer cancel()
	downstream, err := h.downstream.Echo(downstreamCtx)
	if err == nil {
		err = downstream.Send(req)
	}
	if err != nil {
		tc.result <- err
		return nil
	}
	if req.Message != legacyCascadeCancelRemoteStatus {
		<-tc.downstreamReady
		tc.proxyReady <- struct{}{}
	}
	_, err = downstream.Recv()
	tc.result <- err
	return nil
}

func waitLegacyCascadeCancelResult(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for cascade cancel result")
		return nil
	}
}

func assertLegacyCascadeCancelStatus(t *testing.T, err error, wantCascade bool) {
	t.Helper()
	test.Assert(t, err != nil)
	st, ok := status.FromError(err)
	test.Assert(t, ok, err)
	test.Assert(t, st.Code() == codes.Canceled, st.Code())
	test.Assert(t, st.IsCascadeCancel() == wantCascade, st)
}

func TestGRPCCascadeCancel(t *testing.T) {
	testCases := newLegacyCascadeCancelTestCases()

	downstreamLn := serverutils.Listen()
	downstreamAddr := downstreamLn.Addr().String()
	downstreamSvr := RunGRPCPBServer(&legacyCascadeCancelDownstream{testCases: testCases}, downstreamLn)
	defer downstreamSvr.Stop()
	downstreamCli := grpcpbservice.MustNewClient("cascade-cancel-downstream", client.WithHostPorts(downstreamAddr))

	proxyLn := serverutils.Listen()
	proxyAddr := proxyLn.Addr().String()
	proxySvr := RunGRPCPBServer(&legacyCascadeCancelProxy{testCases: testCases, downstream: downstreamCli}, proxyLn)
	defer proxySvr.Stop()
	cli := grpcpbservice.MustNewClient("cascade-cancel-proxy", client.WithHostPorts(proxyAddr))

	for _, mode := range []string{
		legacyCascadeCancelDirect,
		legacyCascadeCancelWithCancel,
		legacyCascadeCancelWithTimeout,
		legacyCascadeCancelWithCancelCause,
	} {
		t.Run(mode, func(t *testing.T) {
			tc := testCases.add(mode)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			upstream, err := cli.Echo(ctx)
			test.Assert(t, err == nil, err)
			err = upstream.Send(&grpc_pb.Request{Message: mode})
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
			_, err = upstream.Recv()
			assertLegacyCascadeCancelStatus(t, err, false)
			assertLegacyCascadeCancelStatus(t, waitLegacyCascadeCancelResult(t, tc.result), true)
		})
	}

	t.Run(legacyCascadeCancelRemoteStatus, func(t *testing.T) {
		tc := testCases.add(legacyCascadeCancelRemoteStatus)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		upstream, err := cli.Echo(ctx)
		test.Assert(t, err == nil, err)
		err = upstream.Send(&grpc_pb.Request{Message: legacyCascadeCancelRemoteStatus})
		test.Assert(t, err == nil, err)

		// C returns a gRPC status whose numeric code is 1. B must retain the
		// status code without mistaking a remote business error for a cascade.
		assertLegacyCascadeCancelStatus(t, waitLegacyCascadeCancelResult(t, tc.result), false)
	})
}
