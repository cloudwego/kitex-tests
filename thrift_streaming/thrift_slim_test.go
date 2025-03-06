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
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/streamclient"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/remote/codec/thrift"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_slim/echo"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_slim/echo/echoservice"
)

var svrCodecOpt = server.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite))

func RunSlimThriftServer(handler echo.EchoService, ln net.Listener, opts ...server.Option) server.Server {
	opts = append(opts, server.WithListener(ln))
	opts = append(opts, server.WithExitWaitTime(time.Millisecond*10))
	opts = append(opts, server.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)))
	svr := echoservice.NewServer(handler, opts...)
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	return svr
}

func TestSlimKitexThriftBidirectional(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		streamclient.WithHostPorts(slimAddr))
	t.Run("normal", func(t *testing.T) {
		n := 3
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(n))
		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)
		wg := new(sync.WaitGroup)
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				req := &echo.EchoRequest{Message: "bidirectional-" + strconv.Itoa(i)}
				err = stream.Send(req)
				test.Assert(t, err == nil, err)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				resp, err := stream.Recv()
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Message == "bidirectional-"+strconv.Itoa(i), resp.Message)
			}
		}()
		wg.Wait()
	})
}

func TestSlimKitexThriftClient(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		streamclient.WithHostPorts(slimAddr))
	t.Run("normal", func(t *testing.T) {
		n := 3
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(n))
		stream, err := cli.EchoClient(ctx)
		test.Assert(t, err == nil, err)
		for i := 0; i < n; i++ {
			req := &echo.EchoRequest{Message: "client"}
			err = stream.Send(req)
			test.Assert(t, err == nil, err)
		}
		resp, err := stream.CloseAndRecv()
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == strconv.Itoa(n), resp.Message)
	})
}

func TestSlimKitexThriftServer(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		streamclient.WithHostPorts(slimAddr))

	t.Run("normal", func(t *testing.T) {
		n := 3
		req := &echo.EchoRequest{Message: "server"}
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(n))
		stream, err := cli.EchoServer(ctx, req)
		test.Assert(t, err == nil, err)
		for i := 0; i < n; i++ {
			resp, err := stream.Recv()
			test.Assert(t, err == nil, err)
			test.Assert(t, resp.Message == "server-"+strconv.Itoa(i), resp.Message)
		}
	})
}

func TestSlimKitexThriftEchoUnary(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		streamclient.WithHostPorts(slimAddr))

	t.Run("normal", func(t *testing.T) {
		req := &echo.EchoRequest{Message: "unary"}
		resp, err := cli.EchoUnary(context.Background(), req)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "unary", resp.Message)
	})

	t.Run("exception", func(t *testing.T) {
		ctx := metainfo.WithValue(context.Background(), "error", "unary")
		_, err := cli.EchoUnary(ctx, &echo.EchoRequest{Message: "unary"})
		test.Assert(t, err != nil, err)
	})
}

func TestSlimKitexThriftEchoPingPong(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(slimAddr),
		client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		client.WithTransportProtocol(transport.TTHeaderFramed))

	t.Run("normal", func(t *testing.T) {
		req1 := &echo.EchoRequest{Message: "hello"}
		req2 := &echo.EchoRequest{Message: "world"}
		resp, err := cli.EchoPingPong(context.Background(), req1, req2)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "hello-world", resp.Message)
	})

	t.Run("exception", func(t *testing.T) {
		ctx := metainfo.WithValue(context.Background(), KeyError, "pingpong")
		_, err := cli.EchoPingPong(ctx, &echo.EchoRequest{Message: "hello"}, &echo.EchoRequest{Message: "world"})
		test.Assert(t, err != nil, err)
	})
}

func TestSlimKitexThriftEchoOneway(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		client.WithHostPorts(slimAddr),
		client.WithTransportProtocol(transport.TTHeaderFramed))

	req := &echo.EchoRequest{Message: "oneway"}

	t.Run("normal", func(t *testing.T) {
		ctx := metainfo.WithValue(context.Background(), KeyMessage, req.Message)
		err := cli.EchoOneway(ctx, req)
		test.Assert(t, err == nil, err)
	})
	t.Run("normal", func(t *testing.T) {
		err := cli.EchoOneway(context.Background(), req)
		test.Assert(t, err != nil, err)
	})
}

func TestSlimKitexThriftPing(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		client.WithHostPorts(slimAddr),
		client.WithTransportProtocol(transport.TTHeaderFramed))
	t.Run("normal", func(t *testing.T) {
		err := cli.Ping(context.Background())
		test.Assert(t, err == nil, err)
	})
	t.Run("exception", func(t *testing.T) {
		ctx := metainfo.WithValue(context.Background(), KeyError, "ping")
		err := cli.Ping(ctx)
		test.Assert(t, err != nil, err)
	})
}

func TestSlimKitexClientMiddleware(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		client.WithHostPorts(slimAddr),
		client.WithTransportProtocol(transport.TTHeaderFramed),
		client.WithMiddleware(func(endpoint endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, args, result interface{}) (err error) {
				req := args.(*echo.EchoServiceEchoPingPongArgs).Req1
				test.Assert(t, req.Message == "request", req.Message)
				result.(*echo.EchoServiceEchoPingPongResult).Success = &echo.EchoResponse{Message: "response"}
				return nil
			}
		}),
	)
	req := &echo.EchoRequest{Message: "request"}
	resp, err := cli.EchoPingPong(context.Background(), req, &echo.EchoRequest{Message: ""})
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "response", resp.Message)
}

func TestSlimKitexStreamClientMiddlewareUnary(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		streamclient.WithHostPorts(slimAddr),
		streamclient.WithMiddleware(func(endpoint endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, args, result interface{}) (err error) {
				req := args.(*echo.EchoServiceEchoUnaryArgs).Req1
				test.Assert(t, req.Message == "request", req.Message)
				result.(*echo.EchoServiceEchoUnaryResult).Success = &echo.EchoResponse{Message: "response"}
				return nil
			}
		}),
	)
	req := &echo.EchoRequest{Message: "request"}
	resp, err := cli.EchoUnary(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "response", resp.Message)
}

func TestSlimKitexStreamClientMiddlewareBidirectional(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithHostPorts(slimAddr),
		streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		streamclient.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
			return func(stream streaming.Stream, req interface{}) (err error) {
				test.Assert(t, req.(*echo.EchoRequest).Message == "request")
				return next(stream, req)
			}
		}),
		streamclient.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
			return func(stream streaming.Stream, resp interface{}) (err error) {
				err = next(stream, resp)
				rsp := resp.(*echo.EchoResponse)
				test.Assert(t, rsp.Message == "request")
				rsp.Message = "response-modified-in-mw"
				return err
			}
		}),
	)

	ctx := metainfo.WithValue(context.Background(), KeyCount, "1")
	stream, err := cli.EchoBidirectional(ctx)
	test.Assert(t, err == nil, err)

	err = stream.Send(&echo.EchoRequest{Message: "request"})
	test.Assert(t, err == nil, err)

	resp, err := stream.Recv()
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "response-modified-in-mw", resp.Message)
}

func TestSlimKitexStreamClientMiddlewareClient(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		streamclient.WithHostPorts(slimAddr),
		streamclient.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
			return func(stream streaming.Stream, req interface{}) (err error) {
				test.Assert(t, req.(*echo.EchoRequest).Message == "request")
				return next(stream, req)
			}
		}),
		streamclient.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
			return func(stream streaming.Stream, resp interface{}) (err error) {
				err = next(stream, resp)
				rsp := resp.(*echo.EchoResponse)
				test.Assert(t, rsp.Message == "2")
				rsp.Message = "response-modified-in-mw"
				return err
			}
		}),
	)
	ctx := metainfo.WithValue(context.Background(), KeyCount, "2")
	stream, err := cli.EchoClient(ctx)
	test.Assert(t, err == nil, err)

	err = stream.Send(&echo.EchoRequest{Message: "request"})
	test.Assert(t, err == nil, err)

	err = stream.Send(&echo.EchoRequest{Message: "request"})
	test.Assert(t, err == nil, err)

	resp, err := stream.CloseAndRecv()
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "response-modified-in-mw", resp.Message)
}

func TestSlimKitexStreamClientMiddlewareServer(t *testing.T) {
	recvCount := 0
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
		streamclient.WithHostPorts(slimAddr),
		streamclient.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
			return func(stream streaming.Stream, req interface{}) (err error) {
				test.Assert(t, req.(*echo.EchoRequest).Message == "request")
				return next(stream, req)
			}
		}),
		streamclient.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
			return func(stream streaming.Stream, resp interface{}) (err error) {
				err = next(stream, resp)
				rsp := resp.(*echo.EchoResponse)
				test.Assert(t, rsp.Message == "request-"+strconv.Itoa(recvCount))
				recvCount += 1
				rsp.Message = "response-modified-in-mw"
				return err
			}
		}),
	)
	ctx := metainfo.WithValue(context.Background(), KeyCount, "2")
	stream, err := cli.EchoServer(ctx, &echo.EchoRequest{Message: "request"})
	test.Assert(t, err == nil, err)

	resp, err := stream.Recv()
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "response-modified-in-mw", resp.Message)

	resp, err = stream.Recv()
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "response-modified-in-mw", resp.Message)
}

func TestSlimKitexServerMiddleware(t *testing.T) {
	t.Run("pingpong", func(t *testing.T) {
		ln := serverutils.Listen()
		addr := ln.Addr().String()
		svr := RunSlimThriftServer(&SlimEchoServiceImpl{}, ln, svrCodecOpt,
			server.WithMiddleware(func(e endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, args, result interface{}) (err error) {
					realArgs := args.(*echo.EchoServiceEchoPingPongArgs)
					req1, req2 := realArgs.Req1, realArgs.Req2
					test.Assert(t, req1.Message == "ping")
					test.Assert(t, req2.Message == "pong")
					err = e(ctx, args, result)
					resp := result.(*echo.EchoServiceEchoPingPongResult)
					test.Assert(t, resp.Success.Message == "ping-pong")
					resp.Success.Message = "response"
					return nil
				}
			}))
		defer svr.Stop()

		cli := echoservice.MustNewClient(
			"service",
			client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
			client.WithHostPorts(addr),
			client.WithTransportProtocol(transport.TTHeaderFramed),
		)
		req1, req2 := &echo.EchoRequest{Message: "ping"}, &echo.EchoRequest{Message: "pong"}
		resp, err := cli.EchoPingPong(context.Background(), req1, req2)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "response", resp.Message)
	})

	t.Run("unary", func(t *testing.T) {
		ln := serverutils.Listen()
		addr := ln.Addr().String()
		svr := RunSlimThriftServer(&SlimEchoServiceImpl{}, ln, svrCodecOpt,
			server.WithMiddleware(func(e endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, args, result interface{}) (err error) {
					realArgs, ok := args.(*echo.EchoServiceEchoUnaryArgs)
					if !ok {
						return fmt.Errorf("invalid args type: %T", args)
					}
					if realArgs.Req1.Message != "unary" {
						return fmt.Errorf("invalid realArgs.Req1.Message: %v", realArgs.Req1.Message)
					}
					if err = e(ctx, args, result); err != nil {
						return err
					}
					resp, ok := result.(*echo.EchoServiceEchoUnaryResult)
					if !ok {
						return fmt.Errorf("invalid result type: %T", result)
					}
					if resp.Success.Message != "unary" {
						return fmt.Errorf("invalid resp.Success.Message: %v", resp.Success.Message)
					}
					resp.Success.Message = "response"
					return nil
				}
			}),
			server.WithRecvMiddleware(func(e endpoint.RecvEndpoint) endpoint.RecvEndpoint {
				return func(stream streaming.Stream, resp interface{}) (err error) {
					t.Errorf("recv middleware should not be called")
					return errors.New("recv middleware should not be called")
				}
			}),
			server.WithSendMiddleware(func(e endpoint.SendEndpoint) endpoint.SendEndpoint {
				return func(stream streaming.Stream, req interface{}) (err error) {
					t.Errorf("send middleware should not be called")
					return errors.New("send middleware should not be called")
				}
			}))
		defer svr.Stop()

		cli := echoservice.MustNewStreamClient("service",
			streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
			streamclient.WithHostPorts(addr))
		req := &echo.EchoRequest{Message: "unary"}
		resp, err := cli.EchoUnary(context.Background(), req)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "response", resp.Message)
	})

	t.Run("bidirectional streaming", func(t *testing.T) {
		ln := serverutils.Listen()
		addr := ln.Addr().String()
		svr := RunSlimThriftServer(&SlimEchoServiceImpl{}, ln, svrCodecOpt,
			server.WithMiddleware(func(e endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, args, result interface{}) (err error) {
					_, ok := args.(*streaming.Args)
					test.Assert(t, ok)
					return e(ctx, args, result)
				}
			}),
			server.WithRecvMiddleware(func(e endpoint.RecvEndpoint) endpoint.RecvEndpoint {
				return func(stream streaming.Stream, req interface{}) (err error) {
					err = e(stream, req)
					realReq, ok := req.(*echo.EchoRequest)
					test.Assert(t, ok)
					test.Assert(t, realReq.Message == "request")
					realReq.Message = "bidirectional"
					return
				}
			}),
			server.WithSendMiddleware(func(e endpoint.SendEndpoint) endpoint.SendEndpoint {
				return func(stream streaming.Stream, resp interface{}) (err error) {
					realResp, ok := resp.(*echo.EchoResponse)
					test.Assert(t, ok)
					test.Assert(t, realResp.Message == "bidirectional")
					realResp.Message = "response"
					return e(stream, realResp)
				}
			}))
		defer svr.Stop()

		cli := echoservice.MustNewStreamClient("service",
			streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
			streamclient.WithHostPorts(addr))
		ctx := metainfo.WithValue(context.Background(), KeyCount, "1")
		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)

		req := &echo.EchoRequest{Message: "request"}
		err = stream.Send(req)
		test.Assert(t, err == nil, err)

		resp, err := stream.Recv()
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "response", resp.Message)
	})

	t.Run("server streaming", func(t *testing.T) {
		ln := serverutils.Listen()
		addr := ln.Addr().String()
		count := 0
		svr := RunSlimThriftServer(&SlimEchoServiceImpl{}, ln, svrCodecOpt,
			server.WithMiddleware(func(e endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, args, result interface{}) (err error) {
					_, ok := args.(*streaming.Args)
					test.Assert(t, ok)
					return e(ctx, args, result)
				}
			}),
			server.WithRecvMiddleware(func(e endpoint.RecvEndpoint) endpoint.RecvEndpoint {
				return func(stream streaming.Stream, req interface{}) (err error) {
					err = e(stream, req)
					realReq, ok := req.(*echo.EchoRequest)
					test.Assert(t, ok)
					test.Assert(t, realReq.Message == "request")
					realReq.Message = "server"
					return
				}
			}),
			server.WithSendMiddleware(func(e endpoint.SendEndpoint) endpoint.SendEndpoint {
				return func(stream streaming.Stream, resp interface{}) (err error) {
					realResp, ok := resp.(*echo.EchoResponse)
					test.Assert(t, ok)
					test.Assert(t, realResp.Message == "server-"+strconv.Itoa(count))
					realResp.Message = "response-" + strconv.Itoa(count)
					count += 1
					return e(stream, realResp)
				}
			}))
		defer svr.Stop()

		cli := echoservice.MustNewStreamClient("service",
			streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
			streamclient.WithHostPorts(addr))
		ctx := metainfo.WithValue(context.Background(), KeyCount, "2")
		stream, err := cli.EchoServer(ctx, &echo.EchoRequest{Message: "request"})
		test.Assert(t, err == nil, err)

		resp, err := stream.Recv()
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "response-0", resp.Message)

		resp, err = stream.Recv()
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "response-1", resp.Message)
	})

	t.Run("client streaming", func(t *testing.T) {
		ln := serverutils.Listen()
		addr := ln.Addr().String()
		count := 0
		svr := RunSlimThriftServer(&SlimEchoServiceImpl{}, ln, svrCodecOpt,
			server.WithMiddleware(func(e endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, args, result interface{}) (err error) {
					_, ok := args.(*streaming.Args)
					test.Assert(t, ok)
					return e(ctx, args, result)
				}
			}),
			server.WithRecvMiddleware(func(e endpoint.RecvEndpoint) endpoint.RecvEndpoint {
				return func(stream streaming.Stream, req interface{}) (err error) {
					err = e(stream, req)
					realReq, ok := req.(*echo.EchoRequest)
					test.Assert(t, ok)
					test.Assert(t, realReq.Message == "request-"+strconv.Itoa(count), realReq.Message)
					count += 1
					realReq.Message += "_recv_middleware"
					return
				}
			}),
			server.WithSendMiddleware(func(e endpoint.SendEndpoint) endpoint.SendEndpoint {
				return func(stream streaming.Stream, resp interface{}) (err error) {
					realResp, ok := resp.(*echo.EchoResponse)
					test.Assert(t, ok)
					test.Assert(t, realResp.Message == "2")
					realResp.Message += "_send_middleware"
					return e(stream, realResp)
				}
			}))
		defer svr.Stop()

		cli := echoservice.MustNewStreamClient("service",
			streamclient.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalReadWrite)),
			streamclient.WithHostPorts(addr))
		ctx := metainfo.WithValue(context.Background(), KeyCount, "2")
		stream, err := cli.EchoClient(ctx)
		test.Assert(t, err == nil, err)

		err = stream.Send(&echo.EchoRequest{Message: "request-0"})
		test.Assert(t, err == nil, err)

		err = stream.Send(&echo.EchoRequest{Message: "request-1"})
		test.Assert(t, err == nil, err)

		resp, err := stream.CloseAndRecv()
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "2_send_middleware", resp.Message)
	})
}
