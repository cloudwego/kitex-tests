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
	"strings"
	"sync"

	"github.com/cloudwego/kitex-tests/streamx/kitex_gen/echo"
	"github.com/cloudwego/kitex-tests/streamx/kitex_gen/echo/testservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/server"
)

var startChan chan struct{}

func init() {
	startChan = make(chan struct{})
	server.RegisterStartHook(func() {
		startChan <- struct{}{}
	})
}

func newGenericClient(g generic.Generic, targetIPPort string, cliOpts ...client.Option) genericclient.Client {
	cliOpts = append(cliOpts, client.WithHostPorts(targetIPPort))
	cli, err := genericclient.NewClient("destService", g, cliOpts...)
	if err != nil {
		panic(err)
	}
	return cli
}

func newMockTestServer(handler echo.TestService, ln net.Listener, opts ...server.Option) server.Server {
	opts = append(opts, server.WithListener(ln))
	svr := testservice.NewServer(handler, opts...)
	go func() {
		err := svr.Run()
		if err != nil {
			panic(err)
		}
	}()
	// wait for server starting to avoid data race
	<-startChan
	return svr
}

type serviceImpl struct {
}

func (s *serviceImpl) PingPong(ctx context.Context, req *echo.EchoClientRequest) (r *echo.EchoClientResponse, err error) {
	return nil, nil
}

func (s *serviceImpl) Unary(ctx context.Context, req *echo.EchoClientRequest) (r *echo.EchoClientResponse, err error) {
	fmt.Println("UnaryTest called")
	r = &echo.EchoClientResponse{Message: "hello " + req.Message}
	return
}

func (s *serviceImpl) EchoClient(ctx context.Context, stream echo.TestService_EchoClientServer) (err error) {
	fmt.Println("ClientStreamingTest called")
	var msgs []string
	for {
		req, err := stream.Recv(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Printf("Recv: %s\n", req.Message)
		msgs = append(msgs, req.Message)
	}
	return stream.SendAndClose(ctx, &echo.EchoClientResponse{Message: "all message: " + strings.Join(msgs, ", ")})
}

func (s *serviceImpl) EchoServer(ctx context.Context, req *echo.EchoClientRequest, stream echo.TestService_EchoServerServer) (err error) {
	fmt.Println("ServerStreamingTest called")
	resp := &echo.EchoClientResponse{}
	for i := 0; i < 3; i++ {
		resp.Message = fmt.Sprintf("%v response", req.Message)
		err := stream.Send(ctx, resp)
		if err != nil {
			return err
		}
	}
	return
}

func (s *serviceImpl) EchoBidi(ctx context.Context, stream echo.TestService_EchoBidiServer) (err error) {
	fmt.Println("BidirectionalStreamingTest called")
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("panic: %v", p)
			}
			wg.Done()
		}()
		for {
			msg, recvErr := stream.Recv(ctx)
			if recvErr == io.EOF {
				return
			} else if recvErr != nil {
				err = recvErr
				return
			}
			fmt.Printf("BidirectionaStreamingTest: received message = %s\n", msg.Message)
		}
	}()

	go func() {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("panic: %v", p)
			}
			wg.Done()
		}()
		resp := &echo.EchoClientResponse{}
		for i := 0; i < 3; i++ {
			resp.Message = "Hello from Server"
			if sendErr := stream.Send(ctx, resp); sendErr != nil {
				err = sendErr
				return
			}
			fmt.Printf("BidirectionaStreamingTest: sent message = %s\n", resp)
		}
	}()
	wg.Wait()
	return
}
