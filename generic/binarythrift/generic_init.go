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

	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/server/genericserver"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/echoservice"
)

const serviceName = "EchoService"

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
	if method != "Echo" {
		return nil, fmt.Errorf("method not match")
	}
	args := &tenant.EchoServiceEchoArgs{}
	err = thrift.FastUnmarshal(request.([]byte), args)
	if err != nil {
		return nil, err
	}
	if args.Request.Msg != "hello world" {
		return nil, fmt.Errorf("message not match")
	}
	res := &tenant.EchoServiceEchoResult{
		Success: &tenant.EchoResponse{Msg: "hello world"},
	}
	buf := thrift.FastMarshal(res)
	return buf, nil
}

func streamingUnknownHandler(ctx context.Context, service, method string, stream generic.BidiStreamingServer) (err error) {
	if service != serviceName {
		return fmt.Errorf("service not match")
	}
	if method == "Echo" {
		for {
			req, err := stream.Recv(ctx)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			request := &tenant.EchoServiceEchoArgs{}
			err = thrift.FastUnmarshal(req.([]byte), request)
			if err != nil {
				return err
			}
			if request.Request.Msg != "hello world" {
				return fmt.Errorf("message not match")
			}
			response := &tenant.EchoServiceEchoResult{Success: &tenant.EchoResponse{Msg: "hello world"}}
			buf := thrift.FastMarshal(response)
			err = stream.Send(ctx, buf)
			if err != nil {
				return err
			}
		}
	} else {
		if method != "EchoClient" && method != "EchoServer" && method != "EchoBidi" {
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
			request := &tenant.EchoRequest{}
			err = thrift.FastUnmarshal(req.([]byte), request)
			if err != nil {
				return err
			}
			if request.Msg != "hello world" {
				return fmt.Errorf("message not match")
			}
			response := &tenant.EchoResponse{Msg: "hello world"}
			buf := thrift.FastMarshal(response)
			err = stream.Send(ctx, buf)
			if err != nil {
				return err
			}
		}
	}
}

func newMockTestServer(handler tenant.EchoService, ln net.Listener, opts ...server.Option) server.Server {
	opts = append(opts, server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler))
	svr := echoservice.NewServer(handler, opts...)
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

type serviceImpl struct {
	tenant.EchoService
}

func (s *serviceImpl) Echo(ctx context.Context, req *tenant.EchoRequest) (r *tenant.EchoResponse, err error) {
	if req.Msg != "hello world" {
		return nil, fmt.Errorf("message not match")
	}
	return &tenant.EchoResponse{Msg: "hello world"}, nil
}

func (s *serviceImpl) EchoClient(ctx context.Context, stream tenant.EchoService_EchoClientServer) (err error) {
	for {
		req, err := stream.Recv(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if req.Msg != "hello world" {
			return fmt.Errorf("message not match")
		}
	}
	return stream.SendAndClose(ctx, &tenant.EchoResponse{Msg: "hello world"})
}

func (s *serviceImpl) EchoServer(ctx context.Context, req *tenant.EchoRequest, stream tenant.EchoService_EchoServerServer) (err error) {
	if req.Msg != "hello world" {
		return fmt.Errorf("message not match")
	}
	resp := &tenant.EchoResponse{
		Msg: "hello world",
	}
	return stream.Send(ctx, resp)
}

func (s *serviceImpl) EchoBidi(ctx context.Context, stream tenant.EchoService_EchoBidiServer) (err error) {
	for {
		req, err := stream.Recv(ctx)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if req.Msg != "hello world" {
			return fmt.Errorf("message not match")
		}
		res := &tenant.EchoResponse{Msg: "hello world"}
		err = stream.Send(ctx, res)
		if err != nil {
			return err
		}
	}
}
