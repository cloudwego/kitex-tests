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
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
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
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi/testpbcancelservice"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/testcancelservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
)

var (
	CancelFinishChan = make(chan struct{}, 1)
	CancelTChan      = make(chan *testing.T, 1)
	CancelSendChan   = make(chan struct{}, 1)
)

const (
	clientStreamingWithoutSendingAnyReq    string = "client_streaming_without_sending_any_req"
	clientStreamingDuringNormalInteraction string = "client_streaming_during_normal_interaction"
	clientStreamingDeferCancel             string = "client_streaming_defer_cancel"

	serverStreamingRemoteRespondingSlowly  string = "server_streaming_remote_responding_slowly"
	serverStreamingDuringNormalInteraction string = "server_streaming_during_normal_interaction"
	serverStreamingDeferCancel             string = "server_streaming_defer_cancel"

	bidiStreamingIndependentSendRecv string = "bidi_streaming_independent_send_recv"

	scenarioKey       = "SCENARIO"
	bizSpecialMessage = "biz_special_message"

	cancelInterval = 20 * time.Millisecond
)

func RunTestCancelServer(listenAddr string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", listenAddr)
	evtHdlTracer, testChan := newServerEventHandlerTracer(CancelFinishChan)
	hdl := newServerMockEventHandler(testChan)
	svr := testcancelservice.NewServer(&cancelThriftImpl{evtHdl: hdl},
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
	err := testpbcancelservice.RegisterService(svr, &cancelPbImpl{evtHdl: hdl})
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

type cancelThriftImpl struct {
	evtHdl *ServerMockEventHandler
}

func (s cancelThriftImpl) CancelBidi(ctx context.Context, stream tenant.TestCancelService_CancelBidiServer) (err error) {
	return commonCancelBidiImpl[tenant.EchoRequest, tenant.EchoResponse, tenant.TestCancelService_CancelBidiServer](
		ctx, stream,
		func(s string) *tenant.EchoResponse {
			return &tenant.EchoResponse{Msg: s}
		},
		s.evtHdl,
	)
}

func (s cancelThriftImpl) CancelClient(ctx context.Context, stream tenant.TestCancelService_CancelClientServer) (err error) {
	return commonCancelClientImpl[tenant.EchoRequest, tenant.EchoResponse, tenant.TestCancelService_CancelClientServer](ctx, stream, s.evtHdl)
}

func (s cancelThriftImpl) CancelServer(ctx context.Context, req *tenant.EchoRequest, stream tenant.TestCancelService_CancelServerServer) (err error) {
	return commonCancelServerImpl[tenant.EchoResponse, tenant.TestCancelService_CancelServerServer](
		ctx, stream,
		func(s string) *tenant.EchoResponse {
			return &tenant.EchoResponse{Msg: s}
		},
		s.evtHdl,
	)
}

type cancelPbImpl struct {
	evtHdl *ServerMockEventHandler
}

func (s cancelPbImpl) CancelPbBidi(ctx context.Context, stream pbapi.TestPbCancelService_CancelPbBidiServer) (err error) {
	return commonCancelBidiImpl[pbapi.MockReq, pbapi.MockResp, pbapi.TestPbCancelService_CancelPbBidiServer](
		ctx, stream,
		func(s string) *pbapi.MockResp {
			return &pbapi.MockResp{Message: s}
		},
		s.evtHdl,
	)
}

func (s cancelPbImpl) CancelPbClient(ctx context.Context, stream pbapi.TestPbCancelService_CancelPbClientServer) (err error) {
	return commonCancelClientImpl[pbapi.MockReq, pbapi.MockResp, pbapi.TestPbCancelService_CancelPbClientServer](ctx, stream, s.evtHdl)
}

func (s cancelPbImpl) CancelPbServer(ctx context.Context, req *pbapi.MockReq, stream pbapi.TestPbCancelService_CancelPbServerServer) (err error) {
	return commonCancelServerImpl[pbapi.MockResp, pbapi.TestPbCancelService_CancelPbServerServer](
		ctx, stream,
		func(s string) *pbapi.MockResp {
			return &pbapi.MockResp{Message: s}
		},
		s.evtHdl,
	)
}

func isGRPCFunc(t *testing.T, ctx context.Context) bool {
	ri := rpcinfo.GetRPCInfo(ctx)
	test.Assert(t, ri != nil, ctx)
	switch ri.Config().TransportProtocol() {
	case transport.GRPC:
		return true
	case transport.TTHeaderStreaming:
		return false
	default:
		t.Fatal(fmt.Sprintf("not expected transport protocol: %v", ri.Config().TransportProtocol()))
		return false
	}
}

func commonCancelBidiImpl[Req, Res any, Stream streaming.BidiStreamingServer[Req, Res]](
	ctx context.Context, stream Stream,
	resGetter func(string) *Res,
	evtHdl *ServerMockEventHandler,
) error {
	t := <-CancelTChan
	scenario, ok := getScenario(ctx)
	test.Assert(t, ok)
	isGRPC := isGRPCFunc(t, ctx)

	switch scenario {
	case bidiStreamingIndependentSendRecv:
		var wg sync.WaitGroup
		wg.Add(2)
		var recvTimes, sendTimes uint32
		go func() {
			defer wg.Done()
			for {
				recvTimes++
				_, rErr := stream.Recv(ctx)
				if rErr == nil {
					continue
				}
				verifyServerSideCancelCase(t, ctx, rErr, isGRPC)
				break
			}
		}()
		go func() {
			defer wg.Done()
			for {
				res := resGetter(scenario)
				sendTimes++
				sErr := stream.Send(ctx, res)
				if sErr == nil {
					continue
				}
				verifyServerSideCancelCase(t, ctx, sErr, isGRPC)
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

func commonCancelClientImpl[Req, Res any, Stream streaming.ClientStreamingServer[Req, Res]](
	ctx context.Context, stream Stream,
	evtHdl *ServerMockEventHandler,
) error {
	t := <-CancelTChan
	isGRPC := isGRPCFunc(t, ctx)

	scenario, ok := getScenario(ctx)
	test.Assert(t, ok)
	switch scenario {
	case clientStreamingWithoutSendingAnyReq:
		// the first Recv result would be err
		_, rErr := stream.Recv(ctx)
		test.Assert(t, rErr != nil)
		t.Logf("scenario[%s] Recv err: %v", scenario, rErr)
		verifyServerSideCancelCase(t, ctx, rErr, isGRPC)
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   1,
			sendTimes:   0,
			finishTimes: 1,
		})
	case clientStreamingDuringNormalInteraction:
		var recvTimes uint32
		for {
			recvTimes++
			_, rErr := stream.Recv(ctx)
			if rErr == nil {
				continue
			}
			t.Logf("scenario[%s] Recv err: %v", scenario, rErr)
			verifyServerSideCancelCase(t, ctx, rErr, isGRPC)
			break
		}
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   recvTimes,
			sendTimes:   0,
			finishTimes: 1,
		})
	case clientStreamingDeferCancel:
		var recvTimes uint32
		for {
			recvTimes++
			_, rErr := stream.Recv(ctx)
			if rErr == nil {
				continue
			}
			t.Logf("scenario[%s] Recv err: %v", scenario, rErr)
			verifyServerSideCancelCase(t, ctx, rErr, isGRPC)
			break
		}
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   recvTimes,
			sendTimes:   0,
			finishTimes: 1,
		})
	}
	return nil
}

func commonCancelServerImpl[Res any, Stream streaming.ServerStreamingServer[Res]](
	ctx context.Context, stream Stream,
	resGetter func(string) *Res,
	evtHdl *ServerMockEventHandler,
) error {
	t := <-CancelTChan
	isGRPC := isGRPCFunc(t, ctx)

	scenario, ok := getScenario(ctx)
	test.Assert(t, ok)
	switch scenario {
	case serverStreamingRemoteRespondingSlowly:
		<-CancelSendChan
		// in case Rst Frame has not been processed due to dispatch or other reasons
		// we start a loop to Send until returning err
		var sendTimes uint32
		for {
			res := resGetter(scenario)
			sendTimes++
			sErr := stream.Send(ctx, res)
			if sErr == nil {
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
	case serverStreamingDuringNormalInteraction:
		var sendTimes uint32
		for {
			res := resGetter(scenario)
			sendTimes++
			sErr := stream.Send(ctx, res)
			if sErr == nil {
				continue
			}
			t.Logf("scenario[%s] Recv err: %v", scenario, sErr)
			verifyServerSideCancelCase(t, ctx, sErr, isGRPC)
			break
		}
		evtHdl.AssertCalledTimes(t, ServerCalledTimes{
			startTimes:  1,
			recvTimes:   1,
			sendTimes:   sendTimes,
			finishTimes: 1,
		})
	case serverStreamingDeferCancel:
		var sendTimes uint32
		for i := 0; i < 5; i++ {
			res := resGetter(scenario)
			sendTimes++
			sErr := stream.Send(ctx, res)
			test.Assert(t, sErr == nil, sErr)
		}
		specialRes := resGetter(bizSpecialMessage)
		sendTimes++
		sErr := stream.Send(ctx, specialRes)
		test.Assert(t, sErr == nil, sErr)
		// waiting for remote service cancel
		for {
			res := resGetter(scenario)
			sendTimes++
			sErr = stream.Send(ctx, res)
			if sErr == nil {
				continue
			}
			t.Logf("scenario[%s] Recv err: %v", scenario, sErr)
			verifyServerSideCancelCase(t, ctx, sErr, isGRPC)
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

func setScenario(ctx context.Context, scenario string) context.Context {
	return metainfo.WithValue(ctx, scenarioKey, scenario)
}

func getScenario(ctx context.Context) (string, bool) {
	return metainfo.GetValue(ctx, scenarioKey)
}

func verifyClientSideCancelCase(t *testing.T, ctx context.Context, err error, isGRPC, isSend bool) {
	verifyCancelContext(t, ctx)
	verifyClientSideCancelErr(t, err, isGRPC, isSend)
}

func verifyServerSideCancelCase(t *testing.T, ctx context.Context, err error, isGRPC bool) {
	verifyCancelContext(t, ctx)
	verifyServerSideCancelErr(t, err, isGRPC)
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

func verifyClientSideCancelErr(t *testing.T, err error, isGRPC, isSend bool) {
	if isGRPC {
		st, ok := status.FromError(err)
		test.Assert(t, ok)
		code := st.Code()
		if isSend {
			test.Assert(t, code == codes.Canceled || code == codes.Internal, code, st.Message())
		} else {
			test.Assert(t, code == codes.Canceled, code, st.Message())
		}
	} else {
		test.Assert(t, errors.Is(err, kerrors.ErrStreamingCanceled), err)
	}
}

func verifyServerSideCancelErr(t *testing.T, err error, isGRPC bool) {
	if isGRPC {
		st, ok := status.FromError(err)
		test.Assert(t, ok)
		test.Assert(t, st.Code() == codes.Canceled, st.Code())
	} else {
		test.Assert(t, errors.Is(err, kerrors.ErrStreamingCanceled), err)
		test.Assert(t, strings.Contains(err.Error(), "canceled path"), err)
	}
}

type cancelClient[Req, Res any] interface {
	CancelBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[Req, Res], err error)
	CancelClient(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.ClientStreamingClient[Req, Res], err error)
	CancelServer(ctx context.Context, Req *Req, callOptions ...streamcall.Option) (stream streaming.ServerStreamingClient[Res], err error)
}

type thriftCancelClient struct {
	cli testcancelservice.Client
}

func (t thriftCancelClient) CancelBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[tenant.EchoRequest, tenant.EchoResponse], err error) {
	return t.cli.CancelBidi(ctx, callOptions...)
}

func (t thriftCancelClient) CancelClient(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.ClientStreamingClient[tenant.EchoRequest, tenant.EchoResponse], err error) {
	return t.cli.CancelClient(ctx, callOptions...)
}

func (t thriftCancelClient) CancelServer(ctx context.Context, req *tenant.EchoRequest, callOptions ...streamcall.Option) (stream streaming.ServerStreamingClient[tenant.EchoResponse], err error) {
	return t.cli.CancelServer(ctx, req, callOptions...)
}

type pbCancelClient struct {
	cli testpbcancelservice.Client
}

func (p pbCancelClient) CancelBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[pbapi.MockReq, pbapi.MockResp], err error) {
	return p.cli.CancelPbBidi(ctx, callOptions...)
}

func (p pbCancelClient) CancelClient(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.ClientStreamingClient[pbapi.MockReq, pbapi.MockResp], err error) {
	return p.cli.CancelPbClient(ctx, callOptions...)
}

func (p pbCancelClient) CancelServer(ctx context.Context, req *pbapi.MockReq, callOptions ...streamcall.Option) (stream streaming.ServerStreamingClient[pbapi.MockResp], err error) {
	return p.cli.CancelPbServer(ctx, req, callOptions...)
}

func TestThriftCancel(t *testing.T, addr string) {
	cliSvcName := "client-side cancel service"
	reqGetter := func(s string) *tenant.EchoRequest {
		return &tenant.EchoRequest{Msg: s}
	}
	resMsgSetter := func(p *tenant.EchoResponse) string {
		return p.Msg
	}

	t.Run("TTHeader Streaming", func(t *testing.T) {
		evtHdl := newClientMockEventHandler()
		ttstreamCli := testcancelservice.MustNewClient("service",
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
		)
		commonTestCancel[tenant.EchoRequest, tenant.EchoResponse](t, thriftCancelClient{ttstreamCli}, false, reqGetter, resMsgSetter, evtHdl)
	})
	t.Run("GRPC", func(t *testing.T) {
		evtHdl := newClientMockEventHandler()
		grpcCli := testcancelservice.MustNewClient("service",
			client.WithHostPorts(addr),
			client.WithTransportProtocol(transport.GRPCStreaming),
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
		)
		commonTestCancel[tenant.EchoRequest, tenant.EchoResponse](t, thriftCancelClient{grpcCli}, true, reqGetter, resMsgSetter, evtHdl)
	})
}

func TestPbCancel(t *testing.T, addr string) {
	cliSvcName := "client-side cancel service"
	reqGetter := func(s string) *pbapi.MockReq {
		return &pbapi.MockReq{Message: s}
	}
	resMsgSetter := func(p *pbapi.MockResp) string {
		return p.Message
	}

	// Todo: add ttstream test back when ttstream with pb is merged
	//t.Run("TTHeader Streaming", func(t *testing.T) {
	//	ttstreamCli := testpbcancelservice.MustNewClient("service",
	//		client.WithHostPorts(addr),
	//		client.WithTransportProtocol(transport.PurePayload),
	//		client.WithTransportProtocol(transport.TTHeaderStreaming),
	//		client.WithMetaHandler(transmeta.ClientHTTP2Handler),
	//		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
	//		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
	//			ServiceName: cliSvcName,
	//		}),
	//	)
	//	commonTestCancel[pbapi.MockReq, pbapi.MockResp](t, pbCancelClient{ttstreamCli}, false, reqGetter, resMsgSetter)
	//})
	t.Run("GRPC", func(t *testing.T) {
		evtHdl := newClientMockEventHandler()
		grpcCli := testpbcancelservice.MustNewClient("service",
			client.WithHostPorts(addr),
			client.WithTransportProtocol(transport.GRPCStreaming),
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
		)
		commonTestCancel[pbapi.MockReq, pbapi.MockResp](t, pbCancelClient{grpcCli}, true, reqGetter, resMsgSetter, evtHdl)
	})
}

func commonTestCancel[Req, Res any](t *testing.T, cli cancelClient[Req, Res], isGRPC bool,
	reqGetter func(string) *Req, resMsgGetter func(*Res) string,
	evtHdl *ClientMockEventHandler,
) {
	t.Run(clientStreamingWithoutSendingAnyReq, func(t *testing.T) {
		defer func() {
			<-CancelFinishChan
			evtHdl.Reset()
		}()
		evtHdl.handleStreamFinishEvent = func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {
			test.Assert(t, evt.GRPCTrailer == nil, evt)
			test.Assert(t, evt.TTStreamTrailer == nil, evt)
		}
		CancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, clientStreamingWithoutSendingAnyReq)
		cliSt, err := cli.CancelClient(ctx)
		test.Assert(t, err == nil, err)
		cancel()
		// just for triggering cancel fast
		err = cliSt.RecvMsg(cliSt.Context(), reqGetter(clientStreamingWithoutSendingAnyReq))
		verifyClientSideCancelCase(t, cliSt.Context(), err, isGRPC, false)
		evtHdl.AssertCalledTimes(t, ClientCalledTimes{
			startTimes:      1,
			recvHeaderTimes: 0,
			recvTimes:       1,
			sendTimes:       0,
			finishTimes:     1,
		})
	})
	t.Run(clientStreamingDuringNormalInteraction, func(t *testing.T) {
		defer func() {
			<-CancelFinishChan
			evtHdl.Reset()
		}()
		evtHdl.handleStreamFinishEvent = func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {
			test.Assert(t, evt.GRPCTrailer == nil, evt)
			test.Assert(t, evt.TTStreamTrailer == nil, evt)
		}
		evtHdl.handleStreamSendEvent = func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamSendEvent) {
			test.Assert(t, !evt.Time.IsZero(), evt)
		}
		CancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, clientStreamingDuringNormalInteraction)
		cliSt, err := cli.CancelClient(ctx)
		test.Assert(t, err == nil, err)
		go func() {
			time.Sleep(cancelInterval)
			cancel()
		}()
		var sendTimes uint32
		for {
			sendTimes++
			err = cliSt.Send(cliSt.Context(), reqGetter(clientStreamingDuringNormalInteraction))
			if err == nil {
				continue
			}
			verifyClientSideCancelCase(t, cliSt.Context(), err, isGRPC, true)
			break
		}
		evtHdl.AssertCalledTimes(t, ClientCalledTimes{
			startTimes:      1,
			recvHeaderTimes: 0,
			recvTimes:       0,
			sendTimes:       sendTimes,
			finishTimes:     1,
		})
	})
	t.Run(clientStreamingDeferCancel, func(t *testing.T) {
		defer func() {
			<-CancelFinishChan
			evtHdl.AssertCalledTimes(t, ClientCalledTimes{
				startTimes:      1,
				recvHeaderTimes: 0,
				recvTimes:       0,
				sendTimes:       10,
				finishTimes:     1,
			})
			evtHdl.Reset()
		}()
		evtHdl.handleStreamFinishEvent = func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {
			test.Assert(t, evt.GRPCTrailer == nil, evt)
			test.Assert(t, evt.TTStreamTrailer == nil, evt)
		}
		CancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, clientStreamingDeferCancel)
		cliSt, err := cli.CancelClient(ctx)
		test.Assert(t, err == nil, err)
		defer cancel()
		for i := 0; i < 10; i++ {
			err = cliSt.Send(cliSt.Context(), reqGetter(clientStreamingDeferCancel))
			test.Assert(t, err == nil, err)
		}
	})
	t.Run(serverStreamingRemoteRespondingSlowly, func(t *testing.T) {
		defer func() {
			<-CancelFinishChan
			evtHdl.Reset()
		}()
		evtHdl.handleStreamFinishEvent = func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {
			test.Assert(t, evt.GRPCTrailer == nil, evt)
			test.Assert(t, evt.TTStreamTrailer == nil, evt)
		}
		CancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, serverStreamingRemoteRespondingSlowly)
		srvSt, err := cli.CancelServer(ctx, reqGetter(serverStreamingRemoteRespondingSlowly))
		test.Assert(t, err == nil, err)
		time.Sleep(cancelInterval)
		cancel()
		CancelSendChan <- struct{}{}
		_, err = srvSt.Recv(srvSt.Context())
		verifyClientSideCancelCase(t, srvSt.Context(), err, isGRPC, false)
		evtHdl.AssertCalledTimes(t, ClientCalledTimes{
			startTimes:      1,
			recvHeaderTimes: 0,
			recvTimes:       1,
			sendTimes:       1,
			finishTimes:     1,
		})
	})
	t.Run(serverStreamingDuringNormalInteraction, func(t *testing.T) {
		defer func() {
			<-CancelFinishChan
			evtHdl.Reset()
		}()
		evtHdl.handleStreamFinishEvent = func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {
			test.Assert(t, evt.GRPCTrailer == nil, evt)
			test.Assert(t, evt.TTStreamTrailer == nil, evt)
		}
		CancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, serverStreamingDuringNormalInteraction)
		srvSt, err := cli.CancelServer(ctx, reqGetter(serverStreamingDuringNormalInteraction))
		test.Assert(t, err == nil, err)
		go func() {
			// It takes some time to wait for the remote to reply.
			time.Sleep(cancelInterval * 3)
			cancel()
		}()
		var recvTimes uint32
		for {
			recvTimes++
			res, rErr := srvSt.Recv(srvSt.Context())
			if rErr == nil {
				test.Assert(t, resMsgGetter(res) == serverStreamingDuringNormalInteraction, res)
				continue
			}
			verifyClientSideCancelCase(t, srvSt.Context(), rErr, isGRPC, false)
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
	t.Run(serverStreamingDeferCancel, func(t *testing.T) {
		var recvTimes uint32
		defer func() {
			<-CancelFinishChan
			evtHdl.AssertCalledTimes(t, ClientCalledTimes{
				startTimes:      1,
				recvHeaderTimes: 1,
				recvTimes:       recvTimes,
				sendTimes:       1,
				finishTimes:     1,
			})
			evtHdl.Reset()
		}()
		evtHdl.handleStreamFinishEvent = func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {
			test.Assert(t, evt.GRPCTrailer == nil, evt)
			test.Assert(t, evt.TTStreamTrailer == nil, evt)
		}
		CancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, serverStreamingDeferCancel)
		srvSt, err := cli.CancelServer(ctx, reqGetter(serverStreamingDeferCancel))
		test.Assert(t, err == nil, err)
		defer cancel()
		for {
			recvTimes++
			res, err := srvSt.Recv(srvSt.Context())
			test.Assert(t, err == nil, err)
			// special exit sign that has been agreed with downstream in the business
			if resMsgGetter(res) == bizSpecialMessage {
				break
			}
			test.Assert(t, resMsgGetter(res) == serverStreamingDeferCancel)
		}
	})
	t.Run(bidiStreamingIndependentSendRecv, func(t *testing.T) {
		defer func() {
			<-CancelFinishChan
			evtHdl.Reset()
		}()
		evtHdl.handleStreamFinishEvent = func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {
			test.Assert(t, evt.GRPCTrailer == nil, evt)
			test.Assert(t, evt.TTStreamTrailer == nil, evt)
		}
		CancelTChan <- t
		ctx, cancel := context.WithCancel(context.Background())
		ctx = setScenario(ctx, bidiStreamingIndependentSendRecv)
		bidiSt, err := cli.CancelBidi(ctx)
		test.Assert(t, err == nil, err)
		go func() {
			time.Sleep(cancelInterval)
			cancel()
		}()
		var sendTimes, recvTimes uint32
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for {
				sendTimes++
				sErr := bidiSt.Send(bidiSt.Context(), reqGetter(bidiStreamingIndependentSendRecv))
				if sErr == nil {
					continue
				}
				verifyClientSideCancelCase(t, bidiSt.Context(), sErr, isGRPC, true)
				break
			}
		}()
		go func() {
			defer wg.Done()
			for {
				recvTimes++
				res, rErr := bidiSt.Recv(bidiSt.Context())
				if rErr == nil {
					test.Assert(t, resMsgGetter(res) == bidiStreamingIndependentSendRecv, res)
					continue
				}
				verifyClientSideCancelCase(t, bidiSt.Context(), rErr, isGRPC, false)
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
