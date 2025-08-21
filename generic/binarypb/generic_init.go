/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/server/genericserver"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi/mock"
)

const (
	serviceName = "Mock"
	packageName = "pbapi"
)

func newGenericClient(g generic.Generic, targetIPPort string, cliOpts ...client.Option) genericclient.Client {
	cliOpts = append(cliOpts, client.WithHostPorts(targetIPPort),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithMetaHandler(transmeta.ClientHTTP2Handler))
	cli, err := genericclient.NewClient("destService", g, cliOpts...)
	if err != nil {
		panic(err)
	}
	return cli
}

func newGenericServer(handler *genericserver.UnknownServiceOrMethodHandler, ln net.Listener, opts ...server.Option) server.Server {
	opts = append(opts, server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler))
	svr := genericserver.NewUnknownServiceOrMethodServer(handler, opts...)
	go func() {
		err := svr.Run()
		if err != nil {
			panic(err)
		}
	}()
	// wait for server starting to avoid data race
	time.Sleep(100 * time.Millisecond)
	return svr
}

func pingPongUnknownHandler(ctx context.Context, service, method string, request interface{}) (response interface{}, err error) {
	ri := rpcinfo.GetRPCInfo(ctx)
	if service != serviceName && ri.Config().TransportProtocol() != transport.Framed {
		return nil, fmt.Errorf("service not match")
	}
	if method != "UnaryTest" {
		return nil, fmt.Errorf("method not match")
	}
	args := &mock.UnaryTestArgs{}
	err = args.Unmarshal(request.([]byte))
	if err != nil {
		return nil, err
	}
	if args.Req.Message != "hello world" {
		return nil, fmt.Errorf("message not match")
	}
	res := &mock.UnaryTestResult{
		Success: &pbapi.MockResp{Message: "hello world"},
	}
	buf, _ := res.Marshal(nil)
	return buf, nil
}

func streamingUnknownHandler(ctx context.Context, service, method string, stream generic.BidiStreamingServer) (err error) {
	if service != serviceName {
		return fmt.Errorf("service not match")
	}
	if method == "UnaryTest" {
		for {
			req, err := stream.Recv(ctx)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			request := &mock.UnaryTestArgs{}
			err = request.Unmarshal(req.([]byte))
			if err != nil {
				return err
			}
			if request.Req.Message != "hello world" {
				return fmt.Errorf("message not match")
			}
			response := &mock.UnaryTestResult{Success: &pbapi.MockResp{Message: "hello world"}}
			buf, _ := response.Marshal(nil)
			err = stream.Send(ctx, buf)
			if err != nil {
				return err
			}
		}
	} else {
		if method != "ClientStreamingTest" && method != "ServerStreamingTest" && method != "BidirectionalStreamingTest" {
			return fmt.Errorf("method not match")
		}
		for {
			req, err := stream.Recv(ctx)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			request := &pbapi.MockReq{}
			err = request.Unmarshal(req.([]byte))
			if err != nil {
				return err
			}
			if request.Message != "hello world" {
				return fmt.Errorf("message not match")
			}
			response := &pbapi.MockResp{Message: "hello world"}
			buf, _ := response.Marshal(nil)
			err = stream.Send(ctx, buf)
			if err != nil {
				return err
			}
		}
	}
}

func newMockTestServer(handler pbapi.Mock, ln net.Listener, opts ...server.Option) server.Server {
	opts = append(opts, server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler))
	svr := mock.NewServer(handler, opts...)
	go func() {
		err := svr.Run()
		if err != nil {
			panic(err)
		}
	}()
	// wait for server starting to avoid data race
	time.Sleep(100 * time.Millisecond)
	return svr
}

var _ pbapi.Mock = &StreamingTestImpl{}

type StreamingTestImpl struct{}

func (s *StreamingTestImpl) UnaryTest(ctx context.Context, req *pbapi.MockReq) (resp *pbapi.MockResp, err error) {
	if req.Message != "hello world" {
		return nil, fmt.Errorf("message not match")
	}
	return &pbapi.MockResp{Message: "hello world"}, nil
}

func (s *StreamingTestImpl) ClientStreamingTest(ctx context.Context, stream pbapi.Mock_ClientStreamingTestServer) (err error) {
	for {
		req, err := stream.Recv(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if req.Message != "hello world" {
			return fmt.Errorf("message not match")
		}
	}
	return stream.SendAndClose(ctx, &pbapi.MockResp{Message: "hello world"})
}

func (s *StreamingTestImpl) ServerStreamingTest(ctx context.Context, req *pbapi.MockReq, stream pbapi.Mock_ServerStreamingTestServer) (err error) {
	if req.Message != "hello world" {
		return fmt.Errorf("message not match")
	}
	resp := &pbapi.MockResp{
		Message: "hello world",
	}
	return stream.Send(ctx, resp)
}

func (s *StreamingTestImpl) BidirectionalStreamingTest(ctx context.Context, stream pbapi.Mock_BidirectionalStreamingTestServer) (err error) {
	for {
		req, err := stream.Recv(ctx)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if req.Message != "hello world" {
			return fmt.Errorf("message not match")
		}
		res := &pbapi.MockResp{Message: "hello world"}
		err = stream.Send(ctx, res)
		if err != nil {
			return err
		}
	}
}
