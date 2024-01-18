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
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex-tests/common"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/a/b/c"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/combine"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/combine/combineservice"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo/abcservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/streamclient"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/logid"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo/echoservice"
)

func TestKitexThriftBidirectional(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithHostPorts(thriftAddr))
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

func TestKitexThriftClient(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithHostPorts(thriftAddr))
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

func TestKitexThriftServer(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithHostPorts(thriftAddr),
	)

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

func TestKitexThriftEchoUnary(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithHostPorts(thriftAddr))

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

func TestKitexThriftEchoPingPong(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr),
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

func TestKitexThriftEchoOneway(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr),
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

func TestKitexThriftPing(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr),
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

func TestKitexClientMiddleware(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr),
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

func TestKitexStreamClientMiddlewareUnary(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithHostPorts(thriftAddr),
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

func TestKitexStreamClientMiddlewareBidirectional(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithHostPorts(thriftAddr),
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

func TestKitexStreamClientMiddlewareClient(t *testing.T) {
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithHostPorts(thriftAddr),
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

func TestKitexStreamClientMiddlewareServer(t *testing.T) {
	recvCount := 0
	cli := echoservice.MustNewStreamClient(
		"service",
		streamclient.WithHostPorts(thriftAddr),
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

func TestKitexServerMiddleware(t *testing.T) {
	addr := addrAllocator()
	t.Run("pingpong", func(t *testing.T) {
		svr := RunThriftServer(&EchoServiceImpl{}, addr,
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
			client.WithHostPorts(addr),
			client.WithTransportProtocol(transport.TTHeaderFramed),
		)
		req1, req2 := &echo.EchoRequest{Message: "ping"}, &echo.EchoRequest{Message: "pong"}
		resp, err := cli.EchoPingPong(context.Background(), req1, req2)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "response", resp.Message)
	})

	t.Run("unary", func(t *testing.T) {
		svr := RunThriftServer(&EchoServiceImpl{}, addr,
			server.WithMiddleware(func(e endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, args, result interface{}) (err error) {
					realArgs, ok := args.(*echo.EchoServiceEchoUnaryArgs)
					if !ok {
						return fmt.Errorf("invalid args type: not %T", args)
					}
					if realArgs.Req1.Message != "unary" {
						return fmt.Errorf("invalid realArgs.Req1.Message: %v", realArgs.Req1.Message)
					}
					if err = e(ctx, args, result); err != nil {
						return err
					}
					resp, ok := result.(*echo.EchoServiceEchoUnaryResult)
					if !ok {
						return fmt.Errorf("invalid result type: not %T", result)
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

		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		req := &echo.EchoRequest{Message: "unary"}
		resp, err := cli.EchoUnary(context.Background(), req)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "response", resp.Message)
	})

	t.Run("bidirectional streaming", func(t *testing.T) {
		svr := RunThriftServer(&EchoServiceImpl{}, addr,
			server.WithMiddleware(func(e endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, args, result interface{}) (err error) {
					_, ok := args.(*streaming.Args)
					test.Assert(t, ok)
					return e(ctx, args, result)
				}
			}),
			server.WithRecvMiddleware(func(e endpoint.RecvEndpoint) endpoint.RecvEndpoint {
				return func(stream streaming.Stream, req interface{}) (err error) {
					if err = e(stream, req); err != nil {
						return
					}
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

		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
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
		count := 0
		svr := RunThriftServer(&EchoServiceImpl{}, addr,
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

		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
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
		count := 0
		svr := RunThriftServer(&EchoServiceImpl{}, addr,
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

		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
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

func TestTimeoutRecvSend(t *testing.T) {
	t.Run("client send timeout", func(t *testing.T) {
		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(thriftAddr),
			streamclient.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
				return func(stream streaming.Stream, message interface{}) (err error) {
					time.Sleep(time.Millisecond * 100)
					return next(stream, message)
				}
			}),
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx = metainfo.WithValue(ctx, KeyCount, "1")
		ctx = AddKeyValueHTTP2(ctx, KeyServerRecvTimeoutMS, "100")
		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)

		err = streaming.CallWithTimeout(time.Millisecond*50, cancel, func() (err error) {
			err = stream.Send(&echo.EchoRequest{Message: "request"})
			return
		})
		test.Assertf(t, err != nil, "err = %v", err)
		test.Assertf(t, strings.Contains(err.Error(), "timeout in business code"), "err = %v", err)
	})

	t.Run("client recv timeout", func(t *testing.T) {
		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(thriftAddr),
			streamclient.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
				return func(stream streaming.Stream, message interface{}) (err error) {
					time.Sleep(time.Millisecond * 100)
					return next(stream, message)
				}
			}),
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx = metainfo.WithValue(ctx, KeyCount, "1")
		ctx = AddKeyValueHTTP2(ctx, KeyServerSendTimeoutMS, "100")
		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)

		err = stream.Send(&echo.EchoRequest{Message: "request"})
		test.Assertf(t, err == nil, "err = %v", err)

		var resp *echo.EchoResponse
		err = streaming.CallWithTimeout(time.Millisecond*50, cancel, func() (err error) {
			resp, err = stream.Recv()
			return
		})
		test.Assertf(t, err != nil, "err = %v", err)
		test.Assertf(t, strings.Contains(err.Error(), "timeout in business code"), "err = %v", err)
		test.Assertf(t, resp == nil, "resp = %v", resp)
	})

	t.Run("client recv timeout with MetaHandler", func(t *testing.T) {
		addr := addrAllocator()
		svr := RunThriftServer(&thriftServerTimeoutImpl{
			Handler: func(stream echo.EchoService_EchoBidirectionalServer) (err error) {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
		}, addr)
		defer svr.Stop()

		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr),
			WithClientContextCancel(),
		)
		stream, err := cli.EchoBidirectional(context.Background())
		test.Assert(t, err == nil, err)
		err = CallWithCtx(stream.Context(), time.Millisecond*50, func() (err error) {
			_, err = stream.Recv()
			klog.Infof("client recv: err = %v", err)
			test.Assert(t, err != nil, err)
			return err
		})
		test.Assert(t, err != nil, err)
		test.Assert(t, strings.Contains(err.Error(), "timeout in business code"), err.Error())
	})

	t.Run("server recv timeout with MetaHandler", func(t *testing.T) {
		addr := addrAllocator()
		svr := RunThriftServer(
			&thriftServerTimeoutImpl{
				Handler: func(stream echo.EchoService_EchoBidirectionalServer) (err error) {
					err = CallWithCtx(stream.Context(), time.Millisecond*50, func() (err error) {
						_, err = stream.Recv()
						klog.Infof("server recv: err = %v", err)
						test.Assert(t, err != nil, err)
						return
					})
					klog.Infof("server recv: err = %v", err)
					return err
				},
			},
			addr,
			WithServerContextCancel(),
		)
		defer svr.Stop()

		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		stream, err := cli.EchoBidirectional(context.Background())
		test.Assert(t, err == nil, err)
		_, err = stream.Recv()
		test.Assertf(t, err != nil, "err = %v", err)
	})
}

type thriftServerTimeoutImpl struct {
	echo.EchoService
	Handler func(stream echo.EchoService_EchoBidirectionalServer) (err error)
}

func (h *thriftServerTimeoutImpl) EchoBidirectional(stream echo.EchoService_EchoBidirectionalServer) (err error) {
	return h.Handler(stream)
}

func TestThriftStreamingMetaData(t *testing.T) {
	t.Run("metainfo", func(t *testing.T) {
		metainfoKeys := map[string]bool{
			"HELLO":       true,
			"HELLO_WORLD": true,
			// lower case keys will be ignored
			"Foo":         false,
			"Foo_Bar":     false,
			"hello":       false,
			"hello_world": false,
		}

		metadataKeys := []string{
			"hello", "WORLD", "foo_bar", "BAR_FOO",
			"HELLO", // case-insensitive, will be merged into "hello"
		}

		headers := metadata.MD{
			"hello":   []string{"3", "4"},
			"foo_bar": []string{"5", "6"},
			"bar-foo": []string{"7", "8"},
			"x1":      []string{"9"},
			"x_1":     []string{"10"},
			"1_x":     []string{"11"},
			// capital letters not allowed, will cause error at client side
		}
		trailers := metadata.MD{
			"xxx":     []string{"9"},
			"yyy_zzz": []string{"10"},
			"zzz-yyy": []string{"11"},
			"x1":      []string{"12"},
			"x_1":     []string{"13"},
			"1_x":     []string{"14"},
		}

		addr := addrAllocator()

		svr := RunThriftServer(
			&thriftServerTimeoutImpl{
				Handler: func(stream echo.EchoService_EchoBidirectionalServer) (err error) {
					values := metainfo.GetAllValues(stream.Context())
					klog.Infof("metainfo.GetAllValues: v = %v", values)
					for key, shouldExists := range metainfoKeys {
						value, exists := values[key]
						test.Assertf(t, exists == shouldExists, "key = %v, shouldExists = %v, exists = %v",
							key, shouldExists, exists)
						if shouldExists {
							test.Assertf(t, value == "1", "key = %v, v = %v, exists = %v", key, value, exists)
						}
					}

					values = metainfo.GetAllPersistentValues(stream.Context())
					klog.Infof("metainfo.GetAllPersistentValues: v = %v", values)
					for key, shouldExists := range metainfoKeys {
						value, exists := metainfo.GetPersistentValue(stream.Context(), key)
						test.Assertf(t, exists == shouldExists, "key = %v, shouldExists = %v, exists = %v",
							key, shouldExists, exists)
						if shouldExists {
							test.Assertf(t, value == "1", "v = %v, exists = %v", value, exists)
						}
					}

					// TODO: client unable to receive backward values
					metainfo.SendBackwardValue(stream.Context(), "backward", "1")
					metainfo.SendBackwardValue(stream.Context(), "BACKWARD", "1")

					md, ok := metadata.FromIncomingContext(stream.Context())
					klog.Infof("metadata: md = %v, ok = %v", md, ok)
					for _, key := range metadataKeys {
						value := md.Get(key)
						if strings.ToLower(key) == "hello" { // case-insensitive, same keys merged
							test.Assert(t, len(value) == 4, "value = %v", value)
							test.Assert(t, reflect.DeepEqual(value, []string{"1", "2", "1", "2"}), value)
						} else {
							test.Assert(t, len(value) == 2, "value = %v", value)
							test.Assert(t, reflect.DeepEqual(value, []string{"1", "2"}), value)
						}
					}

					err = stream.SendHeader(headers)
					test.Assert(t, err == nil, err)

					err = stream.Send(&echo.EchoResponse{Message: "ok"})
					test.Assert(t, err == nil, err)

					stream.SetTrailer(trailers)
					return nil
				},
			},
			addr,
		)
		defer svr.Stop()
		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))

		ctx := metainfo.WithBackwardValues(context.Background())

		for k := range metainfoKeys {
			ctx = metainfo.WithValue(ctx, k, "1")
			ctx = metainfo.WithPersistentValue(ctx, k, "1")
		}

		for _, k := range metadataKeys {
			ctx = metadata.AppendToOutgoingContext(ctx, k, "1", k, "2")
		}

		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)

		err = stream.Send(&echo.EchoRequest{Message: "hello"})
		test.Assert(t, err == nil, err)

		headerReceived, err := stream.Header()
		test.Assertf(t, err == nil, "err = %v", err)
		klog.Infof("headerReceived: %v", headerReceived)
		for k, expected := range headers {
			got := headerReceived[k]
			test.Assertf(t, reflect.DeepEqual(got, expected), "key = %v, got = %v, expected = %v", k, got, expected)
		}

		trailerReceived := stream.Trailer()
		klog.Infof("trailerReceived: %v", trailerReceived)
		for k, expected := range trailers {
			got := trailerReceived[k]
			test.Assertf(t, reflect.DeepEqual(got, expected), "key = %v, got = %v, expected = %v", k, got, expected)
		}

		// TODO: not implemented?
		klog.Infof("metainfo.BackwardValues: %v", metainfo.RecvAllBackwardValues(stream.Context()))
	})

	t.Run("header-failure", func(t *testing.T) {
		addr := addrAllocator()
		svr := RunThriftServer(
			&thriftServerTimeoutImpl{
				Handler: func(stream echo.EchoService_EchoBidirectionalServer) (err error) {
					_, err = stream.Recv()
					test.Assert(t, err == nil, err)

					err = stream.SendHeader(metadata.MD{
						"HELLO": []string{"1"}, // upper case keys cause error at client side
					})
					test.Assert(t, err == nil, err)

					return stream.Send(&echo.EchoResponse{Message: "ok"})
				},
			},
			addr,
		)
		defer svr.Stop()

		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		ctx := metainfo.WithBackwardValues(context.Background())
		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)

		err = stream.Send(&echo.EchoRequest{Message: "hello"})
		test.Assert(t, err == nil, err)

		receivedHeaders, err := stream.Header()
		klog.Infof("receivedHeaders: %v, err = %v", receivedHeaders, err)
		test.Assert(t, err != nil, err)
		test.Assert(t, strings.Contains(err.Error(), `invalid header field name "HELLO"`), err)
	})

	t.Run("trailer-failure", func(t *testing.T) {
		addr := addrAllocator()
		svr := RunThriftServer(
			&thriftServerTimeoutImpl{
				Handler: func(stream echo.EchoService_EchoBidirectionalServer) (err error) {
					_, err = stream.Recv()
					test.Assert(t, err == nil, err)

					err = stream.Send(&echo.EchoResponse{Message: "ok"})
					test.Assert(t, err == nil, err)

					stream.SetTrailer(metadata.MD{
						"HELLO": []string{"1"}, // upper case keys cause error at client side
					})
					return nil
				},
			},
			addr,
		)
		defer svr.Stop()

		cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		ctx := metainfo.WithBackwardValues(context.Background())
		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)

		err = stream.Send(&echo.EchoRequest{Message: "hello"})
		test.Assert(t, err == nil, err)

		receivedTrailers := stream.Trailer()
		klog.Infof("receivedTrailers: %v, err = %v", receivedTrailers, err)
		test.Assertf(t, len(receivedTrailers) == 0, "receivedTrailers = %v", receivedTrailers)
	})
}

func TestCombineService(t *testing.T) {
	t.Run("combine-foo", func(t *testing.T) {
		cli := combineservice.MustNewClient("combineservice", client.WithHostPorts(combineAddr))
		rsp, err := cli.Foo(context.Background(), &combine.Req{Message: "hello"})
		test.Assert(t, err == nil, err)
		test.Assert(t, rsp.Message == "hello", rsp.Message)
	})
	t.Run("combine-bar", func(t *testing.T) {
		streamCli := combineservice.MustNewStreamClient("combineservice", streamclient.WithHostPorts(combineAddr))
		stream, err := streamCli.Bar(context.Background())
		test.Assert(t, err == nil, err)
		err = stream.Send(&combine.Req{Message: "hello"})
		test.Assert(t, err == nil, err)
		rsp, err := stream.Recv()
		test.Assert(t, err == nil, err)
		test.Assert(t, rsp.Message == "hello", rsp.Message)
		err = stream.Close()
		test.Assert(t, err == nil, err)
	})
}

func TestThriftStreamLogID(t *testing.T) {
	defer logid.SetLogIDGenerator(logid.DefaultLogIDGenerator)

	expectedLogID := ""
	checkLogID := func(ctx context.Context) (got string, match bool) {
		got = logid.GetStreamLogID(ctx)
		return got, got == expectedLogID
	}

	addr := addrAllocator()
	svr := RunThriftServer(
		&thriftServerTimeoutImpl{
			Handler: func(stream echo.EchoService_EchoBidirectionalServer) (err error) {
				if got, ok := checkLogID(stream.Context()); !ok {
					return fmt.Errorf("EchoBidirectional got logID = %v", got)
				}
				if _, err = stream.Recv(); err != nil {
					return err
				}
				return stream.Send(&echo.EchoResponse{Message: "ok"})
			},
		},
		addr,
		server.WithExitWaitTime(time.Millisecond*10),
		server.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				if got, ok := checkLogID(ctx); !ok {
					return fmt.Errorf("svr mw got logID = %v", got)
				}
				return next(ctx, req, resp)
			}
		}),
		server.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
			return func(stream streaming.Stream, message interface{}) (err error) {
				if got, ok := checkLogID(stream.Context()); !ok {
					return fmt.Errorf("svr recv mw got logID = %v", got)
				}
				return next(stream, message)
			}
		}),
		server.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
			return func(stream streaming.Stream, message interface{}) (err error) {
				if got, ok := checkLogID(stream.Context()); !ok {
					return fmt.Errorf("svr send mw got logID = %v", got)
				}
				return next(stream, message)
			}
		}),
	)
	defer svr.Stop()

	cli := echoservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr),
		streamclient.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				// streamLogID is injected in user defined middleware, and is available for inner middlewares
				ctx = logid.NewCtxWithStreamLogID(ctx, logid.GenerateStreamLogID(ctx))
				expectedLogID = logid.GetStreamLogID(ctx)
				klog.Infof("expectedLogID = %v", expectedLogID)
				test.Assertf(t, expectedLogID != "", "expectedLogID = %v", expectedLogID)
				return next(ctx, req, resp)
			}
		}),
		streamclient.WithRecvMiddleware(func(next endpoint.RecvEndpoint) endpoint.RecvEndpoint {
			return func(stream streaming.Stream, resp interface{}) (err error) {
				if got, ok := checkLogID(stream.Context()); !ok {
					return fmt.Errorf("cli recv mw got logID = %v", got)
				}
				return next(stream, resp)
			}
		}),
		streamclient.WithSendMiddleware(func(next endpoint.SendEndpoint) endpoint.SendEndpoint {
			return func(stream streaming.Stream, req interface{}) (err error) {
				if got, ok := checkLogID(stream.Context()); !ok {
					return fmt.Errorf("cli send mw got logID = %v", got)
				}
				return next(stream, req)
			}
		}),
	)
	stream, err := cli.EchoBidirectional(context.Background())
	test.Assert(t, err == nil, err)

	err = stream.Send(&echo.EchoRequest{Message: "hello"})
	test.Assert(t, err == nil, err)

	_, err = stream.Recv()
	test.Assert(t, err == nil, err)

	if got, ok := checkLogID(stream.Context()); !ok {
		t.Errorf("stream ctx got logID = %v", got)
	}
}

func RunABCServer(handler echo.ABCService, addr string, opts ...server.Option) server.Server {
	opts = append(opts, WithServerAddr(addr))
	opts = append(opts, server.WithExitWaitTime(time.Millisecond*10))
	svr := abcservice.NewServer(handler, opts...)
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	common.WaitServer(addr)
	return svr
}

type abcServerImpl struct {
}

func (a abcServerImpl) Echo(ctx context.Context, req1 *c.Request, req2 *c.Request) (r *c.Response, err error) {
	r = &c.Response{
		Message: req1.Message + req2.Message,
	}
	return
}

func (a abcServerImpl) EchoBidirectional(stream echo.ABCService_EchoBidirectionalServer) (err error) {
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	return stream.Send(&c.Response{Message: req.Message})
}

func (a abcServerImpl) EchoServer(req *c.Request, stream echo.ABCService_EchoServerServer) (err error) {
	return stream.Send(&c.Response{Message: req.Message})
}

func (a abcServerImpl) EchoClient(stream echo.ABCService_EchoClientServer) (err error) {
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	return stream.SendAndClose(&c.Response{Message: req.Message})
}

func (a abcServerImpl) EchoUnary(ctx context.Context, req1 *c.Request) (r *c.Response, err error) {
	r = &c.Response{
		Message: req1.Message,
	}
	return
}

func TestABCService(t *testing.T) {
	addr := addrAllocator()
	svr := RunABCServer(&abcServerImpl{}, addr)
	defer svr.Stop()
	t.Run("echo", func(t *testing.T) {
		cli := abcservice.MustNewClient("service", client.WithHostPorts(addr))
		rsp, err := cli.Echo(context.Background(), &c.Request{Message: "hello"}, &c.Request{Message: "world"})
		test.Assert(t, err == nil, err)
		test.Assert(t, rsp.Message == "helloworld")
	})
	t.Run("echo-unary", func(t *testing.T) {
		cli := abcservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		rsp, err := cli.EchoUnary(context.Background(), &c.Request{Message: "hello"})
		test.Assert(t, err == nil, err)
		test.Assert(t, rsp.Message == "hello")
	})
	t.Run("echo-bidirectional", func(t *testing.T) {
		cli := abcservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		stream, err := cli.EchoBidirectional(context.Background())
		test.Assert(t, err == nil, err)
		err = stream.Send(&c.Request{Message: "hello"})
		test.Assert(t, err == nil, err)
		rsp, err := stream.Recv()
		test.Assert(t, err == nil, err)
		test.Assert(t, rsp.Message == "hello")
	})
	t.Run("echo-server", func(t *testing.T) {
		cli := abcservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		stream, err := cli.EchoServer(context.Background(), &c.Request{Message: "hello"})
		test.Assert(t, err == nil, err)
		rsp, err := stream.Recv()
		test.Assert(t, err == nil, err)
		test.Assert(t, rsp.Message == "hello")
	})
	t.Run("echo-client", func(t *testing.T) {
		cli := abcservice.MustNewStreamClient("service", streamclient.WithHostPorts(addr))
		stream, err := cli.EchoClient(context.Background())
		test.Assert(t, err == nil, err)
		err = stream.Send(&c.Request{Message: "hello"})
		test.Assert(t, err == nil, err)
		rsp, err := stream.CloseAndRecv()
		test.Assert(t, err == nil, err)
		test.Assert(t, rsp.Message == "hello")
	})
}
