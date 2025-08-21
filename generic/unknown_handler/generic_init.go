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

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/echoservice"
)

const (
	serviceName        = "EchoService"
	unknownServiceName = "UnknownService"
	unknownMethodName  = "UnknownMethod"
	unknownMessage     = "UnknownMessage"
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

func defaultUnknownHandler(ctx context.Context, service, method string, request interface{}) (response interface{}, err error) {
	ri := rpcinfo.GetRPCInfo(ctx)
	if (service != unknownServiceName && service != serviceName) && ri.Config().TransportProtocol() != transport.Framed {
		return nil, fmt.Errorf("service not match")
	}
	if method != unknownMethodName {
		return nil, fmt.Errorf("method not match")
	}
	if string(request.([]byte)) != unknownMessage {
		return nil, fmt.Errorf("message not match")
	}
	return []byte(unknownMessage), nil
}

func streamingUnknownHandler(ctx context.Context, service, method string, stream generic.BidiStreamingServer) (err error) {
	if service != unknownServiceName && service != serviceName {
		return fmt.Errorf("service not match")
	}
	if method != unknownMethodName {
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
		if string(req.([]byte)) != unknownMessage {
			return fmt.Errorf("message not match")
		}
		err = stream.Send(ctx, []byte(unknownMessage))
		if err != nil {
			return err
		}
	}
}

func newMockTestServer(handler tenant.EchoService, ln net.Listener, opts ...server.Option) server.Server {
	opts = append(opts, server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler))
	svr := echoservice.NewServer(handler, opts...)
	err := genericserver.RegisterUnknownServiceOrMethodHandler(svr, &genericserver.UnknownServiceOrMethodHandler{
		DefaultHandler:   defaultUnknownHandler,
		StreamingHandler: streamingUnknownHandler,
	})
	if err != nil {
		panic(err)
	}
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
