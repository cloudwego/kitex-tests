// Copyright 2025 CloudWeGo Authors
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

package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/echoservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
)

const (
	testMessage = "hello world"
	bizErrMsg   = "biz error"
	bizErrCode  = 502
	testTimeout = 5 * time.Second
)

var (
	proxyAddr  net.Addr
	proxySvr   server.Server
	backendSvr server.Server
)

func TestMain(m *testing.M) {
	// 启动后端服务
	backendLn := serverutils.Listen()
	backendSvr = newBackendServer(backendLn)

	// 启动代理服务
	proxyLn := serverutils.Listen()
	proxyAddr = proxyLn.Addr()
	proxySvr = newGenericProxyServer(proxyLn, backendLn.Addr())

	m.Run()

	// 清理资源
	_ = backendSvr.Stop()
	_ = proxySvr.Stop()
}

func newBackendServer(ln net.Listener) server.Server {
	handler := &backendHandler{}

	svr := echoservice.NewServer(handler,
		server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler))

	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()

	// 等待服务启动
	time.Sleep(100 * time.Millisecond)
	return svr
}

type backendHandler struct {
}

func (h *backendHandler) Echo(_ context.Context, req *tenant.EchoRequest) (r *tenant.EchoResponse, err error) {
	// 测试 biz status error
	if req.Msg == bizErrMsg {
		return nil, kerrors.NewBizStatusErrorWithExtra(bizErrCode, bizErrMsg, map[string]string{"detail": "test biz error"})
	}

	if req.Msg != testMessage {
		return nil, errors.New("message not match")
	}

	return &tenant.EchoResponse{Msg: testMessage}, nil
}

func (h *backendHandler) EchoOneway(_ context.Context, _ *tenant.EchoRequest) (err error) {
	return nil
}

func (h *backendHandler) EchoBidi(ctx context.Context, stream tenant.EchoService_EchoBidiServer) (err error) {
	for {
		req, err := stream.Recv(ctx)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if req.Msg == bizErrMsg {
			return kerrors.NewBizStatusErrorWithExtra(bizErrCode, bizErrMsg, map[string]string{"detail": "test biz error"})
		}
		err = stream.Send(ctx, &tenant.EchoResponse{Msg: req.Msg})
		if err != nil {
			return err
		}
	}
}

func (h *backendHandler) EchoClient(ctx context.Context, stream tenant.EchoService_EchoClientServer) (err error) {
	for {
		req, err := stream.Recv(ctx)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if req.Msg == bizErrMsg {
			return kerrors.NewBizStatusErrorWithExtra(bizErrCode, bizErrMsg, map[string]string{"detail": "test biz error"})
		}
		err = stream.SendAndClose(ctx, &tenant.EchoResponse{Msg: req.Msg})
		if err != nil {
			return err
		}
	}
}

func (h *backendHandler) EchoServer(ctx context.Context, req *tenant.EchoRequest, stream tenant.EchoService_EchoServerServer) (err error) {
	for i := 0; i < 3; i++ {
		if req.Msg == bizErrMsg {
			return kerrors.NewBizStatusErrorWithExtra(bizErrCode, bizErrMsg, map[string]string{"detail": "test biz error"})
		}
		err = stream.Send(ctx, &tenant.EchoResponse{Msg: fmt.Sprintf("%s_%d", testMessage, i)})
		if err != nil {
			return err
		}
	}
	return nil
}

func newEchoClient(targetAddr net.Addr, opts ...client.Option) echoservice.Client {
	defaultOpts := []client.Option{
		client.WithHostPorts(targetAddr.String()),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithMetaHandler(transmeta.ClientHTTP2Handler),
		client.WithTransportProtocol(transport.TTHeader | transport.TTHeaderStreaming),
	}
	opts = append(defaultOpts, opts...)

	cli, err := echoservice.NewClient("echoService", opts...)
	if err != nil {
		panic(err)
	}
	return cli
}

func TestProxyGeneratedCall(t *testing.T) {
	cli := newEchoClient(proxyAddr)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	resp, err := cli.Echo(ctx, &tenant.EchoRequest{Msg: testMessage})
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Msg == testMessage)
}

func TestProxyBizStatusError(t *testing.T) {
	cli := newEchoClient(proxyAddr)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	_, err := cli.Echo(ctx, &tenant.EchoRequest{Msg: bizErrMsg})
	test.Assert(t, err != nil)

	// 验证是 biz status error
	bizErr, ok := kerrors.FromBizStatusError(err)
	test.Assert(t, ok, "should be biz status error")
	test.Assert(t, bizErr.BizStatusCode() == bizErrCode)
	test.Assert(t, bizErr.BizMessage() == bizErrMsg)
}

func TestProxyStreamingBidi(t *testing.T) {
	cli := newEchoClient(proxyAddr)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	stream, err := cli.EchoBidi(ctx)
	test.Assert(t, err == nil, err)

	// 发送消息
	err = stream.Send(stream.Context(), &tenant.EchoRequest{Msg: testMessage})
	test.Assert(t, err == nil, err)

	// 接收响应
	resp, err := stream.Recv(stream.Context())
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Msg == testMessage)

	// 关闭发送
	err = stream.CloseSend(stream.Context())
	test.Assert(t, err == nil, err)

	// 等待接收结束
	_, err = stream.Recv(stream.Context())
	test.Assert(t, err == io.EOF, err)
}

func TestProxyStreamingClient(t *testing.T) {
	cli := newEchoClient(proxyAddr)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	stream, err := cli.EchoClient(ctx)
	test.Assert(t, err == nil, err)

	// 发送消息
	err = stream.Send(stream.Context(), &tenant.EchoRequest{Msg: testMessage})
	test.Assert(t, err == nil, err)

	// 关闭发送并接收响应
	resp, err := stream.CloseAndRecv(stream.Context())
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Msg == testMessage)
}

func TestProxyStreamingServer(t *testing.T) {
	cli := newEchoClient(proxyAddr)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	stream, err := cli.EchoServer(ctx, &tenant.EchoRequest{Msg: testMessage})
	test.Assert(t, err == nil, err)

	// 接收多个响应
	for i := 0; i < 3; i++ {
		resp, err := stream.Recv(stream.Context())
		test.Assert(t, err == nil, err)
		expectedMsg := fmt.Sprintf("%s_%d", testMessage, i)
		test.Assert(t, resp.Msg == expectedMsg)
	}

	// 等待接收结束
	_, err = stream.Recv(stream.Context())
	test.Assert(t, err == io.EOF, err)
}

func TestProxyStreamingBizStatusError(t *testing.T) {
	cli := newEchoClient(proxyAddr)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	stream, err := cli.EchoBidi(ctx)
	test.Assert(t, err == nil, err)

	// 发送错误消息
	err = stream.Send(stream.Context(), &tenant.EchoRequest{Msg: bizErrMsg})
	test.Assert(t, err == nil, err)

	// 等待 biz status error
	_, err = stream.Recv(stream.Context())
	test.Assert(t, err != nil)

	// 验证是 biz status error
	bizErr, ok := kerrors.FromBizStatusError(err)
	test.Assert(t, ok, "should be biz status error")
	test.Assert(t, bizErr.BizStatusCode() == bizErrCode)
	test.Assert(t, bizErr.BizMessage() == bizErrMsg)
}
