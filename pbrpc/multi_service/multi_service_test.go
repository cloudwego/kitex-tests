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

package multi_service

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pb_multi_service"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pb_multi_service/servicea"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pb_multi_service/serviceb"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pb_multi_service/servicec"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

type ServiceAHandler struct{}

func (*ServiceAHandler) Chat1(ctx context.Context, req *pb_multi_service.Request) (resp *pb_multi_service.Reply, err error) {
	println("servicea Chat1 called, req:", req.Name)
	return &pb_multi_service.Reply{Message: "servicea Chat1"}, nil
}

type ServiceBHandler struct{}

func (*ServiceBHandler) Chat2(ctx context.Context, req *pb_multi_service.Request) (resp *pb_multi_service.Reply, err error) {
	println("serviceb Chat2 called, req:", req.Name)
	return &pb_multi_service.Reply{Message: "serviceb Chat2"}, nil
}

type ServiceCHandler struct{}

func (*ServiceCHandler) Chat1(ctx context.Context, req *pb_multi_service.Request) (resp *pb_multi_service.Reply, err error) {
	println("servicec Chat1 called, req:", req.Name)
	return &pb_multi_service.Reply{Message: "servicec Chat1"}, nil
}

func GetServer(hostport string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", hostport)

	return server.NewServer(server.WithServiceAddr(addr))
}

func TestRegisterService(t *testing.T) {
	ip := "localhost:9901"
	svr := GetServer(ip)
	err := servicea.RegisterService(svr, new(ServiceAHandler), true)
	test.Assert(t, err == nil)
	err = serviceb.RegisterService(svr, new(ServiceBHandler))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCHandler))
	test.Assert(t, err == nil)

	svr = GetServer(ip)
	test.PanicAt(t, func() {
		_ = servicea.RegisterService(svr, new(ServiceAHandler), true)
		_ = serviceb.RegisterService(svr, new(ServiceBHandler), true)
	}, func(err interface{}) bool {
		if errMsg, ok := err.(string); ok {
			return strings.Contains(errMsg, "multiple fallback services cannot be registered")
		}
		return true
	})

	svr = GetServer(ip)
	err = servicea.RegisterService(svr, new(ServiceAHandler))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCHandler))
	test.Assert(t, err == nil)
	err = svr.Run()
	test.Assert(t, err != nil)
	test.Assert(t, err.Error() == "method name [Chat1] is conflicted between services but no fallback service is specified")
}

func TestMultiService(t *testing.T) {
	ip := "localhost:9900"
	svr := GetServer(ip)
	servicea.RegisterService(svr, new(ServiceAHandler))
	serviceb.RegisterService(svr, new(ServiceBHandler))
	servicec.RegisterService(svr, new(ServiceCHandler), true)
	go svr.Run()
	defer svr.Stop()

	req := &pb_multi_service.Request{Name: "pb multi_service req"}

	time.Sleep(time.Second)
	clientA, err := servicea.NewClient("ServiceA", client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err := clientA.Chat1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicea Chat1")

	clientAWithTTHeader, err := servicea.NewClient("ServiceA", client.WithTransportProtocol(transport.TTHeader), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientAWithTTHeader.Chat1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicea Chat1")

	clientB, err := serviceb.NewClient("ServiceB", client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientB.Chat2(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "serviceb Chat2")

	clientC, err := servicec.NewClient("ServiceC", client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientC.Chat1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicec Chat1")
}
