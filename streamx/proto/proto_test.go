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

package streamx_proto_test

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/endpoint/cep"
	"github.com/cloudwego/kitex/pkg/endpoint/sep"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	echo "github.com/cloudwego/kitex-tests/streamx/kitex_gen/echopb"
	"github.com/cloudwego/kitex-tests/streamx/kitex_gen/echopb/testservice"
)

const maxReceiveTimes = 10

type serviceImpl struct {
}

func (s *serviceImpl) Unary(ctx context.Context, req *echo.EchoClientRequest) (r *echo.EchoClientResponse, err error) {
	if v, ok := metainfo.GetValue(ctx, "METAKEY"); !ok || v != "METAVALUE" {
		return nil, errors.New("metainfo is not set")
	}
	if req.Message == "test_unary_timeout" {
		time.Sleep(500 * time.Millisecond)
		return &echo.EchoClientResponse{
			Message: "pong",
		}, nil
	}
	if req.Message != "ping" {
		return nil, errors.New("invalid message")
	}
	if ctx.Value("test_unary_middleware_builder").(string) != "test_unary_middleware_builder" {
		return nil, errors.New("test_unary_middleware_builder is not set")
	}
	if ctx.Value("test_unary_middleware").(string) != "test_unary_middleware" {
		return nil, errors.New("test_unary_middleware is not set")
	}
	ri := rpcinfo.GetRPCInfo(ctx)
	if ri.Config().TransportProtocol()&transport.GRPC == transport.GRPC {
		// create and send header
		header := metadata.Pairs("unaryheaderkey", "unaryheadervalue")
		nphttp2.SendHeader(ctx, header)
		// create and set trailer
		trailer := metadata.Pairs("unarytrailerkey", "unarytrailervalue")
		nphttp2.SetTrailer(ctx, trailer)
	}
	return &echo.EchoClientResponse{
		Message: "pong",
	}, nil
}

func (s *serviceImpl) EchoBidi(ctx context.Context, stream echo.TestService_EchoBidiServer) (err error) {
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
				return nil
			}
			return err
		}
		receivedTimes++
		if req.Message != "ping" {
			return errors.New("invalid message")
		}
		err = stream.Send(ctx, &echo.EchoClientResponse{
			Message: "pong",
		})
		if err != nil {
			return err
		}
	}
}

func (s *serviceImpl) EchoClient(ctx context.Context, stream echo.TestService_EchoClientServer) (err error) {
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
				err = stream.SendAndClose(ctx, &echo.EchoClientResponse{
					Message: "pong",
				})
				if err != nil {
					return err
				}
				if b != 1 {
					return errors.New("send middleware call times is not 1")
				}
				if d != 1 {
					return errors.New("send middleware builder call times is not 1")
				}
			}
			return err
		}
		receivedTimes++
		if req.Message != "ping" {
			return errors.New("invalid message")
		}
	}
}

func (s *serviceImpl) EchoServer(ctx context.Context, req *echo.EchoClientRequest, stream echo.TestService_EchoServerServer) (err error) {
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
	if req.Message != "ping" {
		return errors.New("invalid message")
	}
	err = stream.SendHeader(map[string]string{
		"headerkey1": "headervalue1",
	})
	if err != nil {
		return err
	}
	for i := 0; i < maxReceiveTimes; i++ {
		err := stream.Send(ctx, &echo.EchoClientResponse{
			Message: "pong",
		})
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
	return nil
}

func runServer(listenaddr string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", listenaddr)
	svr := testservice.NewServer(&serviceImpl{}, server.WithServiceAddr(addr), server.WithExitWaitTime(1*time.Second),
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
		})))
	go func() {
		if err := svr.Run(); err != nil {
			println(err)
		}
	}()
	return svr
}

var thriftTestAddr string

func TestMain(m *testing.M) {
	thriftTestAddr = serverutils.NextListenAddr()
	klog.SetLevel(klog.LevelFatal)
	svc := runServer(thriftTestAddr)
	serverutils.Wait(thriftTestAddr)
	m.Run()
	svc.Stop()
}

func TestGRPCProtobuf(t *testing.T) {
	runClient(t, false)
}

// TODO: support protobuf payload in ttheader streaming
//func TestTTHeaderStreaming(t *testing.T) {
//	runClient(t, true)
//}

func runClient(t *testing.T, isTTHeaderStreaming bool) {
	prot := transport.TTHeader
	if isTTHeaderStreaming {
		// ttheader streaming has higher priority than grpc streaming
		prot |= transport.TTHeaderStreaming
	}
	cli := testservice.MustNewClient("service", client.WithHostPorts(thriftTestAddr),
		client.WithTransportProtocol(prot),
		client.WithMetaHandler(transmeta.ClientHTTP2Handler), client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithUnaryOptions(client.WithUnaryRPCTimeout(200*time.Millisecond),
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
		})))

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

	_, err := cli.Unary(ctx, &echo.EchoClientRequest{
		Message: "test_unary_timeout",
	})
	test.Assert(t, kerrors.IsTimeoutError(err))

	// test uanry middleware
	a, b, c, d, e, f, g, h = 0, 0, 0, 0, 0, 0, 0, 0
	var header, trailer metadata.MD
	ctx = nphttp2.GRPCHeader(ctx, &header)
	ctx = nphttp2.GRPCTrailer(ctx, &trailer)
	res, err := cli.Unary(ctx, &echo.EchoClientRequest{Message: "ping"})
	test.Assert(t, err == nil)
	test.Assert(t, res.Message == "pong")
	test.Assert(t, a == 1)
	test.Assert(t, b == 1)
	if !isTTHeaderStreaming {
		test.Assert(t, header["unaryheaderkey"][0] == "unaryheadervalue")
		// test.Assert(t, trailer["unarytrailerkey"][0] == "unarytrailervalue") // fix a bug, could be removed if decode grpc receive buffer twice is done.
	}

	// test bidi bidiStream middleware
	a, b, c, d, e, f, g, h = 0, 0, 0, 0, 0, 0, 0, 0
	bidiStream, err := cli.EchoBidi(ctx)
	test.Assert(t, err == nil)
	for i := 0; i < maxReceiveTimes; i++ {
		err = bidiStream.Send(ctx, &echo.EchoClientRequest{
			Message: "ping",
		})
		test.Assert(t, err == nil)
		res, err = bidiStream.Recv(ctx)
		test.Assert(t, err == nil)
		test.Assert(t, res.Message == "pong")
	}
	err = bidiStream.CloseSend(ctx)
	test.Assert(t, err == nil)
	_, err = bidiStream.Recv(ctx)
	test.Assert(t, err == io.EOF)
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

	// test client bidiStream middleware
	a, b, c, d, e, f, g, h = 0, 0, 0, 0, 0, 0, 0, 0
	cliStream, err := cli.EchoClient(ctx)
	test.Assert(t, err == nil)
	for i := 0; i < maxReceiveTimes; i++ {
		err = cliStream.Send(ctx, &echo.EchoClientRequest{
			Message: "ping",
		})
		test.Assert(t, err == nil)
	}
	res, err = cliStream.CloseAndRecv(ctx)
	test.Assert(t, err == nil)
	test.Assert(t, res.Message == "pong")
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

	// test server bidiStream middleware
	a, b, c, d, e, f, g, h = 0, 0, 0, 0, 0, 0, 0, 0
	serverStream, err := cli.EchoServer(ctx, &echo.EchoClientRequest{
		Message: "ping",
	})
	test.Assert(t, err == nil)
	for i := 0; i < maxReceiveTimes; i++ {
		res, err = serverStream.Recv(ctx)
		test.Assert(t, err == nil)
		test.Assert(t, res.Message == "pong")
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
}
