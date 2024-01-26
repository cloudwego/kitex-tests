// Copyright 2024 CloudWeGo Authors
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
	"io"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex-tests/common"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo/echoservice"
	"github.com/cloudwego/kitex/client/streamclient"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/server"
)

/*
 * Tests
 * - Server 正常结束（client 读到 EOF）
 * - Client send error (after stream closed)
 * - Server return error （client 读到 EOF 以及 return 的err）
 * - Server return BizError（client 读到 EOF 以及 return 的 bizError，code+msg）
 * - Client 侧超时 （client 读到 error）
 * - Server 异常断开（client 读到 transport is closing error）
 */

var (
	_ stats.Tracer                = (*testTracer)(nil)
	_ rpcinfo.StreamEventReporter = (*testTracer)(nil)
)

type testTracer struct {
	recvCount      int
	sendCount      int
	recvSize       uint64
	sendSize       uint64
	finishSendSize uint64
	finishRecvSize uint64
	finishCalled   bool
}

func (t *testTracer) Reset() {
	t.recvCount = 0
	t.sendCount = 0
	t.recvSize = 0
	t.sendSize = 0
	t.finishSendSize = 0
	t.finishRecvSize = 0
	t.finishCalled = false
}

func (t *testTracer) ReportStreamEvent(ctx context.Context, ri rpcinfo.RPCInfo, event rpcinfo.Event) {
	switch event.Event() {
	case stats.StreamSend:
		t.sendCount++
		t.sendSize += ri.Stats().LastSendSize()
		if t.sendSize != ri.Stats().SendSize() {
			klog.Errorf("sendSize = %v, ri.Stats().SendSize() = %v", t.sendSize, ri.Stats().SendSize())
			panic("sendSize != ri.Stats().SendSize()")
		}
	case stats.StreamRecv:
		t.recvCount++
		t.recvSize += ri.Stats().LastRecvSize()
		if t.recvSize != ri.Stats().RecvSize() {
			klog.Errorf("recvSize = %v, ri.Stats().RecvSize() = %v", t.recvSize, ri.Stats().RecvSize())
			panic("recvSize != ri.Stats().RecvSize()")
		}
	default:
		panic(fmt.Sprintf("unknown event: %d", event.Event()))
	}
}

func (t *testTracer) Start(ctx context.Context) context.Context {
	return ctx
}

func (t *testTracer) Finish(ctx context.Context) {
	ri := rpcinfo.GetRPCInfo(ctx)
	t.finishSendSize = ri.Stats().SendSize()
	t.finishRecvSize = ri.Stats().RecvSize()
	t.finishCalled = true
	return
}

func (tr *testTracer) finishCheck(t *testing.T, info string) {
	test.Assert(t, tr.finishCalled, tr)
	test.Assert(t, tr.sendSize == tr.finishSendSize, tr.sendSize, tr.finishSendSize, info)
	test.Assert(t, tr.recvSize == tr.finishRecvSize, tr.recvSize, tr.finishRecvSize, info)
}

var _ streaming.Stream = (*wrapStream)(nil)

type wrapStream struct {
	s streaming.Stream
}

func (w *wrapStream) SetHeader(md metadata.MD) error {
	return w.s.SetHeader(md)
}

func (w *wrapStream) SendHeader(md metadata.MD) error {
	return w.s.SendHeader(md)
}

func (w *wrapStream) SetTrailer(md metadata.MD) {
	w.s.SetTrailer(md)
}

func (w *wrapStream) Header() (metadata.MD, error) {
	return w.s.Header()
}

func (w *wrapStream) Trailer() metadata.MD {
	return w.s.Trailer()
}

func (w *wrapStream) Context() context.Context {
	return w.s.Context()
}

func (w *wrapStream) RecvMsg(m interface{}) error {
	return w.s.RecvMsg(m)
}

func (w *wrapStream) SendMsg(m interface{}) error {
	return w.s.SendMsg(m)
}

func (w *wrapStream) Close() error {
	return w.s.Close()
}

var _ streaming.WithDoFinish = (*wrapStreamWithDoFinish)(nil)

type wrapStreamWithDoFinish struct {
	*wrapStream
}

func (w *wrapStreamWithDoFinish) DoFinish(err error) {
	w.wrapStream.s.(streaming.WithDoFinish).DoFinish(err)
}

func TestTracerNormalEndOfStream(t *testing.T) {
	addr := addrAllocator()

	serverTracer := &testTracer{}
	svr := RunThriftServer(
		&EchoServiceImpl{},
		addr,
		server.WithExitWaitTime(time.Millisecond*10),
		server.WithTracer(serverTracer),
	)
	defer svr.Stop()

	clientTracer := &testTracer{}
	cli := echoservice.MustNewStreamClient("service",
		streamclient.WithHostPorts(addr),
		streamclient.WithTracer(clientTracer),
	)

	t.Run("unary api", func(t *testing.T) {
		serverTracer.Reset()
		clientTracer.Reset()
		_, err := cli.EchoUnary(context.Background(), &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err == nil, err)
		// Client Unary API: same as non-streaming API
		test.Assert(t, clientTracer.recvCount == 0 && clientTracer.sendCount == 0, clientTracer)
		// Server Unary API: same as non-streaming API for Thrift Streaming (skipped recv/send mw)
		test.Assert(t, serverTracer.recvCount == 0 && serverTracer.sendCount == 0, serverTracer)
		// for unary api, recv/send mw is not available
		test.Assert(t, clientTracer.recvSize == 0 && clientTracer.sendSize == 0, clientTracer)
		test.Assert(t, clientTracer.finishSendSize > 0 && clientTracer.finishRecvSize > 0, clientTracer)
		// save for server
		test.Assert(t, serverTracer.recvSize == 0 && serverTracer.sendSize == 0, serverTracer)
		test.Assert(t, serverTracer.finishSendSize > 0 && serverTracer.finishRecvSize > 0, serverTracer)
	})

	t.Run("bidirectional api", func(t *testing.T) {
		serverTracer.Reset()
		clientTracer.Reset()
		count := 3
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))

		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)
		for i := 0; i < count; i++ {
			err = stream.Send(&echo.EchoRequest{Message: "hello"})
			test.Assert(t, err == nil, err)
			_, err = stream.Recv()
			test.Assert(t, err == nil, err)
		}
		// to trigger an io.EOF for invoking the DoFinish
		_, err = stream.Recv()
		test.Assert(t, err == io.EOF, err)
		test.Assert(t, clientTracer.recvCount == count+1, clientTracer) // one extra call for io.EOF
		test.Assert(t, clientTracer.sendCount == count, clientTracer)
		test.Assert(t, serverTracer.recvCount == count, serverTracer)
		test.Assert(t, serverTracer.sendCount == count, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})

	t.Run("bidirectional api", func(t *testing.T) {
		serverTracer.Reset()
		clientTracer.Reset()
		count := 3
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))

		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)
		for i := 0; i < count; i++ {
			err = stream.Send(&echo.EchoRequest{Message: "hello"})
			test.Assert(t, err == nil, err)
			_, err = stream.Recv()
			test.Assert(t, err == nil, err)
		}
		// to trigger an io.EOF for invoking the DoFinish
		_, err = stream.Recv()
		test.Assert(t, err == io.EOF, err)
		test.Assert(t, clientTracer.recvCount == count+1, clientTracer) // one extra call for io.EOF
		test.Assert(t, clientTracer.sendCount == count, clientTracer)
		test.Assert(t, serverTracer.recvCount == count, serverTracer)
		test.Assert(t, serverTracer.sendCount == count, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})

	t.Run("server streaming api", func(t *testing.T) {
		serverTracer.Reset()
		clientTracer.Reset()
		count := 3
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))

		stream, err := cli.EchoServer(ctx, &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err == nil, err)
		for {
			if _, err = stream.Recv(); err == io.EOF {
				break
			}
			test.Assert(t, err == nil, err)
		}
		test.Assert(t, clientTracer.sendCount == 1, clientTracer)
		test.Assert(t, clientTracer.recvCount == count+1, clientTracer) // one extra call for io.EOF
		test.Assert(t, serverTracer.sendCount == count, serverTracer)
		test.Assert(t, serverTracer.recvCount == 1, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})

	t.Run("client streaming api", func(t *testing.T) {
		serverTracer.Reset()
		clientTracer.Reset()
		count := 3
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))

		stream, err := cli.EchoClient(ctx)
		test.Assert(t, err == nil, err)
		for i := 0; i < count; i++ {
			err = stream.Send(&echo.EchoRequest{Message: "hello"})
			test.Assert(t, err == nil, err)
		}
		_, err = stream.CloseAndRecv()
		test.Assert(t, err == nil, err)
		test.Assert(t, clientTracer.sendCount == count, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		test.Assert(t, serverTracer.sendCount == 1, serverTracer)
		test.Assert(t, serverTracer.recvCount == count, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})

	t.Run("client streaming with wrapped stream without DoFinish", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		count := 3
		cli := echoservice.MustNewStreamClient("service",
			streamclient.WithHostPorts(addr),
			streamclient.WithTracer(clientTracer),
			streamclient.WithMiddleware(func(endpoint endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, req, resp interface{}) (err error) {
					err = endpoint(ctx, req, resp)
					streamResult := resp.(*streaming.Result)
					streamResult.Stream = &wrapStream{s: streamResult.Stream}
					return
				}
			}),
		)
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))

		stream, err := cli.EchoClient(ctx)
		test.Assert(t, err == nil, err)
		for i := 0; i < count; i++ {
			err = stream.Send(&echo.EchoRequest{Message: "hello"})
			test.Assert(t, err == nil, err)
		}
		_, err = stream.CloseAndRecv()
		test.Assert(t, err == nil, err)
		test.Assert(t, clientTracer.sendCount == count, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		// regardless of whether wrapped stream implements WithDoFinish, it's done within client.stream.RecvMsg
		test.Assert(t, clientTracer.finishCalled, clientTracer)
	})

	t.Run("client streaming with wrapped stream with DoFinish", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		count := 3
		cli := echoservice.MustNewStreamClient("service",
			streamclient.WithHostPorts(addr),
			streamclient.WithTracer(clientTracer),
			streamclient.WithMiddleware(func(endpoint endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, req, resp interface{}) (err error) {
					err = endpoint(ctx, req, resp)
					streamResult := resp.(*streaming.Result)
					streamResult.Stream = &wrapStreamWithDoFinish{
						wrapStream: &wrapStream{
							s: streamResult.Stream,
						},
					}
					return
				}
			}),
		)
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(count))

		stream, err := cli.EchoClient(ctx)
		test.Assert(t, err == nil, err)
		for i := 0; i < count; i++ {
			err = stream.Send(&echo.EchoRequest{Message: "hello"})
			test.Assert(t, err == nil, err)
		}
		_, err = stream.CloseAndRecv()
		test.Assert(t, err == nil, err)
		test.Assert(t, clientTracer.sendCount == count, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		// wrapped stream doesn't implement WithDoFinish, so the DoFinish won't be called
		test.Assert(t, clientTracer.finishCalled, clientTracer)
	})
}

type thriftTraceHandler struct {
	echo.EchoService
	echoBidirectional func(echo.EchoService_EchoBidirectionalServer) error
	echoClient        func(echo.EchoService_EchoClientServer) error
	echoServer        func(*echo.EchoRequest, echo.EchoService_EchoServerServer) error
	echoUnary         func(context.Context, *echo.EchoRequest) (*echo.EchoResponse, error)
}

func (t *thriftTraceHandler) EchoBidirectional(stream echo.EchoService_EchoBidirectionalServer) (err error) {
	return t.echoBidirectional(stream)
}

func (t *thriftTraceHandler) EchoClient(stream echo.EchoService_EchoClientServer) (err error) {
	return t.echoClient(stream)
}

func (t *thriftTraceHandler) EchoServer(req *echo.EchoRequest, stream echo.EchoService_EchoServerServer) (err error) {
	return t.echoServer(req, stream)
}

func (t *thriftTraceHandler) EchoUnary(ctx context.Context, req1 *echo.EchoRequest) (r *echo.EchoResponse, err error) {
	return t.echoUnary(ctx, req1)
}

func TestTracingSendError(t *testing.T) {
	clientTracer := &testTracer{}
	cli := echoservice.MustNewStreamClient("server",
		streamclient.WithHostPorts(thriftAddr),
		streamclient.WithTracer(clientTracer))

	t.Run("client-send-err", func(t *testing.T) {
		clientTracer.Reset()
		stream, err := cli.EchoClient(context.Background())
		test.Assert(t, err == nil, err)
		stream.Close()
		err = stream.Send(&echo.EchoRequest{Message: "hello"})
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 1, clientTracer)
		test.Assert(t, clientTracer.recvCount == 0, clientTracer)
		clientTracer.finishCheck(t, "client")
	})

	t.Run("bidirectional-send-err", func(t *testing.T) {
		clientTracer.Reset()
		stream, err := cli.EchoBidirectional(context.Background())
		test.Assert(t, err == nil, err)
		stream.Close()
		err = stream.Send(&echo.EchoRequest{Message: "hello"})
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 1, clientTracer)
		test.Assert(t, clientTracer.recvCount == 0, clientTracer)
		clientTracer.finishCheck(t, "client")
	})
}

func TestTracingServerReturnError(t *testing.T) {
	mockError := fmt.Errorf("mock error")
	addr := addrAllocator()
	serverTracer := &testTracer{}
	svr := RunThriftServer(&thriftTraceHandler{
		echoUnary: func(ctx context.Context, request *echo.EchoRequest) (*echo.EchoResponse, error) {
			return nil, mockError
		},
		echoServer: func(request *echo.EchoRequest, serverServer echo.EchoService_EchoServerServer) error {
			return mockError
		},
		echoClient: func(clientServer echo.EchoService_EchoClientServer) error {
			return mockError
		},
		echoBidirectional: func(bidirectionalServer echo.EchoService_EchoBidirectionalServer) error {
			return mockError
		},
	}, addr, server.WithExitWaitTime(time.Millisecond*10), server.WithTracer(serverTracer))
	defer svr.Stop()

	clientTracer := &testTracer{}
	cli := echoservice.MustNewStreamClient("server",
		streamclient.WithHostPorts(addr),
		streamclient.WithTracer(clientTracer))

	t.Run("unary", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		_, err := cli.EchoUnary(context.Background(), &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err != nil, err)
		// recv/send event is not reported to tracer for unary api
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 0, clientTracer)
		test.Assert(t, serverTracer.sendCount == 0, serverTracer)
		test.Assert(t, serverTracer.recvCount == 0, serverTracer)
		test.Assert(t, clientTracer.finishCalled, clientTracer)
		test.Assert(t, serverTracer.finishCalled, serverTracer)
	})

	t.Run("server", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		stream, err := cli.EchoServer(context.Background(), &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err == nil, err)
		_, err = stream.Recv()
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 1, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		test.Assert(t, serverTracer.sendCount == 0, serverTracer)
		test.Assert(t, serverTracer.recvCount == 1, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})

	t.Run("client", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		stream, err := cli.EchoClient(context.Background())
		test.Assert(t, err == nil, err)
		_, err = stream.CloseAndRecv()
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		test.Assert(t, serverTracer.sendCount == 0, serverTracer)
		test.Assert(t, serverTracer.recvCount == 0, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})

	t.Run("bidirectional", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		stream, err := cli.EchoBidirectional(context.Background())
		test.Assert(t, err == nil, err)
		_, err = stream.Recv()
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		test.Assert(t, serverTracer.sendCount == 0, serverTracer)
		test.Assert(t, serverTracer.recvCount == 0, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})
}

func TestTracingServerReturnBizError(t *testing.T) {
	mockError := kerrors.NewGRPCBizStatusError(100, "biz error")
	addr := addrAllocator()
	serverTracer := &testTracer{}
	svr := RunThriftServer(&thriftTraceHandler{
		echoUnary: func(ctx context.Context, request *echo.EchoRequest) (*echo.EchoResponse, error) {
			return nil, mockError
		},
		echoServer: func(request *echo.EchoRequest, serverServer echo.EchoService_EchoServerServer) error {
			return mockError
		},
		echoClient: func(clientServer echo.EchoService_EchoClientServer) error {
			return mockError
		},
		echoBidirectional: func(bidirectionalServer echo.EchoService_EchoBidirectionalServer) error {
			return mockError
		},
	}, addr, server.WithExitWaitTime(time.Millisecond*10), server.WithTracer(serverTracer))
	defer svr.Stop()

	clientTracer := &testTracer{}
	cli := echoservice.MustNewStreamClient("server",
		streamclient.WithHostPorts(addr),
		streamclient.WithTracer(clientTracer))

	t.Run("unary", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		_, err := cli.EchoUnary(context.Background(), &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err.(kerrors.BizStatusErrorIface).BizStatusCode() == 100, err)
		// recv/send event is not reported to tracer for unary api
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 0, clientTracer)
		test.Assert(t, serverTracer.sendCount == 0, serverTracer)
		test.Assert(t, serverTracer.recvCount == 0, serverTracer)
		test.Assert(t, clientTracer.finishCalled, clientTracer)
		test.Assert(t, serverTracer.finishCalled, serverTracer)
	})

	// Waiting for fix for streaming apis with biz error
	t.Run("server", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		stream, err := cli.EchoServer(context.Background(), &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err == nil, err)
		_, err = stream.Recv()
		bizErr, ok := kerrors.FromBizStatusError(err)
		test.Assert(t, ok && bizErr.BizStatusCode() == 100, err)
		test.Assert(t, clientTracer.sendCount == 1, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		test.Assert(t, serverTracer.sendCount == 0, serverTracer)
		test.Assert(t, serverTracer.recvCount == 1, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})

	t.Run("client", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		stream, err := cli.EchoClient(context.Background())
		test.Assert(t, err == nil, err)
		_, err = stream.CloseAndRecv()
		bizErr, ok := kerrors.FromBizStatusError(err)
		test.Assert(t, ok && bizErr.BizStatusCode() == 100, err)
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		test.Assert(t, serverTracer.sendCount == 0, serverTracer)
		test.Assert(t, serverTracer.recvCount == 0, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})

	t.Run("bidirectional", func(t *testing.T) {
		clientTracer.Reset()
		serverTracer.Reset()
		stream, err := cli.EchoBidirectional(context.Background())
		test.Assert(t, err == nil, err)
		_, err = stream.Recv()
		bizErr, ok := kerrors.FromBizStatusError(err)
		test.Assert(t, ok && bizErr.BizStatusCode() == 100, err)
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		test.Assert(t, serverTracer.sendCount == 0, serverTracer)
		test.Assert(t, serverTracer.recvCount == 0, serverTracer)
		clientTracer.finishCheck(t, "client")
		serverTracer.finishCheck(t, "server")
	})
}

func TestTracingClientTimeout(t *testing.T) {
	addr := addrAllocator()
	svr := RunThriftServer(&thriftTraceHandler{
		echoUnary: func(ctx context.Context, request *echo.EchoRequest) (*echo.EchoResponse, error) {
			time.Sleep(time.Millisecond * 100)
			return &echo.EchoResponse{Message: request.Message}, nil
		},
		echoServer: func(request *echo.EchoRequest, serverServer echo.EchoService_EchoServerServer) error {
			time.Sleep(time.Millisecond * 100)
			return nil
		},
		echoClient: func(clientServer echo.EchoService_EchoClientServer) error {
			time.Sleep(time.Millisecond * 100)
			return nil
		},
		echoBidirectional: func(bidirectionalServer echo.EchoService_EchoBidirectionalServer) error {
			time.Sleep(time.Millisecond * 100)
			return nil
		},
	}, addr, server.WithExitWaitTime(time.Millisecond*10))
	defer svr.Stop()

	clientTracer := &testTracer{}
	cli := echoservice.MustNewStreamClient("server",
		streamclient.WithHostPorts(addr),
		streamclient.WithTracer(clientTracer))

	t.Run("unary", func(t *testing.T) {
		clientTracer.Reset()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*30)
		defer cancel()
		_, err := cli.EchoUnary(ctx, &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err != nil, err)
		// recv/send event is not reported to tracer for unary api
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 0, clientTracer)
		test.Assert(t, clientTracer.finishCalled, clientTracer)
	})

	t.Run("server", func(t *testing.T) {
		clientTracer.Reset()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*30)
		defer cancel()
		stream, err := cli.EchoServer(ctx, &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err == nil, err)
		_, err = stream.Recv()
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 1, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		clientTracer.finishCheck(t, "client")
	})

	t.Run("client", func(t *testing.T) {
		clientTracer.Reset()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*30)
		defer cancel()
		stream, err := cli.EchoClient(ctx)
		test.Assert(t, err == nil, err)
		_, err = stream.CloseAndRecv()
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		clientTracer.finishCheck(t, "client")
	})

	t.Run("bidirectional", func(t *testing.T) {
		clientTracer.Reset()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*30)
		defer cancel()
		stream, err := cli.EchoBidirectional(ctx)
		test.Assert(t, err == nil, err)
		_, err = stream.Recv()
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		clientTracer.finishCheck(t, "client")
	})
}

func TestTracingServerStop(t *testing.T) {
	addr := "127.0.0.1:9999"
	clientTracer := &testTracer{}
	cli := echoservice.MustNewStreamClient("server",
		streamclient.WithHostPorts(addr),
		streamclient.WithTracer(clientTracer))

	runServer := func() *exec.Cmd {
		cmd := exec.Command("binaries/exitserver", "-addr", addr)
		err := cmd.Start()
		test.Assert(t, err == nil, err)
		common.WaitServer(addr)
		return cmd
	}

	t.Run("unary", func(t *testing.T) {
		cmd := runServer()
		defer cmd.Wait()
		clientTracer.Reset()
		_, err := cli.EchoUnary(context.Background(), &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err != nil, err)
		// recv/send event is not reported to tracer for unary api
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 0, clientTracer)
		test.Assert(t, clientTracer.finishCalled, clientTracer)
	})

	t.Run("server", func(t *testing.T) {
		cmd := runServer()
		defer cmd.Wait()

		clientTracer.Reset()
		stream, err := cli.EchoServer(context.Background(), &echo.EchoRequest{Message: "hello"})
		test.Assert(t, err == nil, err)
		_, err = stream.Recv()
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 1, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		clientTracer.finishCheck(t, "client")
	})

	t.Run("client", func(t *testing.T) {
		cmd := runServer()
		defer cmd.Wait()

		clientTracer.Reset()
		stream, err := cli.EchoClient(context.Background())
		test.Assert(t, err == nil, err)
		_ = stream.Close()
		_, err = stream.CloseAndRecv()
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		clientTracer.finishCheck(t, "client")
	})

	t.Run("bidirectional", func(t *testing.T) {
		cmd := runServer()
		defer cmd.Wait()

		clientTracer.Reset()
		stream, err := cli.EchoBidirectional(context.Background())
		test.Assert(t, err == nil, err)
		_ = stream.Close()
		_, err = stream.Recv()
		test.Assert(t, err != nil, err)
		test.Assert(t, clientTracer.sendCount == 0, clientTracer)
		test.Assert(t, clientTracer.recvCount == 1, clientTracer)
		clientTracer.finishCheck(t, "client")
	})
}
