// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
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
	"strconv"
	"testing"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/streamclient"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/pkg/utils/kitexutil"
	"github.com/cloudwego/kitex/server"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb"
	grpcpbservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb/pbservice"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb"
	kitexpbservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb/pbservice"
)

func TestGRPCPB(t *testing.T) {
	cli := grpcpbservice.MustNewClient("service", client.WithHostPorts(grpcAddr))
	t.Run("pingpong", func(t *testing.T) {
		resp, err := cli.EchoPingPong(context.Background(), &grpc_pb.Request{
			Message: "hello",
		})
		test.Assert(t, err == nil)
		test.Assert(t, resp.Message == "hello")
	})
	t.Run("client", func(t *testing.T) {
		count := 2
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))
		stream, err := cli.EchoClient(ctx)
		test.Assert(t, err == nil)
		for i := 0; i < count; i++ {
			err = stream.Send(&grpc_pb.Request{
				Message: "hello-" + strconv.Itoa(i),
			})
			test.Assert(t, err == nil)
		}
		resp, err := stream.CloseAndRecv()
		test.Assert(t, err == nil)
		test.Assert(t, resp.Message == strconv.Itoa(count))
	})
	t.Run("server", func(t *testing.T) {
		count := 2
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))
		req := &grpc_pb.Request{
			Message: "hello",
		}
		stream, err := cli.EchoServer(ctx, req)
		test.Assert(t, err == nil)
		for i := 0; i < count; i++ {
			resp, err := stream.Recv()
			test.Assert(t, err == nil)
			test.Assert(t, resp.Message == "hello-"+strconv.Itoa(i))
		}
	})
	t.Run("bidirectional", func(t *testing.T) {
		count := 2
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))
		stream, err := cli.Echo(ctx)
		test.Assert(t, err == nil)
		for i := 0; i < count; i++ {
			req := &grpc_pb.Request{
				Message: "hello-" + strconv.Itoa(i),
			}
			err = stream.Send(req)
			test.Assert(t, err == nil)

			resp, err := stream.Recv()
			test.Assert(t, err == nil)
			test.Assert(t, resp.Message == req.Message, resp.Message)
		}
	})
}

func TestGRPCPBStreamClient(t *testing.T) {
	t.Run("client", func(t *testing.T) {
		count := 2
		cli := grpcpbservice.MustNewStreamClient("service", streamclient.WithHostPorts(grpcAddr),
			streamclient.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, req, resp interface{}) (err error) {
					_, ok := resp.(*streaming.Result)
					test.Assert(t, ok)
					return next(ctx, req, resp)
				}
			}),
			streamclient.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
				i := 0
				return func(stream streaming.Stream, message interface{}) (err error) {
					req := message.(*grpc_pb.Request)
					test.Assert(t, req.Message == "hello-"+strconv.Itoa(i))
					i++
					return next(stream, message)
				}
			}),
			streamclient.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
				return func(stream streaming.Stream, message interface{}) (err error) {
					err = next(stream, message)
					resp := message.(*grpc_pb.Response)
					test.Assert(t, resp.Message == strconv.Itoa(count))
					resp.Message += "_recv_middleware"
					return err
				}
			}))
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))
		stream, err := cli.EchoClient(ctx)
		test.Assert(t, err == nil)
		for i := 0; i < count; i++ {
			err = stream.Send(&grpc_pb.Request{
				Message: "hello-" + strconv.Itoa(i),
			})
			test.Assert(t, err == nil)
		}
		resp, err := stream.CloseAndRecv()
		test.Assert(t, err == nil)
		test.Assert(t, resp.Message == strconv.Itoa(count)+"_recv_middleware")
	})
	t.Run("server", func(t *testing.T) {
		count := 2
		cli := grpcpbservice.MustNewStreamClient("service", streamclient.WithHostPorts(grpcAddr),
			streamclient.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, req, resp interface{}) (err error) {
					_, ok := resp.(*streaming.Result)
					test.Assert(t, ok)
					return next(ctx, req, resp)
				}
			}),
			streamclient.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
				return func(stream streaming.Stream, message interface{}) (err error) {
					req := message.(*grpc_pb.Request)
					test.Assert(t, req.Message == "hello")
					return next(stream, message)
				}
			}),
			streamclient.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
				i := 0
				return func(stream streaming.Stream, message interface{}) (err error) {
					err = next(stream, message)
					resp := message.(*grpc_pb.Response)
					test.Assert(t, resp.Message == "hello-"+strconv.Itoa(i), resp.Message)
					i++
					resp.Message += "_recv_middleware"
					return err
				}
			}))
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))
		req := &grpc_pb.Request{
			Message: "hello",
		}
		stream, err := cli.EchoServer(ctx, req)
		test.Assert(t, err == nil)
		for i := 0; i < count; i++ {
			resp, err := stream.Recv()
			test.Assert(t, err == nil)
			test.Assert(t, resp.Message == "hello-"+strconv.Itoa(i)+"_recv_middleware")
		}
	})
	t.Run("bidirectional", func(t *testing.T) {
		count := 2
		cli := grpcpbservice.MustNewStreamClient("service", streamclient.WithHostPorts(grpcAddr),
			streamclient.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, req, resp interface{}) (err error) {
					_, ok := resp.(*streaming.Result)
					test.Assert(t, ok)
					return next(ctx, req, resp)
				}
			}),
			streamclient.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
				i := 0
				return func(stream streaming.Stream, message interface{}) (err error) {
					req := message.(*grpc_pb.Request)
					test.Assert(t, req.Message == "hello-"+strconv.Itoa(i))
					i++
					return next(stream, message)
				}
			}),
			streamclient.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
				i := 0
				return func(stream streaming.Stream, message interface{}) (err error) {
					err = next(stream, message)
					resp := message.(*grpc_pb.Response)
					test.Assert(t, resp.Message == "hello-"+strconv.Itoa(i))
					i++
					resp.Message += "_recv_middleware"
					return err
				}
			}))
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))
		stream, err := cli.Echo(ctx)
		test.Assert(t, err == nil)
		for i := 0; i < count; i++ {
			req := &grpc_pb.Request{
				Message: "hello-" + strconv.Itoa(i),
			}
			err = stream.Send(req)
			test.Assert(t, err == nil)

			resp, err := stream.Recv()
			test.Assert(t, err == nil)
			test.Assert(t, resp.Message == req.Message+"_recv_middleware", resp.Message)
		}
	})

	t.Run("call kitex_pb server", func(t *testing.T) {
		cli := grpcpbservice.MustNewClient("service", client.WithHostPorts(pbAddr))
		req := &grpc_pb.Request{
			Message: "To KitexPBServer.EchoPingPong",
		}
		resp, err := cli.EchoPingPong(context.Background(), req)
		test.Assert(t, err == nil)
		test.Assert(t, resp.Message == req.Message, resp.Message) // no middleware
	})
}

func TestGRPCPBServerMiddleware(t *testing.T) {
	ln := serverutils.Listen()
	addr := ln.Addr().String()
	svr := RunGRPCPBServer(&GRPCPBServiceImpl{}, ln,
		server.WithMiddleware(func(e endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				method, _ := kitexutil.GetMethod(ctx)
				_, ok := req.(*streaming.Args)
				isKitexPB, _ := metainfo.GetValue(ctx, "TEST_KITEX_PB")
				if !ok && isKitexPB != "1" {
					return fmt.Errorf("invalid request type: %T, method = %v", req, method)
				}
				if isKitexPB == "1" {
					req.(*grpcpbservice.EchoPingPongArgs).Req.Message += "_mwReq"
				}
				err = e(ctx, req, resp)
				if isKitexPB == "1" {
					resp.(*grpcpbservice.EchoPingPongResult).Success.Message += "_mwResp"
				}
				return
			}
		}),
		server.WithRecvMiddleware(func(e endpoint.RecvEndpoint) endpoint.RecvEndpoint {
			return func(stream streaming.Stream, message interface{}) (err error) {
				err = e(stream, message)
				method, _ := kitexutil.GetMethod(stream.Context())
				req := message.(*grpc_pb.Request)
				if req.Message != method {
					return fmt.Errorf("invalid message: %v, method = %v", req.Message, method)
				}
				req.Message += "_recv"
				return
			}
		}),
		server.WithSendMiddleware(func(e endpoint.SendEndpoint) endpoint.SendEndpoint {
			return func(stream streaming.Stream, message interface{}) (err error) {
				resp := message.(*grpc_pb.Response)
				resp.Message += "_send"
				return e(stream, message)
			}
		}),
	)
	defer svr.Stop()

	t.Run("pingpong", func(t *testing.T) {
		cli := grpcpbservice.MustNewClient("service", client.WithHostPorts(addr))
		resp, err := cli.EchoPingPong(context.Background(), &grpc_pb.Request{
			Message: "EchoPingPong",
		})
		test.Assert(t, err == nil)
		test.Assert(t, resp.Message == "EchoPingPong_recv_send", resp.Message)
	})

	t.Run("client", func(t *testing.T) {
		count := 2
		cli := grpcpbservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))
		stream, err := cli.EchoClient(ctx)
		test.Assert(t, err == nil)
		for i := 0; i < count; i++ {
			err = stream.Send(&grpc_pb.Request{
				Message: "EchoClient",
			})
			test.Assert(t, err == nil)
		}
		resp, err := stream.CloseAndRecv()
		test.Assert(t, err == nil)
		test.Assert(t, resp.Message == strconv.Itoa(count)+"_send", resp.Message)
	})

	t.Run("server", func(t *testing.T) {
		count := 2
		cli := grpcpbservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))
		req := &grpc_pb.Request{
			Message: "EchoServer",
		}
		stream, err := cli.EchoServer(ctx, req)
		test.Assert(t, err == nil)
		for i := 0; i < count; i++ {
			resp, err := stream.Recv()
			test.Assert(t, err == nil)
			test.Assert(t, resp.Message == req.Message+"_recv-"+strconv.Itoa(i)+"_send", resp.Message)
		}
	})

	t.Run("bidirectional", func(t *testing.T) {
		count := 2
		cli := grpcpbservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))
		stream, err := cli.Echo(ctx)
		test.Assert(t, err == nil)
		for i := 0; i < count; i++ {
			req := &grpc_pb.Request{
				Message: "Echo",
			}
			err = stream.Send(req)
			test.Assert(t, err == nil)

			resp, err := stream.Recv()
			test.Assert(t, err == nil)
			test.Assert(t, resp.Message == req.Message+"_recv_send", resp.Message)
		}
	})

	t.Run("kitex_pb client", func(t *testing.T) {
		cli := kitexpbservice.MustNewClient("service", client.WithHostPorts(addr))
		ctx := metainfo.WithValue(context.Background(), "TEST_KITEX_PB", "1")
		req := &kitex_pb.Request{
			Message: "EchoPingPong",
		}
		resp, err := cli.EchoPingPong(ctx, req)
		test.Assert(t, err == nil)
		// KitexPB requests will not be processed by recv/send middlewares
		test.Assertf(t, resp.Message == "EchoPingPong_mwReq_mwResp", resp.Message, "msg: %v", resp.Message)
	})
}

func TestGRPCPBServiceWithCompatibleMiddleware(t *testing.T) {
	ln := serverutils.Listen()
	addr := ln.Addr().String()
	svr := RunGRPCPBServer(
		&GRPCPBServiceImpl{},
		ln,
		server.WithCompatibleMiddlewareForUnary(),
		server.WithMiddleware(func(e endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				method, _ := kitexutil.GetMethod(ctx)
				klog.Infof("svr mw method: %s", method)
				if method == "EchoPingPong" {
					realReq, ok := req.(*grpcpbservice.EchoPingPongArgs)
					if !ok {
						return fmt.Errorf("invalid req type %T", req)
					}
					realReq.Req.Message += "_mwReq"
					err = e(ctx, req, resp)
					realResp, ok := resp.(*grpcpbservice.EchoPingPongResult)
					if !ok {
						return fmt.Errorf("invalid resp type %T", resp)
					}
					realResp.Success.Message += "_mwResp"
				} else {
					err = e(ctx, req, resp)
				}
				return err
			}
		}),
		server.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
			return func(stream streaming.Stream, message interface{}) (err error) {
				method, _ := kitexutil.GetMethod(stream.Context())
				klog.Infof("svr recvMW method: %s", method)
				if method == "EchoPingPong" {
					return fmt.Errorf("should not be called")
				} else {
					if err = next(stream, message); err != nil {
						return err
					}
					if _, ok := message.(*grpc_pb.Request); !ok {
						return fmt.Errorf("invalid req type %T", message)
					}
					return err
				}
			}
		}),
		server.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
			return func(stream streaming.Stream, message interface{}) (err error) {
				method, _ := kitexutil.GetMethod(stream.Context())
				klog.Infof("svr sendMW method: %s", method)
				if method == "EchoPingPong" {
					return fmt.Errorf("should not be called")
				} else {
					if err = next(stream, message); err != nil {
						return err
					}
					if _, ok := message.(*grpc_pb.Response); !ok {
						return fmt.Errorf("invalid req type %T", message)
					}
					return err
				}
			}
		}),
	)
	defer svr.Stop()

	// GRPC Over HTTP2
	streamCli := grpcpbservice.MustNewClient("service", client.WithHostPorts(addr))
	resp, err := streamCli.EchoPingPong(context.Background(), &grpc_pb.Request{Message: "hello"})
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "hello_mwReq_mwResp", resp.Message)

	// KitexPB
	cli := kitexpbservice.MustNewClient("service", client.WithHostPorts(addr))
	resp1, err := cli.EchoPingPong(context.Background(), &kitex_pb.Request{Message: "hello"})
	test.Assert(t, err == nil, err)
	test.Assert(t, resp1.Message == "hello_mwReq_mwResp", resp1.Message)

	ctx := metainfo.WithValue(context.Background(), KeyCount, "1")
	stream, err := streamCli.EchoClient(ctx)
	test.Assert(t, err == nil, err)
	err = stream.Send(&grpc_pb.Request{
		Message: "hello",
	})
	test.Assert(t, err == nil, err)
	_, err = stream.CloseAndRecv()
	test.Assert(t, err == nil, err)
}
