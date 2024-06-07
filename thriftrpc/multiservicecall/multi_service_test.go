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

package multiservicecall

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service/servicea"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service/serviceb"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service/servicec"
	"github.com/cloudwego/kitex-tests/pkg/test"
)

// ServiceAImpl implements the last servicea interface defined in the IDL.
type ServiceAImpl struct{}

// Echo1 implements the Echo1 interface.
func (s *ServiceAImpl) Echo1(ctx context.Context, req *multi_service.Request) (resp *multi_service.Response, err error) {
	println("servicea Echo1 called, req:", req.Message)
	return &multi_service.Response{Message: "servicea Echo1"}, nil
}

// ServiceBImpl implements the last serviceb interface defined in the IDL.
type ServiceBImpl struct{}

// Echo2 implements the Echo2 interface.
func (s *ServiceBImpl) Echo2(ctx context.Context, req *multi_service.Request) (resp *multi_service.Response, err error) {
	println("serviceb Echo2 called, req:", req.Message)
	return &multi_service.Response{Message: "serviceb Echo2"}, nil
}

// ServiceCImpl implements the last servicec interface defined in the IDL.
type ServiceCImpl struct{}

// Echo1 implements the Echo1 interface.
func (s *ServiceCImpl) Echo1(ctx context.Context, req *multi_service.Request) (resp *multi_service.Response, err error) {
	println("servicec Echo1 called, req:", req.Message)
	return &multi_service.Response{Message: "servicec Echo1"}, nil
}

func GetServer(hostport string, opts ...server.Option) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", hostport)
	opts = append(opts, server.WithServiceAddr(addr))

	return server.NewServer(opts...)
}

func TestRegisterService(t *testing.T) {
	testRegisterService(t, "localhost:9900")
}

func TestMuxRegisterService(t *testing.T) {
	testRegisterService(t, "localhost:9901", server.WithMuxTransport())
}

func TestMultiServiceWithRefuseTrafficWithoutServiceName(t *testing.T) {
	testMultiServiceWithRefuseTrafficWithoutServiceName(t, "localhost:9902", server.WithRefuseTrafficWithoutServiceName())
}

func TestMuxMultiServiceWithRefuseTrafficWithoutServiceName(t *testing.T) {
	testMultiServiceWithRefuseTrafficWithoutServiceName(t, "localhost:9903", server.WithRefuseTrafficWithoutServiceName(), server.WithMuxTransport())
}

func TestMultiService(t *testing.T) {
	testMultiService(t, "localhost:9904")
}

func TestMuxMultiService(t *testing.T) {
	testMultiService(t, "localhost:9905", server.WithMuxTransport())
}

/*
func TestMutliServiceWithGenericClient(t *testing.T) {
	ip := "localhost:9907"
	svr := GetServer(ip)
	err := servicea.RegisterService(svr, new(ServiceAImpl))
	test.Assert(t, err == nil)
	err = serviceb.RegisterService(svr, new(ServiceBImpl))
	test.Assert(t, err == nil)
	go svr.Run()
	defer svr.Stop()

	p, err := generic.NewThriftFileProvider("./idl/thrift_multi_service.thrift")
	test.Assert(t, err == nil)
	g, err := generic.JSONThriftGeneric(p)
	test.Assert(t, err == nil)
	cli, err := genericclient.NewClient("genericservice", g,
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler))
	test.Assert(t, err == nil)
	resp, err := cli.GenericCall(context.Background(), "Echo1", `{"message":"generic request"}`)
	test.Assert(t, err == nil)
	respStr, ok := resp.(string)
	test.Assert(t, ok)
	test.Assert(t, reflect.DeepEqual(gjson.Get(respStr, "message").String(), "servicea Echo1"))
}
*/

func TestMultiServiceWithCombineServiceClient(t *testing.T) {
	ip := "localhost:9906"
	svr := GetServer(ip)
	err := servicea.RegisterService(svr, new(ServiceAImpl))
	test.Assert(t, err == nil)
	err = serviceb.RegisterService(svr, new(ServiceBImpl))
	test.Assert(t, err == nil)
	go svr.Run()
	defer svr.Stop()

	combineServiceClient, err := combineservice.NewClient("combineservice",
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithHostPorts(ip),
	)
	test.Assert(t, err == nil)
	req := &combine_service.Request{Message: "req from client using combine service"}
	resp, err := combineServiceClient.Echo1(context.Background(), req)
	test.Assert(t, err == nil)
	test.Assert(t, resp.Message == "servicea Echo1")
}

func testRegisterService(t *testing.T, ip string, opts ...server.Option) {
	svr := GetServer(ip, opts...)
	err := servicea.RegisterService(svr, new(ServiceAImpl), server.WithFallbackService())
	test.Assert(t, err == nil)
	err = serviceb.RegisterService(svr, new(ServiceBImpl))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCImpl))
	test.Assert(t, err == nil)

	svr = GetServer(ip, opts...)
	test.PanicAt(t, func() {
		_ = servicea.RegisterService(svr, new(ServiceAImpl), server.WithFallbackService())
		_ = serviceb.RegisterService(svr, new(ServiceBImpl), server.WithFallbackService())
	}, func(err interface{}) bool {
		if errMsg, ok := err.(string); ok {
			return strings.Contains(errMsg, "multiple fallback services cannot be registered")
		}
		return true
	})

	svr = GetServer(ip, opts...)
	err = servicea.RegisterService(svr, new(ServiceAImpl))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCImpl))
	test.Assert(t, err == nil)
	err = svr.Run()
	test.Assert(t, err != nil)
	test.Assert(t, err.Error() == "method name [Echo1] is conflicted between services but no fallback service is specified")
}

func testMultiServiceWithRefuseTrafficWithoutServiceName(t *testing.T, ip string, opts ...server.Option) {
	svr := GetServer(ip, opts...)
	err := servicea.RegisterService(svr, new(ServiceAImpl))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCImpl))
	test.Assert(t, err == nil)
	go svr.Run()
	defer svr.Stop()

	clientA, err := servicea.NewClient("ServiceA", client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	_, err = clientA.Echo1(context.Background(), &multi_service.Request{Message: "multi_service req"})
	test.Assert(t, err != nil)
}

func testMultiService(t *testing.T, ip string, opts ...server.Option) {
	svr := GetServer(ip, opts...)
	servicea.RegisterService(svr, new(ServiceAImpl))
	serviceb.RegisterService(svr, new(ServiceBImpl))
	servicec.RegisterService(svr, new(ServiceCImpl), server.WithFallbackService())
	go svr.Run()
	defer svr.Stop()

	req := &multi_service.Request{Message: "multi_service req"}

	time.Sleep(time.Second)
	clientA, err := servicea.NewClient("ServiceA", client.WithTransportProtocol(transport.TTHeader), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err := clientA.Echo1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicec Echo1")

	clientAWithMetaHandler, err := servicea.NewClient("ServiceA",
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientAWithMetaHandler.Echo1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicea Echo1")

	clientB, err := serviceb.NewClient("ServiceB", client.WithTransportProtocol(transport.TTHeader), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientB.Echo2(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "serviceb Echo2")

	clientBWithMetaHandler, err := serviceb.NewClient("ServiceB",
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientBWithMetaHandler.Echo2(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "serviceb Echo2")

	clientC, err := servicec.NewClient("ServiceC", client.WithTransportProtocol(transport.TTHeader), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientC.Echo1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicec Echo1")

	clientCWithMetaHandler, err := servicec.NewClient("ServiceC",
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)
	resp, err = clientCWithMetaHandler.Echo1(context.Background(), req)
	test.Assert(t, err == nil, err)
	test.Assert(t, resp.Message == "servicec Echo1")
}
