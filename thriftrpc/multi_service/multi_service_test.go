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

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service/servicea"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service/serviceb"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service/servicec"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

// ServiceAImpl implements the last servicea interface defined in the IDL.
type ServiceAImpl struct{}

// Echo1 implements the Echo1 interface.
func (s *ServiceAImpl) Echo1(ctx context.Context, req *multi_service.Request) (resp *multi_service.Response, err error) {
	println(req.Message)
	return &multi_service.Response{Message: "servicea echo1 called"}, nil
}

// ServiceBImpl implements the last serviceb interface defined in the IDL.
type ServiceBImpl struct{}

// Echo2 implements the Echo2 interface.
func (s *ServiceBImpl) Echo2(ctx context.Context, req *multi_service.Request) (resp *multi_service.Response, err error) {
	println(req.Message)
	return &multi_service.Response{Message: "echo2 called"}, nil
}

// ServiceCImpl implements the last servicec interface defined in the IDL.
type ServiceCImpl struct{}

// Echo1 implements the Echo1 interface.
func (s *ServiceCImpl) Echo1(ctx context.Context, req *multi_service.Request) (resp *multi_service.Response, err error) {
	println(req.Message)
	return &multi_service.Response{Message: "servicec echo1 called"}, nil
}

func GetServer(hostport string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", hostport)

	return server.NewServer(server.WithServiceAddr(addr))
}

func TestRegisterService(t *testing.T) {
	ip := "localhost:9900"
	svr := GetServer(ip)
	err := servicea.RegisterService(svr, new(ServiceAImpl), true)
	test.Assert(t, err == nil)
	err = serviceb.RegisterService(svr, new(ServiceBImpl))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCImpl))
	test.Assert(t, err == nil)

	svr = GetServer(ip)
	test.PanicAt(t, func() {
		_ = servicea.RegisterService(svr, new(ServiceAImpl), true)
		_ = serviceb.RegisterService(svr, new(ServiceBImpl), true)
	}, func(err interface{}) bool {
		if errMsg, ok := err.(string); ok {
			return strings.Contains(errMsg, "multiple fallback services cannot be registered")
		}
		return true
	})

	svr = GetServer(ip)
	err = servicea.RegisterService(svr, new(ServiceAImpl))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCImpl))
	test.Assert(t, err == nil)
	err = svr.Run()
	test.Assert(t, err != nil)
	test.Assert(t, err.Error() == "method name [Echo1] is conflicted between services but no fallback service is specified")
}

func TestMultiService(t *testing.T) {
	ip := "localhost:9900"
	svr := GetServer(ip)
	servicea.RegisterService(svr, new(ServiceAImpl))
	serviceb.RegisterService(svr, new(ServiceBImpl))
	servicec.RegisterService(svr, new(ServiceCImpl), true)
	go svr.Run()
	defer svr.Stop()

	req := &multi_service.Request{Message: "multi_service req"}

	time.Sleep(time.Second)
	clientA, err := servicea.NewClient("ServiceA", client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err := clientA.Echo1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicec echo1 called")

	clientAWithTTHeader, err := servicea.NewClient("ServiceA", client.WithTransportProtocol(transport.TTHeader), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientAWithTTHeader.Echo1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicea echo1 called")

	clientB, err := serviceb.NewClient("ServiceB", client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientB.Echo2(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "echo2 called")

	clientBWithTTHeader, err := serviceb.NewClient("ServiceB", client.WithTransportProtocol(transport.TTHeader), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientBWithTTHeader.Echo2(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "echo2 called")

	clientC, err := servicec.NewClient("ServiceC", client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientC.Echo1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicec echo1 called")

	clientCWithTTHeader, err := servicec.NewClient("ServiceC", client.WithTransportProtocol(transport.TTHeader), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientCWithTTHeader.Echo1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicec echo1 called")
}
