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

package main

import (
	"context"
	"flag"
	"net"
	"os"
	"time"

	"github.com/cloudwego/kitex/server"

	"github.com/cloudwego/kitex-tests/thrift_streaming/exitserver/kitex_gen/echo"
	"github.com/cloudwego/kitex-tests/thrift_streaming/exitserver/kitex_gen/echo/echoservice"
)

func WithServerAddr(hostPort string) server.Option {
	addr, err := net.ResolveTCPAddr("tcp", hostPort)
	if err != nil {
		panic(err)
	}
	return server.WithServiceAddr(addr)
}

type handler struct {
	echo.EchoService
}

func (h handler) EchoBidirectional(stream echo.EchoService_EchoBidirectionalServer) (err error) {
	os.Exit(1)
	return nil
}

func (h handler) EchoClient(stream echo.EchoService_EchoClientServer) (err error) {
	os.Exit(2)
	return nil
}

func (h handler) EchoServer(req *echo.EchoRequest, stream echo.EchoService_EchoServerServer) (err error) {
	os.Exit(2)
	return nil
}

func (h handler) EchoUnary(ctx context.Context, req1 *echo.EchoRequest) (r *echo.EchoResponse, err error) {
	os.Exit(2)
	return nil, nil
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", "127.0.0.1:9999", "server address")
	flag.Parse()
	var opts []server.Option
	opts = append(opts, WithServerAddr(addr))
	opts = append(opts, server.WithExitWaitTime(time.Millisecond*10))
	svr := echoservice.NewServer(&handler{}, opts...)
	if err := svr.Run(); err != nil {
		panic(err)
	}
}
