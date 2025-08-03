// Copyright 2025 CloudWeGo Authors
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

package streamx_thrift

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/testcancelservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/streamx"
)

const (
	clientStreamingWithoutSendingAnyReq    string = "client_streaming_without_sending_any_req"
	clientStreamingDuringNormalInteraction string = "client_streaming_during_normal_interaction"
	clientStreamingDeferCancel             string = "client_streaming_defer_cancel"

	serverStreamingRemoteRespondingSlowly  string = "server_streaming_remote_responding_slowly"
	serverStreamingDuringNormalInteraction string = "server_streaming_during_normal_interaction"
	serverStreamingDeferCancel             string = "server_streaming_defer_cancel"

	bidiStreamingIndependentSendRecv string = "bidi_streaming_independent_send_recv"

	scenarioKey       = "scenario"
	bizSpecialMessage = "biz_special_message"
)

func runTestCancelServer(listenAddr string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", listenAddr)
	svr := testcancelservice.NewServer(&cancelImpl{sendCh: cancelSendChan, tChan: cancelTChan},
		server.WithServiceAddr(addr),
		server.WithExitWaitTime(1*time.Second),
		server.WithTracer(streamx.NewTracer()),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler),
	)
	go func() {
		if err := svr.Run(); err != nil {
			println(err)
		}
	}()
	return svr
}

func runTestCancelProxyServer(listenAddr string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", listenAddr)
	svr := testcancelservice.NewServer(&cancelImpl{},
		server.WithServiceAddr(addr),
		server.WithExitWaitTime(1*time.Second),
		server.WithTracer(streamx.NewTracer()),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler),
	)
	go func() {
		if err := svr.Run(); err != nil {
			println(err)
		}
	}()
	return svr
}

type cancelImpl struct {
	sendCh chan struct{}
	tChan  chan *testing.T
}

func (s cancelImpl) CancelBidi(ctx context.Context, stream tenant.TestCancelService_CancelBidiServer) (err error) {
	defer func() {
		cancelFinishChan <- struct{}{}
	}()
	t := <-s.tChan
	scenario, ok := getScenario(ctx)
	test.Assert(t, ok)
	switch scenario {
	case bidiStreamingIndependentSendRecv:
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for {
				resp, rErr := stream.Recv(ctx)
				if rErr == nil {
					test.Assert(t, resp.Msg == bidiStreamingIndependentSendRecv, resp)
					continue
				}
				verifyServerSideCancelCase(t, ctx, rErr)
				break
			}
		}()
		go func() {
			defer wg.Done()
			for {
				sErr := stream.Send(ctx, &tenant.EchoResponse{Msg: bidiStreamingIndependentSendRecv})
				if sErr == nil {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				verifyServerSideCancelCase(t, ctx, sErr)
				break
			}
		}()
		wg.Wait()
	}

	return nil
}

func (s cancelImpl) CancelClient(ctx context.Context, stream tenant.TestCancelService_CancelClientServer) (err error) {
	defer func() {
		cancelFinishChan <- struct{}{}
	}()
	t := <-s.tChan
	scenario, ok := getScenario(ctx)
	test.Assert(t, ok)
	switch scenario {
	case clientStreamingWithoutSendingAnyReq:
		// the first Recv result would be err
		_, rErr := stream.Recv(ctx)
		test.Assert(t, rErr != nil)
		t.Logf("scenario[%s] Recv err: %v", scenario, rErr)
		verifyServerSideCancelCase(t, ctx, rErr)
	case clientStreamingDuringNormalInteraction:
		for {
			_, rErr := stream.Recv(ctx)
			if rErr == nil {
				continue
			}
			t.Logf("scenario[%s] Recv err: %v", scenario, rErr)
			verifyServerSideCancelCase(t, ctx, rErr)
			break
		}
	case clientStreamingDeferCancel:
		for {
			_, rErr := stream.Recv(ctx)
			if rErr == nil {
				continue
			}
			t.Logf("scenario[%s] Recv err: %v", scenario, rErr)
			verifyServerSideCancelCase(t, ctx, rErr)
			break
		}
	}
	return nil
}

func (s cancelImpl) CancelServer(ctx context.Context, req *tenant.EchoRequest, stream tenant.TestCancelService_CancelServerServer) (err error) {
	defer func() {
		cancelFinishChan <- struct{}{}
	}()
	t := <-s.tChan
	scenario, ok := getScenario(ctx)
	test.Assert(t, ok)
	switch scenario {
	case serverStreamingRemoteRespondingSlowly:
		<-s.sendCh
		// in case Rst Frame has not been processed due to dispatch or other reasons
		// we start a loop to Send until returning err
		for {
			resp := &tenant.EchoResponse{Msg: serverStreamingRemoteRespondingSlowly}
			sErr := stream.Send(ctx, resp)
			if sErr == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			t.Logf("scenario[%s] Recv err: %v", scenario, sErr)
			verifyServerSideCancelCase(t, ctx, sErr)
			break
		}
	case serverStreamingDuringNormalInteraction:
		for {
			resp := &tenant.EchoResponse{Msg: serverStreamingDuringNormalInteraction}
			sErr := stream.Send(ctx, resp)
			if sErr == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			klog.CtxErrorf(ctx, "scenario[%s] Recv err: %v", scenario, sErr)
			verifyServerSideCancelCase(t, ctx, sErr)
			break
		}
	case serverStreamingDeferCancel:
		for i := 0; i < 5; i++ {
			resp := &tenant.EchoResponse{Msg: serverStreamingDeferCancel}
			sErr := stream.Send(ctx, resp)
			test.Assert(t, sErr == nil, sErr)
		}
		specialResp := &tenant.EchoResponse{Msg: bizSpecialMessage}
		sErr := stream.Send(ctx, specialResp)
		test.Assert(t, sErr == nil, sErr)
		// waiting for remote service cancel
		for {
			resp := &tenant.EchoResponse{Msg: serverStreamingDeferCancel}
			sErr = stream.Send(ctx, resp)
			if sErr == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			t.Logf("scenario[%s] Recv err: %v", scenario, sErr)
			verifyServerSideCancelCase(t, ctx, sErr)
			break
		}
	}
	return nil
}

func setScenario(ctx context.Context, scenario string) context.Context {
	return metainfo.WithValue(ctx, scenarioKey, scenario)
}

func getScenario(ctx context.Context) (string, bool) {
	return metainfo.GetValue(ctx, scenarioKey)
}

func verifyClientSideCancelCase(t *testing.T, ctx context.Context, err error) {
	verifyCancelContext(t, ctx)
	verifyClientSideCancelErr(t, err)
}

func verifyServerSideCancelCase(t *testing.T, ctx context.Context, err error) {
	verifyCancelContext(t, ctx)
	verifyServerSideCancelErr(t, err)
}

func verifyCancelContext(t *testing.T, ctx context.Context) {
	var ctxDone bool
	select {
	case <-ctx.Done():
		ctxDone = true
	default:
	}
	test.Assert(t, ctxDone)
}

func verifyClientSideCancelErr(t *testing.T, err error) {
	test.Assert(t, errors.Is(err, kerrors.ErrStreamingCanceled), err)
}

func verifyServerSideCancelErr(t *testing.T, err error) {
	test.Assert(t, errors.Is(err, kerrors.ErrStreamingCanceled), err)
	test.Assert(t, strings.Contains(err.Error(), "canceled path"), err)
}

func TestCancel(t *testing.T) {
	cliSvcName := "client-side cancel service"
	cli := testcancelservice.MustNewClient("service",
		client.WithHostPorts(thriftTestCancelAddr),
		client.WithTransportProtocol(transport.TTHeaderStreaming),
		client.WithMetaHandler(transmeta.ClientHTTP2Handler),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: cliSvcName,
		}),
	)
	t.Run(clientStreamingWithoutSendingAnyReq, func(t *testing.T) {
		defer func() {
			<-cancelFinishChan
		}()
		cancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, clientStreamingWithoutSendingAnyReq)
		cliSt, err := cli.CancelClient(ctx)
		test.Assert(t, err == nil, err)
		cancel()
		// todo: modify this timeout
		time.Sleep(6 * time.Second)
		err = cliSt.Send(cliSt.Context(), &tenant.EchoRequest{Msg: clientStreamingWithoutSendingAnyReq})
		verifyClientSideCancelCase(t, cliSt.Context(), err)
	})
	t.Run(clientStreamingDuringNormalInteraction, func(t *testing.T) {
		defer func() {
			<-cancelFinishChan
		}()
		cancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, clientStreamingDuringNormalInteraction)
		cliSt, err := cli.CancelClient(ctx)
		test.Assert(t, err == nil, err)
		go func() {
			time.Sleep(200 * time.Millisecond)
			cancel()
		}()
		for {
			err = cliSt.Send(cliSt.Context(), &tenant.EchoRequest{Msg: clientStreamingDuringNormalInteraction})
			if err == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			verifyClientSideCancelCase(t, cliSt.Context(), err)
			break
		}
	})
	t.Run(clientStreamingDeferCancel, func(t *testing.T) {
		defer func() {
			<-cancelFinishChan
		}()
		cancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, clientStreamingDeferCancel)
		cliSt, err := cli.CancelClient(ctx)
		test.Assert(t, err == nil, err)
		defer cancel()
		for i := 0; i < 10; i++ {
			err = cliSt.Send(cliSt.Context(), &tenant.EchoRequest{Msg: clientStreamingDeferCancel})
			test.Assert(t, err == nil, err)
		}
	})
	t.Run(serverStreamingRemoteRespondingSlowly, func(t *testing.T) {
		defer func() {
			<-cancelFinishChan
		}()
		cancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, serverStreamingRemoteRespondingSlowly)
		srvSt, err := cli.CancelServer(ctx, &tenant.EchoRequest{Msg: serverStreamingRemoteRespondingSlowly})
		test.Assert(t, err == nil, err)
		time.Sleep(50 * time.Millisecond)
		cancel()
		cancelSendChan <- struct{}{}
		_, err = srvSt.Recv(srvSt.Context())
		verifyClientSideCancelCase(t, srvSt.Context(), err)
	})
	t.Run(serverStreamingDuringNormalInteraction, func(t *testing.T) {
		defer func() {
			<-cancelFinishChan
		}()
		cancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, serverStreamingDuringNormalInteraction)
		srvSt, err := cli.CancelServer(ctx, &tenant.EchoRequest{Msg: serverStreamingDuringNormalInteraction})
		test.Assert(t, err == nil, err)
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		for {
			resp, err := srvSt.Recv(srvSt.Context())
			if err == nil {
				test.Assert(t, resp.Msg == serverStreamingDuringNormalInteraction, resp)
				continue
			}
			verifyClientSideCancelCase(t, srvSt.Context(), err)
			break
		}
	})
	t.Run(serverStreamingDeferCancel, func(t *testing.T) {
		defer func() {
			<-cancelFinishChan
		}()
		cancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, serverStreamingDeferCancel)
		srvSt, err := cli.CancelServer(ctx, &tenant.EchoRequest{Msg: serverStreamingDeferCancel})
		test.Assert(t, err == nil, err)
		defer cancel()
		for {
			resp, err := srvSt.Recv(srvSt.Context())
			test.Assert(t, err == nil, err)
			// special exit sign that has been agreed with downstream in the business
			if resp.Msg == bizSpecialMessage {
				break
			}
			test.Assert(t, resp.Msg == serverStreamingDeferCancel)
		}
	})
	t.Run(bidiStreamingIndependentSendRecv, func(t *testing.T) {
		defer func() {
			<-cancelFinishChan
		}()
		cancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, bidiStreamingIndependentSendRecv)
		bidiSt, err := cli.CancelBidi(ctx)
		test.Assert(t, err == nil, err)
		go func() {
			time.Sleep(200 * time.Millisecond)
			cancel()
		}()
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for {
				sErr := bidiSt.Send(bidiSt.Context(), &tenant.EchoRequest{Msg: bidiStreamingIndependentSendRecv})
				if sErr == nil {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				verifyClientSideCancelCase(t, bidiSt.Context(), sErr)
				break
			}
		}()
		go func() {
			defer wg.Done()
			for {
				resp, rErr := bidiSt.Recv(bidiSt.Context())
				if rErr == nil {
					test.Assert(t, resp.Msg == bidiStreamingIndependentSendRecv, resp)
					continue
				}
				verifyClientSideCancelCase(t, bidiSt.Context(), rErr)
				break
			}
		}()
		wg.Wait()
	})
}
