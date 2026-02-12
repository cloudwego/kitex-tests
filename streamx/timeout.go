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
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt/streamcall"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/codes"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/status"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi/testpbtimeoutservice"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/testtimeoutservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
)

var (
	TimeoutFinishChan = make(chan struct{}, 1)
	TimeoutTChan      = make(chan *testing.T, 1)
)

const (
	clientStreamingRecvTimeoutOnly                                 = "client_streaming_recv_timeout"
	clientStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect   = "client_streaming_recv_and_stream_timeout_with_recv_timeout_take_effect"
	clientStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect = "client_streaming_recv_and_stream_timeout_with_stream_timeout_take_effect"
	serverStreamingRecvTimeoutOnly                                 = "server_streaming_recv_timeout"
	serverStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect   = "server_streaming_recv_and_stream_timeout_with_recv_timeout_take_effect"
	serverStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect = "server_streaming_recv_and_stream_timeout_with_stream_timeout_take_effect"
	bidiStreamingRecvTimeoutOnly                                   = "bidi_streaming_recv_timeout"
	bidiStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect     = "bidi_streaming_recv_and_stream_timeout_with_recv_timeout_take_effect"
	bidiStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect   = "bidi_streaming_recv_and_stream_timeout_with_stream_timeout_take_effect"

	timeoutTakeEffect    = 10 * time.Millisecond
	timeoutNotTakeEffect = 400 * time.Millisecond
)

func RunTestTimeoutServer(listenAddr string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", listenAddr)
	evtHdlTracer, testChan := newServerEventHandlerTracer(TimeoutFinishChan)
	hdl := newServerMockEventHandler(testChan)
	svr := testtimeoutservice.NewServer(&timeoutThriftImpl{evtHdl: hdl},
		server.WithServiceAddr(addr),
		server.WithExitWaitTime(100*time.Millisecond),
		server.WithTracer(NewTracer()),
		server.WithTracer(evtHdlTracer),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler),
		server.WithStreamOptions(server.WithStreamEventHandler(
			rpcinfo.ServerStreamEventHandler{
				HandleStreamStartEvent:  hdl.HandleStreamStartEvent,
				HandleStreamRecvEvent:   hdl.HandleStreamRecvEvent,
				HandleStreamSendEvent:   hdl.HandleStreamSendEvent,
				HandleStreamFinishEvent: hdl.HandleStreamFinishEvent,
			},
		)),
	)
	err := testpbtimeoutservice.RegisterService(svr, &timeoutPbImpl{evtHdl: hdl})
	if err != nil {
		panic(err)
	}
	go func() {
		if err := svr.Run(); err != nil {
			println(err)
		}
	}()
	return svr
}

type timeoutThriftImpl struct {
	evtHdl *ServerMockEventHandler
}

func (t timeoutThriftImpl) RecvTimeoutBidi(ctx context.Context, stream tenant.TestTimeoutService_RecvTimeoutBidiServer) (err error) {
	return commonRecvTimeoutBidiImpl[tenant.EchoRequest, tenant.EchoResponse, tenant.TestTimeoutService_RecvTimeoutBidiServer](
		ctx, stream,
		func(s string) *tenant.EchoResponse {
			return &tenant.EchoResponse{Msg: s}
		},
		t.evtHdl,
	)
}

func (t timeoutThriftImpl) RecvTimeoutClient(ctx context.Context, stream tenant.TestTimeoutService_RecvTimeoutClientServer) (err error) {
	return commonRecvTimeoutClientImpl[tenant.EchoRequest, tenant.EchoResponse, tenant.TestTimeoutService_RecvTimeoutClientServer](
		ctx, stream,
		func(s string) *tenant.EchoResponse {
			return &tenant.EchoResponse{Msg: s}
		},
		t.evtHdl,
	)
}

func (t timeoutThriftImpl) RecvTimeoutServer(ctx context.Context, req *tenant.EchoRequest, stream tenant.TestTimeoutService_RecvTimeoutServerServer) (err error) {
	return commonRecvTimeoutServerImpl[tenant.EchoResponse, tenant.TestTimeoutService_RecvTimeoutServerServer](
		ctx, stream,
		func(s string) *tenant.EchoResponse {
			return &tenant.EchoResponse{Msg: s}
		},
		t.evtHdl,
	)
}

type timeoutPbImpl struct {
	evtHdl *ServerMockEventHandler
}

func (t timeoutPbImpl) RecvTimeoutPbBidi(ctx context.Context, stream pbapi.TestPbTimeoutService_RecvTimeoutPbBidiServer) (err error) {
	return commonRecvTimeoutBidiImpl[pbapi.MockReq, pbapi.MockResp, pbapi.TestPbTimeoutService_RecvTimeoutPbBidiServer](
		ctx, stream,
		func(s string) *pbapi.MockResp {
			return &pbapi.MockResp{Message: s}
		},
		t.evtHdl,
	)
}

func (t timeoutPbImpl) RecvTimeoutPbClient(ctx context.Context, stream pbapi.TestPbTimeoutService_RecvTimeoutPbClientServer) (err error) {
	return commonRecvTimeoutClientImpl[pbapi.MockReq, pbapi.MockResp, pbapi.TestPbTimeoutService_RecvTimeoutPbClientServer](
		ctx, stream,
		func(s string) *pbapi.MockResp {
			return &pbapi.MockResp{Message: s}
		},
		t.evtHdl,
	)
}

func (t timeoutPbImpl) RecvTimeoutPbServer(ctx context.Context, req *pbapi.MockReq, stream pbapi.TestPbTimeoutService_RecvTimeoutPbServerServer) (err error) {
	return commonRecvTimeoutServerImpl[pbapi.MockResp, pbapi.TestPbTimeoutService_RecvTimeoutPbServerServer](
		ctx, stream,
		func(s string) *pbapi.MockResp {
			return &pbapi.MockResp{Message: s}
		},
		t.evtHdl,
	)
}

func commonRecvTimeoutBidiImpl[Req, Res any, Stream streaming.BidiStreamingServer[Req, Res]](
	ctx context.Context, stream Stream,
	resGetter func(string) *Res,
	evtHdl *ServerMockEventHandler,
) error {
	t := <-TimeoutTChan
	scenario, ok := getScenario(ctx)
	test.Assert(t, ok)
	isGRPC := isGRPCFunc(t, ctx)

	switch scenario {
	case bidiStreamingRecvTimeoutOnly, bidiStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect:
		var wg sync.WaitGroup
		wg.Add(2)
		finishCh := make(chan struct{})
		var recvTimes, sendTimes uint32
		go func() {
			defer wg.Done()
			for {
				_, rErr := stream.Recv(ctx)
				if rErr == nil {
					recvTimes++
					continue
				}
				test.Assert(t, rErr == io.EOF, rErr)
				close(finishCh)
				break
			}
		}()
		go func() {
			defer wg.Done()
			<-finishCh
			randI := fastrand.Intn(10)
			var sErr error
			for i := 0; ; i++ {
				sendTimes++
				if sErr = stream.Send(ctx, resGetter(scenario)); sErr != nil {
					break
				}
				if i == randI {
					<-ctx.Done()
				}
			}
			verifyServerSideCancelCase(t, ctx, sErr, isGRPC)
		}()
		wg.Wait()
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   recvTimes,
			sendTimes:   sendTimes,
			finishTimes: 1,
		})
	case bidiStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect:
		var wg sync.WaitGroup
		wg.Add(2)
		finishCh := make(chan struct{})
		var recvTimes, sendTimes uint32
		go func() {
			defer wg.Done()
			for {
				_, rErr := stream.Recv(ctx)
				if rErr == nil {
					recvTimes++
					continue
				}
				test.Assert(t, rErr == io.EOF, rErr)
				close(finishCh)
				break
			}
		}()
		go func() {
			defer wg.Done()
			<-finishCh
			randI := fastrand.Intn(10)
			for i := 0; ; i++ {
				sendTimes++
				sErr := stream.Send(ctx, resGetter(scenario))
				if sErr == nil {
					if i != randI {
						continue
					}
					<-ctx.Done()
					continue
				}
				verifyServerSideTimeoutCase(t, ctx, sErr, isGRPC)
				break
			}
		}()
		wg.Wait()
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   recvTimes,
			sendTimes:   sendTimes,
			finishTimes: 1,
		})
	}

	return nil
}

func commonRecvTimeoutClientImpl[Req, Res any, Stream streaming.ClientStreamingServer[Req, Res]](
	ctx context.Context, stream Stream,
	resGetter func(string) *Res,
	evtHdl *ServerMockEventHandler,
) error {
	t := <-TimeoutTChan
	isGRPC := isGRPCFunc(t, ctx)

	scenario, ok := getScenario(ctx)
	test.Assert(t, ok)
	switch scenario {
	case clientStreamingRecvTimeoutOnly, clientStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect:
		var recvTimes uint32
		for {
			_, rErr := stream.Recv(ctx)
			if rErr == nil {
				recvTimes++
				continue
			}
			test.Assert(t, rErr == io.EOF, rErr)
			break
		}
		<-ctx.Done()
		sErr := stream.SendMsg(ctx, resGetter(scenario))
		test.Assert(t, sErr != nil)
		verifyServerSideCancelCase(t, ctx, sErr, isGRPC)
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   recvTimes,
			sendTimes:   1,
			finishTimes: 1,
		})
	case clientStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect:
		var recvTimes uint32
		for {
			_, rErr := stream.Recv(ctx)
			if rErr == nil {
				recvTimes++
				continue
			}
			test.Assert(t, rErr == io.EOF, rErr)
			break
		}
		<-ctx.Done()
		sErr := stream.SendMsg(ctx, resGetter(scenario))
		// gRPC cannot detect stream timeouts through Send operations, so there is a chance that Send will not report an error.
		if !isGRPC {
			test.Assert(t, sErr != nil)
		}
		if sErr != nil {
			verifyServerSideTimeoutCase(t, ctx, sErr, isGRPC)
		}
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   recvTimes,
			sendTimes:   1,
			finishTimes: 1,
		})
	}
	return nil
}

func commonRecvTimeoutServerImpl[Res any, Stream streaming.ServerStreamingServer[Res]](
	ctx context.Context, stream Stream,
	resGetter func(string) *Res,
	evtHdl *ServerMockEventHandler,
) error {
	t := <-TimeoutTChan
	isGRPC := isGRPCFunc(t, ctx)

	scenario, ok := getScenario(ctx)
	test.Assert(t, ok)
	switch scenario {
	case serverStreamingRecvTimeoutOnly, serverStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect:
		n := fastrand.Intn(10)
		var sendTimes uint32
		for i := 0; ; i++ {
			res := resGetter(scenario)
			sendTimes++
			sErr := stream.Send(ctx, res)
			if sErr == nil {
				if i == n {
					<-ctx.Done()
				}
				continue
			}
			t.Logf("scenario[%s] Send err: %v", scenario, sErr)
			verifyServerSideCancelCase(t, ctx, sErr, isGRPC)
			break
		}
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   1,
			sendTimes:   sendTimes,
			finishTimes: 1,
		})
	case serverStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect:
		n := fastrand.Intn(10)
		var sendTimes uint32
		for i := 0; ; i++ {
			res := resGetter(scenario)
			sendTimes++
			sErr := stream.Send(ctx, res)
			if sErr == nil {
				if i == n {
					<-ctx.Done()
				}
				continue
			}
			t.Logf("scenario[%s] Send err: %v", scenario, sErr)
			verifyServerSideTimeoutCase(t, ctx, sErr, isGRPC)
			break
		}
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   1,
			sendTimes:   sendTimes,
			finishTimes: 1,
		})
	}
	return nil
}

type timeoutClient[Req, Res any] interface {
	RecvTimeoutBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[Req, Res], err error)
	RecvTimeoutClient(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.ClientStreamingClient[Req, Res], err error)
	RecvTimeoutServer(ctx context.Context, Req *Req, callOptions ...streamcall.Option) (stream streaming.ServerStreamingClient[Res], err error)
}

type thriftTimeoutClient struct {
	cli testtimeoutservice.Client
}

func (t thriftTimeoutClient) RecvTimeoutBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[tenant.EchoRequest, tenant.EchoResponse], err error) {
	return t.cli.RecvTimeoutBidi(ctx, callOptions...)
}

func (t thriftTimeoutClient) RecvTimeoutClient(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.ClientStreamingClient[tenant.EchoRequest, tenant.EchoResponse], err error) {
	return t.cli.RecvTimeoutClient(ctx, callOptions...)
}

func (t thriftTimeoutClient) RecvTimeoutServer(ctx context.Context, req *tenant.EchoRequest, callOptions ...streamcall.Option) (stream streaming.ServerStreamingClient[tenant.EchoResponse], err error) {
	return t.cli.RecvTimeoutServer(ctx, req, callOptions...)
}

type pbTimeoutClient struct {
	cli testpbtimeoutservice.Client
}

func (p pbTimeoutClient) RecvTimeoutBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[pbapi.MockReq, pbapi.MockResp], err error) {
	return p.cli.RecvTimeoutPbBidi(ctx, callOptions...)
}

func (p pbTimeoutClient) RecvTimeoutClient(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.ClientStreamingClient[pbapi.MockReq, pbapi.MockResp], err error) {
	return p.cli.RecvTimeoutPbClient(ctx, callOptions...)
}

func (p pbTimeoutClient) RecvTimeoutServer(ctx context.Context, req *pbapi.MockReq, callOptions ...streamcall.Option) (stream streaming.ServerStreamingClient[pbapi.MockResp], err error) {
	return p.cli.RecvTimeoutPbServer(ctx, req, callOptions...)
}

func TestThriftTimeout(t *testing.T, addr string) {
	t.Run("TTHeader Streaming", func(t *testing.T) {
		t.Run("client timeout option", func(t *testing.T) {
			runTimeout(t, addr, false, true, true)
		})
		t.Run("call timeout option", func(t *testing.T) {
			runTimeout(t, addr, false, false, true)
		})
	})
	t.Run("GRPC", func(t *testing.T) {
		t.Run("client timeout option", func(t *testing.T) {
			runTimeout(t, addr, true, true, true)
		})
		t.Run("call timeout option", func(t *testing.T) {
			runTimeout(t, addr, true, false, true)
		})
	})
}

func TestPbTimeout(t *testing.T, addr string) {
	// TODO: add test back when ttstream with pb payload is merged
	//t.Run("TTHeader Streaming", func(t *testing.T) {
	//	t.Run("client timeout option", func(t *testing.T) {
	//		runTimeout(t, addr, false, true, true)
	//	})
	//	t.Run("call timeout option", func(t *testing.T) {
	//		runTimeout(t, addr, false, false, true)
	//	})
	//})
	t.Run("GRPC", func(t *testing.T) {
		t.Run("client timeout option", func(t *testing.T) {
			runTimeout(t, addr, true, true, false)
		})
		t.Run("call timeout option", func(t *testing.T) {
			runTimeout(t, addr, true, false, false)
		})
	})
}

func runTimeout(t *testing.T, addr string, isGRPC, isClientTimeout, isThrift bool) {
	cliSvcName := "client-side timeout service"
	evtHdl := newClientMockEventHandler()
	opts := []client.Option{
		client.WithHostPorts(addr),
		client.WithMetaHandler(transmeta.ClientHTTP2Handler),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: cliSvcName,
		}),
		client.WithStreamOptions(
			client.WithStreamEventHandler(rpcinfo.ClientStreamEventHandler{
				HandleStreamStartEvent:      evtHdl.HandleStreamStartEvent,
				HandleStreamRecvHeaderEvent: evtHdl.HandleStreamRecvHeaderEvent,
				HandleStreamRecvEvent:       evtHdl.HandleStreamRecvEvent,
				HandleStreamSendEvent:       evtHdl.HandleStreamSendEvent,
				HandleStreamFinishEvent:     evtHdl.HandleStreamFinishEvent,
			}),
		),
	}
	if isClientTimeout {
		opts = append(opts, client.WithStreamOptions(client.WithStreamRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: 200 * time.Millisecond})))
	}
	if isGRPC {
		opts = append(opts, client.WithTransportProtocol(transport.GRPCStreaming))
	}
	if isThrift {
		reqGetter := func(s string) *tenant.EchoRequest {
			return &tenant.EchoRequest{Msg: s}
		}
		resMsgSetter := func(p *tenant.EchoResponse) string {
			return p.Msg
		}
		cli := testtimeoutservice.MustNewClient("service", opts...)
		commonTestTimeout[tenant.EchoRequest, tenant.EchoResponse](t, thriftTimeoutClient{cli}, isGRPC, isClientTimeout, reqGetter, resMsgSetter, evtHdl)
	} else {
		reqGetter := func(s string) *pbapi.MockReq {
			return &pbapi.MockReq{Message: s}
		}
		resMsgSetter := func(p *pbapi.MockResp) string {
			return p.Message
		}
		cli := testpbtimeoutservice.MustNewClient("service", opts...)
		commonTestTimeout[pbapi.MockReq, pbapi.MockResp](t, pbTimeoutClient{cli}, isGRPC, isClientTimeout, reqGetter, resMsgSetter, evtHdl)
	}
}

func commonTestTimeout[Req, Res any](t *testing.T, cli timeoutClient[Req, Res], isGRPC, isClientTimeout bool,
	reqGetter func(string) *Req, resMsgGetter func(*Res) string,
	evtHdl *ClientMockEventHandler,
) {
	t.Run(clientStreamingRecvTimeoutOnly, func(t *testing.T) {
		defer func() {
			<-TimeoutFinishChan
			evtHdl.Reset()
		}()
		TimeoutTChan <- t
		ctx := setScenario(context.Background(), clientStreamingRecvTimeoutOnly)
		var callOpts []streamcall.Option
		if !isClientTimeout {
			callOpts = append(callOpts, streamcall.WithRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: timeoutTakeEffect}))
		}
		cliSt, err := cli.RecvTimeoutClient(ctx, callOpts...)
		test.Assert(t, err == nil, err)
		var sendTimes uint32
		for i := 0; i < 3; i++ {
			sendTimes++
			err = cliSt.Send(cliSt.Context(), reqGetter(clientStreamingRecvTimeoutOnly))
			test.Assert(t, err == nil, err)
		}
		_, err = cliSt.CloseAndRecv(cliSt.Context())
		test.Assert(t, err != nil)
		t.Log(err)
		verifyRecvTimeout(t, err, isGRPC, false)
		evtHdl.AssertCalledTimes(t, ClientCalledTimes{
			startTimes:      1,
			recvHeaderTimes: 0,
			recvTimes:       1,
			sendTimes:       sendTimes,
			finishTimes:     1,
		})
	})
	t.Run(serverStreamingRecvTimeoutOnly, func(t *testing.T) {
		defer func() {
			<-TimeoutFinishChan
			evtHdl.Reset()
		}()
		TimeoutTChan <- t
		ctx := setScenario(context.Background(), serverStreamingRecvTimeoutOnly)
		var callOpts []streamcall.Option
		if !isClientTimeout {
			callOpts = append(callOpts, streamcall.WithRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: timeoutTakeEffect}))
		}
		srvSt, err := cli.RecvTimeoutServer(ctx, reqGetter(serverStreamingRecvTimeoutOnly), callOpts...)
		test.Assert(t, err == nil, err)
		var recvTimes uint32
		for {
			recvTimes++
			_, err = srvSt.Recv(srvSt.Context())
			if err == nil {
				continue
			}
			t.Log(err)
			verifyRecvTimeout(t, err, isGRPC, false)
			break
		}
		evtHdl.AssertCalledTimes(t, ClientCalledTimes{
			startTimes:      1,
			recvHeaderTimes: 1,
			recvTimes:       recvTimes,
			sendTimes:       1,
			finishTimes:     1,
		})
	})
	t.Run(bidiStreamingRecvTimeoutOnly, func(t *testing.T) {
		defer func() {
			<-TimeoutFinishChan
			evtHdl.Reset()
		}()
		TimeoutTChan <- t
		ctx := setScenario(context.Background(), bidiStreamingRecvTimeoutOnly)
		var callOpts []streamcall.Option
		if !isClientTimeout {
			callOpts = append(callOpts, streamcall.WithRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: timeoutTakeEffect}))
		}
		bidiSt, err := cli.RecvTimeoutBidi(ctx, callOpts...)
		test.Assert(t, err == nil, err)
		var sendTimes, recvTimes uint32
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				sendTimes++
				sErr := bidiSt.Send(bidiSt.Context(), reqGetter(bidiStreamingRecvTimeoutOnly))
				test.Assert(t, sErr == nil, sErr)
			}
			sErr := bidiSt.CloseSend(bidiSt.Context())
			test.Assert(t, sErr == nil, sErr)
		}()
		go func() {
			defer wg.Done()
			for {
				recvTimes++
				_, rErr := bidiSt.Recv(bidiSt.Context())
				if rErr == nil {
					continue
				}
				verifyRecvTimeout(t, rErr, isGRPC, false)
				t.Log(rErr)
				break
			}
		}()
		wg.Wait()
		evtHdl.AssertCalledTimes(t, ClientCalledTimes{
			startTimes:      1,
			recvHeaderTimes: 1,
			recvTimes:       recvTimes,
			sendTimes:       sendTimes,
			finishTimes:     1,
		})
	})
	// for now ttstream only supports stream Recv timeout
	if isGRPC {
		t.Run(clientStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect, func(t *testing.T) {
			defer func() {
				<-TimeoutFinishChan
				evtHdl.Reset()
			}()
			TimeoutTChan <- t
			ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
			defer cancel()
			ctx = setScenario(ctx, clientStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect)
			var callOpts []streamcall.Option
			if !isClientTimeout {
				callOpts = append(callOpts, streamcall.WithRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: timeoutTakeEffect}))
			}
			cliSt, err := cli.RecvTimeoutClient(ctx, callOpts...)
			test.Assert(t, err == nil, err)
			var sendTimes uint32
			for i := 0; i < 3; i++ {
				sendTimes++
				err = cliSt.Send(cliSt.Context(), reqGetter(clientStreamingRecvTimeoutOnly))
				test.Assert(t, err == nil, err)
			}
			_, err = cliSt.CloseAndRecv(cliSt.Context())
			test.Assert(t, err != nil)
			t.Log(err)
			verifyRecvTimeout(t, err, isGRPC, false)
			evtHdl.AssertCalledTimes(t, ClientCalledTimes{
				startTimes:      1,
				recvHeaderTimes: 0,
				recvTimes:       1,
				sendTimes:       sendTimes,
				finishTimes:     1,
			})
		})
		t.Run(clientStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect, func(t *testing.T) {
			defer func() {
				<-TimeoutFinishChan
				evtHdl.Reset()
			}()
			TimeoutTChan <- t
			ctx, cancel := context.WithTimeout(context.Background(), timeoutTakeEffect)
			defer cancel()
			ctx = setScenario(ctx, clientStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect)
			var callOpts []streamcall.Option
			if !isClientTimeout {
				callOpts = append(callOpts, streamcall.WithRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: timeoutNotTakeEffect}))
			}
			cliSt, err := cli.RecvTimeoutClient(ctx, callOpts...)
			test.Assert(t, err == nil, err)
			var sendTimes uint32
			for i := 0; i < 3; i++ {
				sendTimes++
				err = cliSt.Send(cliSt.Context(), reqGetter(clientStreamingRecvTimeoutOnly))
				test.Assert(t, err == nil, err)
			}
			_, err = cliSt.CloseAndRecv(cliSt.Context())
			test.Assert(t, err != nil)
			t.Log(err)
			verifyRecvTimeout(t, err, isGRPC, true)
			evtHdl.AssertCalledTimes(t, ClientCalledTimes{
				startTimes:      1,
				recvHeaderTimes: 0,
				recvTimes:       1,
				sendTimes:       sendTimes,
				finishTimes:     1,
			})
		})
		t.Run(serverStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect, func(t *testing.T) {
			defer func() {
				<-TimeoutFinishChan
				evtHdl.Reset()
			}()
			TimeoutTChan <- t
			ctx, cancel := context.WithTimeout(context.Background(), timeoutNotTakeEffect)
			defer cancel()
			ctx = setScenario(ctx, serverStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect)
			var callOpts []streamcall.Option
			if !isClientTimeout {
				callOpts = append(callOpts, streamcall.WithRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: timeoutTakeEffect}))
			}
			srvSt, err := cli.RecvTimeoutServer(ctx, reqGetter(serverStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect), callOpts...)
			test.Assert(t, err == nil, err)
			var recvTimes uint32
			for {
				recvTimes++
				_, err = srvSt.Recv(srvSt.Context())
				if err == nil {
					continue
				}
				t.Log(err)
				verifyRecvTimeout(t, err, isGRPC, false)
				break
			}
			evtHdl.AssertCalledTimes(t, ClientCalledTimes{
				startTimes:      1,
				recvHeaderTimes: 1,
				recvTimes:       recvTimes,
				sendTimes:       1,
				finishTimes:     1,
			})
		})
		t.Run(serverStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect, func(t *testing.T) {
			defer func() {
				<-TimeoutFinishChan
				evtHdl.Reset()
			}()
			TimeoutTChan <- t
			ctx, cancel := context.WithTimeout(context.Background(), timeoutTakeEffect)
			defer cancel()
			ctx = setScenario(ctx, serverStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect)
			var callOpts []streamcall.Option
			if !isClientTimeout {
				callOpts = append(callOpts, streamcall.WithRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: timeoutNotTakeEffect}))
			}
			srvSt, err := cli.RecvTimeoutServer(ctx, reqGetter(serverStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect), callOpts...)
			test.Assert(t, err == nil, err)
			var recvTimes uint32
			for {
				recvTimes++
				_, err = srvSt.Recv(srvSt.Context())
				if err == nil {
					continue
				}
				t.Log(err)
				verifyRecvTimeout(t, err, isGRPC, true)
				break
			}
			evtHdl.AssertCalledTimes(t, ClientCalledTimes{
				startTimes:      1,
				recvHeaderTimes: 1,
				recvTimes:       recvTimes,
				sendTimes:       1,
				finishTimes:     1,
			})
		})
		t.Run(bidiStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect, func(t *testing.T) {
			defer func() {
				<-TimeoutFinishChan
				evtHdl.Reset()
			}()
			TimeoutTChan <- t
			ctx, cancel := context.WithTimeout(context.Background(), timeoutNotTakeEffect)
			defer cancel()
			ctx = setScenario(ctx, bidiStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect)
			var callOpts []streamcall.Option
			if !isClientTimeout {
				callOpts = append(callOpts, streamcall.WithRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: timeoutTakeEffect}))
			}
			bidiSt, err := cli.RecvTimeoutBidi(ctx, callOpts...)
			test.Assert(t, err == nil, err)
			var sendTimes, recvTimes uint32
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				for i := 0; i < 3; i++ {
					sendTimes++
					sErr := bidiSt.Send(bidiSt.Context(), reqGetter(bidiStreamingRecvTimeoutOnly))
					test.Assert(t, sErr == nil, sErr)
				}
				sErr := bidiSt.CloseSend(bidiSt.Context())
				test.Assert(t, sErr == nil, sErr)
			}()
			go func() {
				defer wg.Done()
				for {
					recvTimes++
					_, rErr := bidiSt.Recv(bidiSt.Context())
					if rErr == nil {
						continue
					}
					verifyRecvTimeout(t, rErr, isGRPC, false)
					t.Log(rErr)
					break
				}
			}()
			wg.Wait()
			evtHdl.AssertCalledTimes(t, ClientCalledTimes{
				startTimes:      1,
				recvHeaderTimes: 1,
				recvTimes:       recvTimes,
				sendTimes:       sendTimes,
				finishTimes:     1,
			})
		})
		t.Run(bidiStreamingRecvAndStreamTimeoutWithStreamTimeoutTakeEffect, func(t *testing.T) {
			defer func() {
				<-TimeoutFinishChan
				evtHdl.Reset()
			}()
			TimeoutTChan <- t
			ctx, cancel := context.WithTimeout(context.Background(), timeoutTakeEffect)
			defer cancel()
			ctx = setScenario(ctx, bidiStreamingRecvAndStreamTimeoutWithRecvTimeoutTakeEffect)
			var callOpts []streamcall.Option
			if !isClientTimeout {
				callOpts = append(callOpts, streamcall.WithRecvTimeoutConfig(streaming.TimeoutConfig{Timeout: timeoutNotTakeEffect}))
			}
			bidiSt, err := cli.RecvTimeoutBidi(ctx, callOpts...)
			test.Assert(t, err == nil, err)
			var sendTimes, recvTimes uint32
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				for i := 0; i < 3; i++ {
					sendTimes++
					sErr := bidiSt.Send(bidiSt.Context(), reqGetter(bidiStreamingRecvTimeoutOnly))
					test.Assert(t, sErr == nil, sErr)
				}
				sErr := bidiSt.CloseSend(bidiSt.Context())
				test.Assert(t, sErr == nil, sErr)
			}()
			go func() {
				defer wg.Done()
				for {
					recvTimes++
					_, rErr := bidiSt.Recv(bidiSt.Context())
					if rErr == nil {
						continue
					}
					verifyRecvTimeout(t, rErr, isGRPC, true)
					t.Log(rErr)
					break
				}
			}()
			wg.Wait()
			evtHdl.AssertCalledTimes(t, ClientCalledTimes{
				startTimes:      1,
				recvHeaderTimes: 1,
				recvTimes:       recvTimes,
				sendTimes:       sendTimes,
				finishTimes:     1,
			})
		})
	}
}

func verifyRecvTimeout(t *testing.T, err error, isGRPC, isStreamTimeout bool) {
	if isGRPC {
		st, ok := status.FromError(err)
		test.Assert(t, ok, err)
		if !isStreamTimeout {
			test.Assert(t, st.Code() == codes.RecvDeadlineExceeded, st.Code())
		} else {
			test.Assert(t, st.Code() == codes.DeadlineExceeded, st.Code())
		}
	} else {
		test.Assert(t, errors.Is(err, kerrors.ErrStreamingTimeout), err)
	}
}

func verifyTimeoutContext(t *testing.T, ctx context.Context) {
	var ctxDone bool
	select {
	case <-ctx.Done():
		ctxDone = true
	default:
	}
	test.Assert(t, ctxDone)
	_, hasDdl := ctx.Deadline()
	test.Assert(t, hasDdl)
}

func verifyServerSideTimeoutErr(t *testing.T, err error, isGRPC bool) {
	if isGRPC {
		st, ok := status.FromError(err)
		test.Assert(t, ok)
		test.Assert(t, st.Code() == codes.DeadlineExceeded || st.Code() == codes.Canceled, st.Code())
	}
}

func verifyServerSideTimeoutCase(t *testing.T, ctx context.Context, err error, isGRPC bool) {
	verifyTimeoutContext(t, ctx)
	verifyServerSideTimeoutErr(t, err, isGRPC)
}
