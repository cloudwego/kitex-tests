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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/client/callopt/streamcall"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/endpoint/cep"
	"github.com/cloudwego/kitex/pkg/endpoint/sep"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/codes"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/status"
	"github.com/cloudwego/kitex/pkg/remote/trans/ttstream"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi/mock"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/echoservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
)

const maxReceiveTimes = 10

type normalThriftImpl struct {
	tenant.EchoService
}

func (s *normalThriftImpl) Echo(ctx context.Context, req *tenant.EchoRequest) (r *tenant.EchoResponse, err error) {
	return commonNormalUnaryImpl[tenant.EchoRequest, tenant.EchoResponse](ctx, req,
		func(t *tenant.EchoRequest) string {
			return t.Msg
		},
		func(s string) *tenant.EchoResponse {
			return &tenant.EchoResponse{Msg: s}
		},
	)
}

func (s *normalThriftImpl) EchoBidi(ctx context.Context, stream tenant.EchoService_EchoBidiServer) (err error) {
	return commonNormalBidiImpl[tenant.EchoRequest, tenant.EchoResponse](ctx, stream,
		func(t *tenant.EchoRequest) string {
			return t.Msg
		},
		func(s string) *tenant.EchoResponse {
			return &tenant.EchoResponse{
				Msg: s,
			}
		},
	)
}

func (s *normalThriftImpl) EchoClient(ctx context.Context, stream tenant.EchoService_EchoClientServer) (err error) {
	return commonNormalClientImpl[tenant.EchoRequest, tenant.EchoResponse, tenant.EchoService_EchoClientServer](ctx, stream,
		func(t *tenant.EchoRequest) string {
			return t.Msg
		},
		func(s string) *tenant.EchoResponse {
			return &tenant.EchoResponse{Msg: s}
		},
	)
}

func (s *normalThriftImpl) EchoServer(ctx context.Context, req *tenant.EchoRequest, stream tenant.EchoService_EchoServerServer) (err error) {
	return commonNormalServerImpl[tenant.EchoRequest, tenant.EchoResponse, tenant.EchoService_EchoServerServer](ctx, req, stream,
		func(t *tenant.EchoRequest) string {
			return t.Msg
		},
		func(s string) *tenant.EchoResponse {
			return &tenant.EchoResponse{Msg: s}
		},
	)
}

type normalPbImpl struct {
	evtHdl *ServerMockEventHandler
}

func (n normalPbImpl) UnaryTest(ctx context.Context, req *pbapi.MockReq) (res *pbapi.MockResp, err error) {
	return commonNormalUnaryImpl[pbapi.MockReq, pbapi.MockResp](ctx, req,
		func(p *pbapi.MockReq) string {
			return p.Message
		},
		func(s string) *pbapi.MockResp {
			return &pbapi.MockResp{Message: s}
		},
	)
}

func (n normalPbImpl) ClientStreamingTest(ctx context.Context, stream pbapi.Mock_ClientStreamingTestServer) (err error) {
	return commonNormalClientImpl[pbapi.MockReq, pbapi.MockResp, pbapi.Mock_ClientStreamingTestServer](ctx, stream,
		func(p *pbapi.MockReq) string {
			return p.Message
		},
		func(s string) *pbapi.MockResp {
			return &pbapi.MockResp{Message: s}
		},
	)
}

func (n normalPbImpl) ServerStreamingTest(ctx context.Context, req *pbapi.MockReq, stream pbapi.Mock_ServerStreamingTestServer) (err error) {
	return commonNormalServerImpl[pbapi.MockReq, pbapi.MockResp, pbapi.Mock_ServerStreamingTestServer](ctx, req, stream,
		func(p *pbapi.MockReq) string {
			return p.Message
		},
		func(s string) *pbapi.MockResp {
			return &pbapi.MockResp{Message: s}
		},
	)
}

func (n normalPbImpl) BidirectionalStreamingTest(ctx context.Context, stream pbapi.Mock_BidirectionalStreamingTestServer) (err error) {
	return commonNormalBidiImpl[pbapi.MockReq, pbapi.MockResp, pbapi.Mock_BidirectionalStreamingTestServer](ctx, stream,
		func(p *pbapi.MockReq) string {
			return p.Message
		},
		func(s string) *pbapi.MockResp {
			return &pbapi.MockResp{Message: s}
		},
	)
}

func commonNormalUnaryImpl[Req, Res any](
	ctx context.Context, req *Req,
	reqMsgGetter func(*Req) string, resGetter func(string) *Res,
) (*Res, error) {
	if v, ok := metainfo.GetValue(ctx, "METAKEY"); !ok || v != "METAVALUE" {
		return nil, errors.New("metainfo is not set")
	}
	if reqMsgGetter(req) == "test_unary_timeout" {
		time.Sleep(100 * time.Millisecond)
		return resGetter("pong"), nil
	}
	if ctx.Value("test_unary_middleware_builder").(string) != "test_unary_middleware_builder" {
		return nil, errors.New("test_unary_middleware_builder is not set")
	}
	if ctx.Value("test_unary_middleware").(string) != "test_unary_middleware" {
		return nil, errors.New("test_unary_middleware is not set")
	}
	if reqMsgGetter(req) != "ping" {
		return nil, errors.New("invalid message")
	}
	ri := rpcinfo.GetRPCInfo(ctx)
	var returnErrMode string
	var err error
	if ri.Config().TransportProtocol()&transport.GRPC == transport.GRPC {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if modeArr := md[ReturnErrModeKey]; len(modeArr) > 0 {
				returnErrMode = md[ReturnErrModeKey][0]
			}
		}
		err = nphttp2.SendHeader(ctx, metadata.MD{ReturnErrModeKey: []string{returnErrMode}})
		if err != nil {
			return nil, err
		}
		err = nphttp2.SetTrailer(ctx, metadata.MD{ReturnErrModeKey: []string{returnErrMode}})
		if err != nil {
			return nil, err
		}
		err = GRPCReturnErr(returnErrMode)
	}
	return resGetter("pong"), err
}

func commonNormalClientImpl[Req, Res any, Stream streaming.ClientStreamingServer[Req, Res]](
	ctx context.Context, stream Stream,
	reqMsgGetter func(*Req) string, resGetter func(string) *Res,
) (err error) {
	ri := rpcinfo.GetRPCInfo(ctx)
	var returnErrMode string
	isGRPC := ri.Config().TransportProtocol()&transport.GRPC == transport.GRPC
	if isGRPC {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || md["metadatakey"][0] != "metadatavalue" {
			return errors.New("metadata is not set")
		}
		if modeArr := md[ReturnErrModeKey]; len(modeArr) > 0 {
			returnErrMode = md[ReturnErrModeKey][0]
		}
	}
	if v, ok := metainfo.GetValue(ctx, "METAKEY"); !ok || v != "METAVALUE" {
		return errors.New("metainfo is not set")
	}
	var a, b, c, d int
	ctx = context.WithValue(ctx, "test_stream_recv_middleware", &a)
	ctx = context.WithValue(ctx, "test_stream_send_middleware", &b)
	ctx = context.WithValue(ctx, "test_stream_recv_middleware_builder", &c)
	ctx = context.WithValue(ctx, "test_stream_send_middleware_builder", &d)
	err = stream.SetHeader(map[string]string{
		"headerkey": "headervalue",
	})
	if err != nil {
		return err
	}
	err = stream.SetTrailer(map[string]string{
		"trailerkey": "trailervalue",
	})
	if err != nil {
		return err
	}
	var receivedTimes int
	for {
		req, err := stream.Recv(ctx)
		if err != nil {
			if err == io.EOF {
				if receivedTimes != maxReceiveTimes {
					return errors.New("received times is not maxReceiveTimes")
				}
				if a != maxReceiveTimes+1 {
					return errors.New("recv middleware call times is not maxReceiveTimes")
				}
				if c != maxReceiveTimes+1 {
					return errors.New("recv middleware builder call times is not maxReceiveTimes")
				}
				if *ri.Invocation().Extra("test_stream_recv_event").(*int) != maxReceiveTimes {
					return errors.New("recv event call times is not maxReceiveTimes")
				}
				err = stream.SendAndClose(ctx, resGetter("pong"))
				if err != nil {
					return err
				}
				if b != 1 {
					return errors.New("send middleware call times is not 1")
				}
				if d != 1 {
					return errors.New("send middleware builder call times is not 1")
				}
				if *ri.Invocation().Extra("test_stream_send_event").(*int) != 1 {
					return errors.New("send event call times is not 1")
				}
				if isGRPC {
					return GRPCReturnErr(returnErrMode)
				}
			}
			return err
		}
		receivedTimes++
		if reqMsgGetter(req) != "ping" {
			return errors.New("invalid message")
		}
	}
}

func commonNormalServerImpl[Req, Res any, Stream streaming.ServerStreamingServer[Res]](
	ctx context.Context, req *Req, stream Stream,
	reqMsgGetter func(*Req) string, resGetter func(string) *Res,
) (err error) {
	ri := rpcinfo.GetRPCInfo(ctx)
	if ri.Config().TransportProtocol()&transport.GRPC == transport.GRPC {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || md["metadatakey"][0] != "metadatavalue" {
			return errors.New("metadata is not set")
		}
	}
	if v, ok := metainfo.GetValue(ctx, "METAKEY"); !ok || v != "METAVALUE" {
		return errors.New("metainfo is not set")
	}
	err = stream.SetHeader(map[string]string{
		"headerkey": "headervalue",
	})
	if err != nil {
		return err
	}
	err = stream.SetTrailer(map[string]string{
		"trailerkey": "trailervalue",
	})
	if err != nil {
		return err
	}
	var b, d int
	ctx = context.WithValue(ctx, "test_stream_send_middleware", &b)
	ctx = context.WithValue(ctx, "test_stream_send_middleware_builder", &d)
	if reqMsgGetter(req) != "ping" {
		return errors.New("invalid message")
	}
	err = stream.SendHeader(map[string]string{
		"headerkey1": "headervalue1",
	})
	if err != nil {
		return err
	}
	for i := 0; i < maxReceiveTimes; i++ {
		err := stream.Send(ctx, resGetter("pong"))
		if err != nil {
			return err
		}
	}
	if b != maxReceiveTimes {
		return errors.New("send middleware call times is not maxReceiveTimes")
	}
	if d != maxReceiveTimes {
		return errors.New("send middleware builder call times is not maxReceiveTimes")
	}
	if *ri.Invocation().Extra("test_stream_recv_event").(*int) != 1 {
		return errors.New("recv event call times is not 1")
	}
	if *ri.Invocation().Extra("test_stream_send_event").(*int) != maxReceiveTimes {
		return errors.New("send event call times is not maxReceiveTimes")
	}
	return nil
}

func commonNormalBidiImpl[Req, Res any, Stream streaming.BidiStreamingServer[Req, Res]](
	ctx context.Context, stream Stream,
	reqMsgGetter func(*Req) string, resGetter func(string) *Res,
) (err error) {
	ri := rpcinfo.GetRPCInfo(ctx)
	if ri.Config().TransportProtocol()&transport.GRPC == transport.GRPC {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || md["metadatakey"][0] != "metadatavalue" {
			return errors.New("metadata is not set")
		}
	}
	if v, ok := metainfo.GetValue(ctx, "METAKEY"); !ok || v != "METAVALUE" {
		return errors.New("metainfo is not set")
	}
	if ctx.Value("test_stream_middleware_builder").(string) != "test_stream_middleware_builder" {
		return errors.New("test_stream_middleware_builder is not set")
	}
	if ctx.Value("test_stream_middleware").(string) != "test_stream_middleware" {
		return errors.New("test_stream_middleware is not set")
	}
	var a, b, c, d int
	ctx = context.WithValue(ctx, "test_stream_recv_middleware", &a)
	ctx = context.WithValue(ctx, "test_stream_send_middleware", &b)
	ctx = context.WithValue(ctx, "test_stream_recv_middleware_builder", &c)
	ctx = context.WithValue(ctx, "test_stream_send_middleware_builder", &d)
	err = stream.SetHeader(map[string]string{
		"headerkey": "headervalue",
	})
	if err != nil {
		return err
	}
	err = stream.SetTrailer(map[string]string{
		"trailerkey": "trailervalue",
	})
	if err != nil {
		return err
	}
	var receivedTimes int
	for {
		req, err := stream.Recv(ctx)
		if err != nil {
			if err == io.EOF {
				if receivedTimes != maxReceiveTimes {
					return errors.New("received times is not maxReceiveTimes")
				}
				if a != maxReceiveTimes+1 {
					return errors.New("recv middleware call times is not maxReceiveTimes")
				}
				if b != maxReceiveTimes {
					return errors.New("send middleware call times is not maxReceiveTimes")
				}
				if c != maxReceiveTimes+1 {
					return errors.New("recv middleware builder call times is not maxReceiveTimes")
				}
				if d != maxReceiveTimes {
					return errors.New("send middleware builder call times is not maxReceiveTimes")
				}
				if *ri.Invocation().Extra("test_stream_send_event").(*int) != maxReceiveTimes {
					return errors.New("send event call times is not maxReceiveTimes")
				}
				if *ri.Invocation().Extra("test_stream_recv_event").(*int) != maxReceiveTimes {
					return errors.New("recv event call times is not maxReceiveTimes")
				}
				return nil
			}
			return err
		}
		if reqMsgGetter(req) == "biz_error" {
			return kerrors.NewBizStatusError(404, "not found")
		}
		receivedTimes++
		if reqMsgGetter(req) != "ping" {
			return errors.New("invalid message")
		}
		err = stream.Send(ctx, resGetter("pong"))
		if err != nil {
			return err
		}
	}
}

func RunNormalServer(listenAddr string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", listenAddr)
	hdl := newServerMockEventHandler(nil)
	svr := echoservice.NewServer(&normalThriftImpl{}, server.WithServiceAddr(addr), server.WithExitWaitTime(1*time.Second), server.WithTracer(NewTracer()),
		server.WithStreamOptions(server.WithStreamEventHandler(
			rpcinfo.ServerStreamEventHandler{
				HandleStreamStartEvent:  hdl.HandleStreamStartEvent,
				HandleStreamRecvEvent:   hdl.HandleStreamRecvEvent,
				HandleStreamSendEvent:   hdl.HandleStreamSendEvent,
				HandleStreamFinishEvent: hdl.HandleStreamFinishEvent,
			},
		)),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler), server.WithMetaHandler(transmeta.ServerHTTP2Handler),
		server.WithUnaryOptions(server.WithUnaryMiddleware(func(next endpoint.UnaryEndpoint) endpoint.UnaryEndpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				ctx = context.WithValue(ctx, "test_unary_middleware", "test_unary_middleware")
				return next(ctx, req, resp)
			}
		}), server.WithUnaryMiddlewareBuilder(func(ctx context.Context) endpoint.UnaryMiddleware {
			return func(next endpoint.UnaryEndpoint) endpoint.UnaryEndpoint {
				return func(ctx context.Context, req, resp interface{}) (err error) {
					ctx = context.WithValue(ctx, "test_unary_middleware_builder", "test_unary_middleware_builder")
					return next(ctx, req, resp)
				}
			}
		})), server.WithStreamOptions(server.WithStreamMiddleware(func(next sep.StreamEndpoint) sep.StreamEndpoint {
			return func(ctx context.Context, stream streaming.ServerStream) (err error) {
				ctx = context.WithValue(ctx, "test_stream_middleware", "test_stream_middleware")
				return next(ctx, stream)
			}
		}), server.WithStreamMiddlewareBuilder(func(ctx context.Context) sep.StreamMiddleware {
			return func(next sep.StreamEndpoint) sep.StreamEndpoint {
				return func(ctx context.Context, stream streaming.ServerStream) (err error) {
					ctx = context.WithValue(ctx, "test_stream_middleware_builder", "test_stream_middleware_builder")
					return next(ctx, stream)
				}
			}
		}), server.WithStreamRecvMiddleware(func(next sep.StreamRecvEndpoint) sep.StreamRecvEndpoint {
			return func(ctx context.Context, stream streaming.ServerStream, req interface{}) (err error) {
				ri := rpcinfo.GetRPCInfo(ctx)
				if ri.Invocation().StreamingMode() == serviceinfo.StreamingServer {
					return next(ctx, stream, req)
				}
				count := ctx.Value("test_stream_recv_middleware").(*int)
				*count++
				return next(ctx, stream, req)
			}
		}), server.WithStreamRecvMiddlewareBuilder(func(ctx context.Context) sep.StreamRecvMiddleware {
			return func(next sep.StreamRecvEndpoint) sep.StreamRecvEndpoint {
				return func(ctx context.Context, stream streaming.ServerStream, req interface{}) (err error) {
					ri := rpcinfo.GetRPCInfo(ctx)
					if ri.Invocation().StreamingMode() == serviceinfo.StreamingServer {
						return next(ctx, stream, req)
					}
					count := ctx.Value("test_stream_recv_middleware_builder").(*int)
					*count++
					return next(ctx, stream, req)
				}
			}
		}), server.WithStreamSendMiddleware(func(next sep.StreamSendEndpoint) sep.StreamSendEndpoint {
			return func(ctx context.Context, stream streaming.ServerStream, req interface{}) (err error) {
				count := ctx.Value("test_stream_send_middleware").(*int)
				*count++
				return next(ctx, stream, req)
			}
		}), server.WithStreamSendMiddlewareBuilder(func(ctx context.Context) sep.StreamSendMiddleware {
			return func(next sep.StreamSendEndpoint) sep.StreamSendEndpoint {
				return func(ctx context.Context, stream streaming.ServerStream, req interface{}) (err error) {
					count := ctx.Value("test_stream_send_middleware_builder").(*int)
					*count++
					return next(ctx, stream, req)
				}
			}
		})), server.WithTTHeaderStreamingOptions(server.WithTTHeaderStreamingTransportOptions(
			ttstream.WithServerHeaderFrameHandler(tthh))))
	if err := mock.RegisterService(svr, &normalPbImpl{}); err != nil {
		panic(err)
	}

	go func() {
		if err := svr.Run(); err != nil {
			println(err)
		}
	}()
	return svr
}

var basicOpts = []client.Option{
	client.WithTracer(NewTracer()),
	client.WithMetaHandler(transmeta.ClientHTTP2Handler), client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
	client.WithUnaryOptions(client.WithUnaryRPCTimeout(10*time.Millisecond),
		client.WithUnaryMiddleware(func(next endpoint.UnaryEndpoint) endpoint.UnaryEndpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				count := ctx.Value("test_unary_middleware").(*int)
				*count++
				return next(ctx, req, resp)
			}
		}),
		client.WithUnaryMiddlewareBuilder(func(ctx context.Context) endpoint.UnaryMiddleware {
			return func(next endpoint.UnaryEndpoint) endpoint.UnaryEndpoint {
				return func(ctx context.Context, req, resp interface{}) (err error) {
					count := ctx.Value("test_unary_middleware_builder").(*int)
					*count++
					return next(ctx, req, resp)
				}
			}
		})),
	client.WithStreamOptions(client.WithStreamMiddleware(func(next cep.StreamEndpoint) cep.StreamEndpoint {
		return func(ctx context.Context) (stream streaming.ClientStream, err error) {
			count := ctx.Value("test_stream_middleware").(*int)
			*count++
			return next(ctx)
		}
	}), client.WithStreamMiddlewareBuilder(func(ctx context.Context) cep.StreamMiddleware {
		return func(next cep.StreamEndpoint) cep.StreamEndpoint {
			return func(ctx context.Context) (stream streaming.ClientStream, err error) {
				count := ctx.Value("test_stream_middleware_builder").(*int)
				*count++
				return next(ctx)
			}
		}
	}), client.WithStreamSendMiddleware(func(next cep.StreamSendEndpoint) cep.StreamSendEndpoint {
		return func(ctx context.Context, stream streaming.ClientStream, message interface{}) (err error) {
			count := ctx.Value("test_stream_send_middleware").(*int)
			*count++
			return next(ctx, stream, message)
		}
	}), client.WithStreamSendMiddlewareBuilder(func(ctx context.Context) cep.StreamSendMiddleware {
		return func(next cep.StreamSendEndpoint) cep.StreamSendEndpoint {
			return func(ctx context.Context, stream streaming.ClientStream, message interface{}) (err error) {
				count := ctx.Value("test_stream_send_middleware_builder").(*int)
				*count++
				return next(ctx, stream, message)
			}
		}
	}), client.WithStreamRecvMiddleware(func(next cep.StreamRecvEndpoint) cep.StreamRecvEndpoint {
		return func(ctx context.Context, stream streaming.ClientStream, message interface{}) (err error) {
			count := ctx.Value("test_stream_recv_middleware").(*int)
			*count++
			return next(ctx, stream, message)
		}
	}), client.WithStreamRecvMiddlewareBuilder(func(ctx context.Context) cep.StreamRecvMiddleware {
		return func(next cep.StreamRecvEndpoint) cep.StreamRecvEndpoint {
			return func(ctx context.Context, stream streaming.ClientStream, message interface{}) (err error) {
				count := ctx.Value("test_stream_recv_middleware_builder").(*int)
				*count++
				return next(ctx, stream, message)
			}
		}
	})),
}

func TestThriftNormal(t *testing.T, addr string) {
	testcases := []struct {
		desc      string
		cliType   clientType
		prot      transport.Protocol
		extraTest func(t *testing.T)
	}{
		{
			desc:    "GRPCStreaming thrift",
			cliType: thriftPingPong_gRPCStreaming,
			prot:    transport.TTHeader | transport.GRPCStreaming,
		},
		{
			desc:    "GRPC thrift",
			cliType: gRPCUnary_gRPCStreaming,
			prot:    transport.GRPC,
		},
		{
			desc:    "TTHeader Streaming thrift",
			cliType: thriftPingPong_TTHeaderStreaming,
			prot:    transport.TTHeader,
			extraTest: func(t *testing.T) {
				test.Assert(t, atomic.LoadUint32(&tthh.executed) == 1)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			var opts []client.Option
			opts = append(opts, basicOpts...)
			opts = append(opts, client.WithHostPorts(addr),
				client.WithTracer(newStreamingProtocolCheckTracer(t, tc.cliType)),
				client.WithTransportProtocol(tc.prot),
			)

			evtHdl := newClientMockEventHandler()
			opts = append(opts, client.WithStreamOptions(
				client.WithStreamEventHandler(rpcinfo.ClientStreamEventHandler{
					HandleStreamStartEvent:      evtHdl.HandleStreamStartEvent,
					HandleStreamRecvHeaderEvent: evtHdl.HandleStreamRecvHeaderEvent,
					HandleStreamRecvEvent:       evtHdl.HandleStreamRecvEvent,
					HandleStreamSendEvent:       evtHdl.HandleStreamSendEvent,
					HandleStreamFinishEvent:     evtHdl.HandleStreamFinishEvent,
				}),
			))

			cli := echoservice.MustNewClient("service", opts...)
			commonTestNormal[tenant.EchoRequest, tenant.EchoResponse](t, thriftNormalClient{cli}, tc.cliType,
				func(s string) *tenant.EchoRequest {
					return &tenant.EchoRequest{Msg: s}
				},
				func(t *tenant.EchoResponse) string {
					return t.Msg
				},
				evtHdl,
			)
			if tc.extraTest != nil {
				tc.extraTest(t)
			}
		})
	}
}

func TestPbNormal(t *testing.T, addr string) {
	testcases := []struct {
		desc      string
		cliType   clientType
		protOpts  []client.Option
		extraTest func(t *testing.T)
	}{
		{
			desc:    "GRPCStreaming thrift",
			cliType: thriftPingPong_gRPCStreaming,
			protOpts: []client.Option{
				client.WithTransportProtocol(transport.PurePayload),
				client.WithTransportProtocol(transport.TTHeader | transport.GRPCStreaming),
			},
		},
		{
			desc:    "GRPC thrift",
			cliType: gRPCUnary_gRPCStreaming,
			protOpts: []client.Option{
				client.WithTransportProtocol(transport.GRPC),
			},
		},
		// Todo: add ttstream test back when ttstream with pb is merged
		//{
		//	desc:    "TTHeader Streaming thrift",
		//	cliType: thriftPingPong_TTHeaderStreaming,
		//	protOpts: []client.Option{
		//		client.WithTransportProtocol(transport.PurePayload),
		//		client.WithTransportProtocol(transport.TTHeader | transport.TTHeaderStreaming),
		//	},
		//	extraTest: func(t *testing.T) {
		//		test.Assert(t, atomic.LoadUint32(&tthh.executed) == 1)
		//	},
		//},
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			var opts []client.Option
			opts = append(opts, basicOpts...)
			opts = append(opts, client.WithHostPorts(addr),
				client.WithTracer(newStreamingProtocolCheckTracer(t, tc.cliType)),
			)
			opts = append(opts, tc.protOpts...)

			evtHdl := newClientMockEventHandler()
			opts = append(opts, client.WithStreamOptions(
				client.WithStreamEventHandler(rpcinfo.ClientStreamEventHandler{
					HandleStreamStartEvent:      evtHdl.HandleStreamStartEvent,
					HandleStreamRecvHeaderEvent: evtHdl.HandleStreamRecvHeaderEvent,
					HandleStreamRecvEvent:       evtHdl.HandleStreamRecvEvent,
					HandleStreamSendEvent:       evtHdl.HandleStreamSendEvent,
					HandleStreamFinishEvent:     evtHdl.HandleStreamFinishEvent,
				}),
			))

			cli := mock.MustNewClient("service", opts...)
			commonTestNormal[pbapi.MockReq, pbapi.MockResp](t, pbNormalClient{cli}, tc.cliType,
				func(s string) *pbapi.MockReq {
					return &pbapi.MockReq{Message: s}
				},
				func(t *pbapi.MockResp) string {
					return t.Message
				},
				evtHdl,
			)
			if tc.extraTest != nil {
				tc.extraTest(t)
			}
		})
	}
}

type normalClient[Req, Res any] interface {
	EchoUnary(ctx context.Context, req *Req, callOptions ...callopt.Option) (res *Res, err error)
	EchoBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[Req, Res], err error)
	EchoClient(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.ClientStreamingClient[Req, Res], err error)
	EchoServer(ctx context.Context, req *Req, callOptions ...streamcall.Option) (stream streaming.ServerStreamingClient[Res], err error)
}

type thriftNormalClient struct {
	cli echoservice.Client
}

func (t thriftNormalClient) EchoUnary(ctx context.Context, req *tenant.EchoRequest, callOptions ...callopt.Option) (r *tenant.EchoResponse, err error) {
	return t.cli.Echo(ctx, req, callOptions...)
}

func (t thriftNormalClient) EchoBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[tenant.EchoRequest, tenant.EchoResponse], err error) {
	return t.cli.EchoBidi(ctx, callOptions...)
}

func (t thriftNormalClient) EchoClient(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.ClientStreamingClient[tenant.EchoRequest, tenant.EchoResponse], err error) {
	return t.cli.EchoClient(ctx, callOptions...)
}

func (t thriftNormalClient) EchoServer(ctx context.Context, req *tenant.EchoRequest, callOptions ...streamcall.Option) (stream streaming.ServerStreamingClient[tenant.EchoResponse], err error) {
	return t.cli.EchoServer(ctx, req, callOptions...)
}

type pbNormalClient struct {
	cli mock.Client
}

func (p pbNormalClient) EchoUnary(ctx context.Context, req *pbapi.MockReq, callOptions ...callopt.Option) (res *pbapi.MockResp, err error) {
	return p.cli.UnaryTest(ctx, req, callOptions...)
}

func (p pbNormalClient) EchoBidi(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.BidiStreamingClient[pbapi.MockReq, pbapi.MockResp], err error) {
	return p.cli.BidirectionalStreamingTest(ctx, callOptions...)
}

func (p pbNormalClient) EchoClient(ctx context.Context, callOptions ...streamcall.Option) (stream streaming.ClientStreamingClient[pbapi.MockReq, pbapi.MockResp], err error) {
	return p.cli.ClientStreamingTest(ctx, callOptions...)
}

func (p pbNormalClient) EchoServer(ctx context.Context, req *pbapi.MockReq, callOptions ...streamcall.Option) (stream streaming.ServerStreamingClient[pbapi.MockResp], err error) {
	return p.cli.ServerStreamingTest(ctx, req, callOptions...)
}

func commonTestNormal[Req, Res any](t *testing.T, cli normalClient[Req, Res], cliType clientType,
	reqGetter func(string) *Req, resMsgGetter func(*Res) string,
	evtHdl *ClientMockEventHandler,
) {
	// test unary timeout
	var a, b, c, d, e, f, g, h int
	ctx := context.WithValue(context.Background(), "test_unary_middleware", &a)
	ctx = context.WithValue(ctx, "test_unary_middleware_builder", &b)
	ctx = context.WithValue(ctx, "test_stream_middleware", &c)
	ctx = context.WithValue(ctx, "test_stream_middleware_builder", &d)
	ctx = context.WithValue(ctx, "test_stream_recv_middleware", &e)
	ctx = context.WithValue(ctx, "test_stream_recv_middleware_builder", &f)
	ctx = context.WithValue(ctx, "test_stream_send_middleware", &g)
	ctx = context.WithValue(ctx, "test_stream_send_middleware_builder", &h)
	// set metainfo kv
	ctx = metainfo.WithValue(ctx, "METAKEY", "METAVALUE")
	// set metadata kv
	ctx = metadata.AppendToOutgoingContext(ctx, "metadatakey", "metadatavalue")

	_, err := cli.EchoUnary(ctx, reqGetter("test_unary_timeout"))
	test.Assert(t, kerrors.IsTimeoutError(err))

	// test unary middleware
	a, b, c, d, e, f, g, h = 0, 0, 0, 0, 0, 0, 0, 0
	res, err := cli.EchoUnary(ctx, reqGetter("ping"))
	test.Assert(t, err == nil)
	test.Assert(t, resMsgGetter(res) == "pong")
	test.Assert(t, a == 1)
	test.Assert(t, b == 1)

	// test gRPC unary retrieving biz error
	if cliType == gRPCUnary_gRPCStreaming {
		evtHdl.Reset()
		nCtx := metadata.AppendToOutgoingContext(ctx, ReturnErrModeKey, ReturnBizErr)
		hd, tl := metadata.MD{}, metadata.MD{}
		nCtx = nphttp2.GRPCHeader(nCtx, &hd)
		nCtx = nphttp2.GRPCTrailer(nCtx, &tl)
		res, err = cli.EchoUnary(nCtx, reqGetter("ping"))
		test.Assert(t, err != nil)
		bizErr, ok := kerrors.FromBizStatusError(err)
		test.Assert(t, ok, err)
		test.Assert(t, bizErr.BizStatusCode() == 10000, err)
		test.Assert(t, bizErr.BizMessage() == "biz", err)

		evtHdl.AssertCalledTimes(t, ClientCalledTimes{
			startTimes:      1,
			recvHeaderTimes: 1,
			recvTimes:       0,
			sendTimes:       0,
			finishTimes:     1,
		})
		// verify GRPCHeader and GRPCTrailer
		test.Assert(t, len(hd.Get(ReturnErrModeKey)) > 0, hd)
		test.Assert(t, hd.Get(ReturnErrModeKey)[0] == ReturnBizErr, hd)
		test.Assert(t, len(hd.Get(ReturnErrModeKey)) > 0, tl)
		test.Assert(t, hd.Get(ReturnErrModeKey)[0] == ReturnBizErr, tl)
	}

	// test bidi bidiStream middleware
	a, b, c, d, e, f, g, h = 0, 0, 0, 0, 0, 0, 0, 0
	evtHdl.Reset()
	bidiStream, err := cli.EchoBidi(ctx)
	test.Assert(t, err == nil)
	for i := 0; i < maxReceiveTimes; i++ {
		err = bidiStream.Send(ctx, reqGetter("ping"))
		test.Assert(t, err == nil)
		res, err = bidiStream.Recv(ctx)
		test.Assert(t, err == nil)
		test.Assert(t, resMsgGetter(res) == "pong")
	}
	err = bidiStream.CloseSend(ctx)
	test.Assert(t, err == nil)
	_, err = bidiStream.Recv(ctx)
	test.Assert(t, err == io.EOF, err)
	test.Assert(t, c == 1)
	test.Assert(t, d == 1)
	test.Assert(t, e == maxReceiveTimes+1)
	test.Assert(t, f == maxReceiveTimes+1)
	test.Assert(t, g == maxReceiveTimes)
	test.Assert(t, h == maxReceiveTimes)
	hd, err := bidiStream.Header()
	test.Assert(t, err == nil)
	test.Assert(t, hd["headerkey"] == "headervalue")
	td, err := bidiStream.Trailer()
	test.Assert(t, err == nil)
	test.Assert(t, td["trailerkey"] == "trailervalue")
	evtHdl.AssertCalledTimes(t, ClientCalledTimes{
		startTimes:      1,
		recvHeaderTimes: 1,
		recvTimes:       uint32(maxReceiveTimes),
		sendTimes:       uint32(maxReceiveTimes),
		finishTimes:     1,
	})

	// test client bidiStream middleware
	a, b, c, d, e, f, g, h = 0, 0, 0, 0, 0, 0, 0, 0
	evtHdl.Reset()
	cliStream, err := cli.EchoClient(ctx)
	test.Assert(t, err == nil)
	for i := 0; i < maxReceiveTimes; i++ {
		err = cliStream.Send(ctx, reqGetter("ping"))
		test.Assert(t, err == nil)
	}
	res, err = cliStream.CloseAndRecv(ctx)
	test.Assert(t, err == nil)
	test.Assert(t, resMsgGetter(res) == "pong")
	test.Assert(t, c == 1)
	test.Assert(t, d == 1)
	test.Assert(t, e == 1)
	test.Assert(t, f == 1)
	test.Assert(t, g == maxReceiveTimes)
	test.Assert(t, h == maxReceiveTimes)
	hd, err = cliStream.Header()
	test.Assert(t, err == nil)
	test.Assert(t, hd["headerkey"] == "headervalue")
	td, err = cliStream.Trailer()
	test.Assert(t, err == nil)
	test.Assert(t, td["trailerkey"] == "trailervalue")
	evtHdl.AssertCalledTimes(t, ClientCalledTimes{
		startTimes:      1,
		recvHeaderTimes: 1,
		recvTimes:       1,
		sendTimes:       uint32(maxReceiveTimes),
		finishTimes:     1,
	})

	// test gRPC client bidiStream retrieving errors
	if cliType == gRPCUnary_gRPCStreaming || cliType == thriftPingPong_gRPCStreaming {
		testcases := []string{
			ReturnGRPCErr,
			ReturnBizErr,
			ReturnOtherErr,
		}
		for _, tc := range testcases {
			nCtx := metadata.AppendToOutgoingContext(ctx, ReturnErrModeKey, tc)
			cliStream, err = cli.EchoClient(nCtx)
			test.Assert(t, err == nil, err)
			for i := 0; i < maxReceiveTimes; i++ {
				err = cliStream.Send(cliStream.Context(), reqGetter("ping"))
				test.Assert(t, err == nil)
			}
			res, err = cliStream.CloseAndRecv(cliStream.Context())
			test.Assert(t, err != nil)
			switch tc {
			case ReturnGRPCErr:
				st, ok := status.FromError(err)
				test.Assert(t, ok)
				test.Assert(t, st.Code() == codes.Internal, st)
				test.Assert(t, st.Message() == "grpc [biz error]", st)
			case ReturnBizErr:
				bizErr, ok := kerrors.FromBizStatusError(err)
				test.Assert(t, ok)
				test.Assert(t, bizErr.BizStatusCode() == 10000, bizErr)
				test.Assert(t, bizErr.BizMessage() == "biz", bizErr)
			case ReturnOtherErr:
				st, ok := status.FromError(err)
				test.Assert(t, ok)
				test.Assert(t, st.Code() == codes.Internal, st)
				test.Assert(t, strings.Contains(st.Message(), "other [biz error]"), st)
			}
		}
	}

	// test server bidiStream middleware
	a, b, c, d, e, f, g, h = 0, 0, 0, 0, 0, 0, 0, 0
	evtHdl.Reset()
	serverStream, err := cli.EchoServer(ctx, reqGetter("ping"))
	test.Assert(t, err == nil)
	for i := 0; i < maxReceiveTimes; i++ {
		res, err = serverStream.Recv(ctx)
		test.Assert(t, err == nil)
		test.Assert(t, resMsgGetter(res) == "pong")
	}
	_, err = serverStream.Recv(ctx)
	test.Assert(t, err == io.EOF)
	test.Assert(t, c == 1)
	test.Assert(t, d == 1)
	test.Assert(t, e == maxReceiveTimes+1)
	test.Assert(t, f == maxReceiveTimes+1)
	test.Assert(t, g == 1)
	test.Assert(t, h == 1)
	hd, err = serverStream.Header()
	test.Assert(t, err == nil)
	test.Assert(t, hd["headerkey"] == "headervalue")
	test.Assert(t, hd["headerkey1"] == "headervalue1")
	td, err = serverStream.Trailer()
	test.Assert(t, err == nil)
	test.Assert(t, td["trailerkey"] == "trailervalue")
	evtHdl.AssertCalledTimes(t, ClientCalledTimes{
		startTimes:      1,
		recvHeaderTimes: 1,
		recvTimes:       uint32(maxReceiveTimes),
		sendTimes:       1,
		finishTimes:     1,
	})

	// test biz error
	bidiStream, err = cli.EchoBidi(ctx)
	test.Assert(t, err == nil)
	err = bidiStream.Send(bidiStream.Context(), reqGetter("biz_error"))
	test.Assert(t, err == nil)
	_, err = bidiStream.Recv(bidiStream.Context())
	bizErr, _ := kerrors.FromBizStatusError(err)
	test.Assert(t, bizErr.BizStatusCode() == 404)
	test.Assert(t, bizErr.BizMessage() == "not found")
}

type clientType string

const (
	gRPCUnary_gRPCStreaming          clientType = "gRPCUnary_gRPCStreaming"
	thriftPingPong_gRPCStreaming     clientType = "thriftPingPong_gRPCStreaming"
	thriftPingPong_TTHeaderStreaming clientType = "thriftPingPong_TTHeaderStreaming"
)

type streamingProtocolCheckTracer struct {
	t       *testing.T
	cliType clientType
}

func newStreamingProtocolCheckTracer(t *testing.T, cliType clientType) *streamingProtocolCheckTracer {
	return &streamingProtocolCheckTracer{
		t:       t,
		cliType: cliType,
	}
}

func (tracer *streamingProtocolCheckTracer) Start(ctx context.Context) context.Context {
	return ctx
}

func (tracer *streamingProtocolCheckTracer) Finish(ctx context.Context) {
	t := tracer.t
	ri := rpcinfo.GetRPCInfo(ctx)
	test.Assert(t, ri != nil)
	mode := ri.Invocation().StreamingMode()
	prot := ri.Config().TransportProtocol()
	switch ri.To().Method() {
	case "EchoBidi":
		test.Assert(t, mode == serviceinfo.StreamingBidirectional, mode)
		switch tracer.cliType {
		case gRPCUnary_gRPCStreaming, thriftPingPong_gRPCStreaming:
			test.Assert(t, prot == transport.GRPC, prot)
		case thriftPingPong_TTHeaderStreaming:
			test.Assert(t, prot&transport.TTHeaderStreaming == transport.TTHeaderStreaming, prot)
		}
	case "EchoClient":
		test.Assert(t, mode == serviceinfo.StreamingClient, mode)
		switch tracer.cliType {
		case gRPCUnary_gRPCStreaming, thriftPingPong_gRPCStreaming:
			test.Assert(t, prot == transport.GRPC, prot)
		case thriftPingPong_TTHeaderStreaming:
			test.Assert(t, prot&transport.TTHeaderStreaming == transport.TTHeaderStreaming, prot)
		}
	case "EchoServer":
		test.Assert(t, mode == serviceinfo.StreamingServer, mode)
		switch tracer.cliType {
		case gRPCUnary_gRPCStreaming, thriftPingPong_gRPCStreaming:
			test.Assert(t, prot == transport.GRPC, prot)
		case thriftPingPong_TTHeaderStreaming:
			test.Assert(t, prot&transport.TTHeaderStreaming == transport.TTHeaderStreaming, prot)
		}
	case "PingPong", "Unary":
		test.Assert(t, mode == serviceinfo.StreamingNone, mode)
		switch tracer.cliType {
		case gRPCUnary_gRPCStreaming:
			test.Assert(t, prot == transport.GRPC, prot)
		case thriftPingPong_gRPCStreaming:
			test.Assert(t, prot&transport.GRPC == 0, prot)
		case thriftPingPong_TTHeaderStreaming:
			test.Assert(t, prot&transport.TTHeaderStreaming == 0, prot)
		}
	}
	return
}
