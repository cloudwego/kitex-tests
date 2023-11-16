// Copyright 2022 CloudWeGo Authors
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

package unknown_handler

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/unknown_handler"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/unknown_handler/servicea"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/unknown_handler/serviceb"
	"github.com/cloudwego/kitex-tests/pkg/test"
)

type ServiceAImpl struct{}

func (s *ServiceAImpl) Echo(ctx context.Context, req *unknown_handler.Request) (rep *unknown_handler.Reply, err error) {
	return &unknown_handler.Reply{Message: req.Name}, nil
}

func handler(ctx context.Context, methodName string, stream streaming.Stream) error {
	req := &unknown_handler.Request{}
	err := stream.RecvMsg(req)
	if err != nil {
		return err
	}
	resp := &unknown_handler.Reply{Message: req.Name}
	return stream.SendMsg(resp)
}

func GetServer(ip string, hasHandler bool) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", ip)
	var opts []server.Option
	opts = append(opts, server.WithServiceAddr(addr))
	if hasHandler {
		opts = append(opts, server.WithGRPCUnknownServiceHandler(handler))
	}
	return serviceb.NewServer(new(unknown_handler.ServiceB), opts...)
}

// TestUnknownServiceError test if the server send unknown method err when there is no matching method.
func TestUnknownServiceError(t *testing.T) {
	ip := "localhost:9899" // avoid conflict with other tests
	svr := GetServer(ip, false)
	go svr.Run()
	defer svr.Stop()

	time.Sleep(time.Second)
	client, err := servicea.NewClient("grpcService", client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)

	req := &unknown_handler.Request{Name: "kitex"}
	_, err = client.Echo(context.Background(), req)
	test.Assert(t, err != nil, err)
	test.Assert(t, strings.Contains(err.Error(), "remote or network error: rpc error: code = 20 desc = unknown service ServiceA"), err)
}

// TestUnknownServiceHandler test if handler works when there is no matching method on server.
func TestUnknownServiceHandler(t *testing.T) {
	ip := "localhost:9898"
	svr := GetServer(ip, true)
	go svr.Run()
	defer svr.Stop()

	time.Sleep(time.Second)
	client, err := servicea.NewClient("grpcService", client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)

	req := &unknown_handler.Request{Name: "kitex"}
	resp, err := client.Echo(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp != nil)
	test.Assert(t, resp.Message == req.Name)
}
