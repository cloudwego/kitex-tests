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
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/combine_service"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/combine_service/combineservice"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service/servicea"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service/serviceb"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service/servicec"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service2"
	servicea2 "github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service2/servicea"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
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

func GetServer(ln net.Listener, opts ...server.Option) server.Server {
	opts = append(opts, server.WithListener(ln),
		server.WithExitWaitTime(20*time.Millisecond),
	)
	return server.NewServer(opts...)
}

func TestRegisterService(t *testing.T) {
	testRegisterService(t, serverutils.Listen())
}

func TestMuxRegisterService(t *testing.T) {
	testRegisterService(t, serverutils.Listen(), server.WithMuxTransport())
}

func TestMultiServiceWithRefuseTrafficWithoutServiceName(t *testing.T) {
	testMultiServiceWithRefuseTrafficWithoutServiceName(t,
		serverutils.Listen(), server.WithRefuseTrafficWithoutServiceName())
}

func TestMuxMultiServiceWithRefuseTrafficWithoutServiceName(t *testing.T) {
	testMultiServiceWithRefuseTrafficWithoutServiceName(t, serverutils.Listen(),
		server.WithRefuseTrafficWithoutServiceName(), server.WithMuxTransport())
}

func TestMultiService(t *testing.T) {
	testMultiService(t, serverutils.Listen())
}

func TestMuxMultiService(t *testing.T) {
	testMultiService(t, serverutils.Listen(), server.WithMuxTransport())
}

func TestMultiServiceWithCombineServiceClient(t *testing.T) {
	ln := serverutils.Listen()
	ip := ln.Addr().String()
	svr := GetServer(ln)
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

func TestUnknownException(t *testing.T) {
	ln := serverutils.Listen()
	ip := ln.Addr().String()
	svr := GetServer(ln)
	servicea.RegisterService(svr, new(ServiceAImpl))
	go svr.Run()
	defer svr.Stop()

	clientB, err := serviceb.NewClient("ServiceB",
		client.WithHostPorts(ip),
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler))
	test.Assert(t, err == nil, err)
	_, err = clientB.Echo2(context.Background(), &multi_service.Request{Message: "multi_service req"})
	test.Assert(t, err != nil)
	test.DeepEqual(t, err.Error(), "remote or network error[remote]: unknown method Echo2")
}

func TestUnknownExceptionWithMultiService(t *testing.T) {
	ln := serverutils.Listen()
	hostport := ln.Addr().String()
	svr := GetServer(ln)
	servicea.RegisterService(svr, new(ServiceAImpl))
	servicec.RegisterService(svr, new(ServiceCImpl), server.WithFallbackService())
	go svr.Run()
	defer svr.Stop()

	// unknown service error
	clientB, err := serviceb.NewClient("ServiceB",
		client.WithHostPorts(hostport),
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler))
	test.Assert(t, err == nil, err)
	_, err = clientB.Echo2(context.Background(), &multi_service.Request{Message: "multi_service req"})
	test.Assert(t, err != nil)
	fmt.Println(err)
	test.DeepEqual(t, err.Error(), "remote or network error[remote]: unknown service ServiceB, method Echo2")

	// unknown method error
	clientA, err := servicea2.NewClient("ServiceA",
		client.WithHostPorts(hostport),
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler))
	test.Assert(t, err == nil, err)
	_, err = clientA.EchoA(context.Background(), &multi_service2.Request{Message: "multi_service req"})
	test.Assert(t, err != nil)
	test.DeepEqual(t, err.Error(), "remote or network error[remote]: unknown method EchoA (service ServiceA)")
}

func testRegisterService(t *testing.T, ln net.Listener, opts ...server.Option) {
	svr := GetServer(ln, opts...)
	err := servicea.RegisterService(svr, new(ServiceAImpl), server.WithFallbackService())
	test.Assert(t, err == nil)
	err = serviceb.RegisterService(svr, new(ServiceBImpl))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCImpl))
	test.Assert(t, err == nil)

	svr = GetServer(ln, opts...)
	test.PanicAt(t, func() {
		_ = servicea.RegisterService(svr, new(ServiceAImpl), server.WithFallbackService())
		_ = serviceb.RegisterService(svr, new(ServiceBImpl), server.WithFallbackService())
	}, func(err interface{}) bool {
		if errMsg, ok := err.(string); ok {
			return strings.Contains(errMsg, "multiple fallback services cannot be registered")
		}
		return true
	})

	svr = GetServer(ln, opts...)
	err = servicea.RegisterService(svr, new(ServiceAImpl))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCImpl))
	test.Assert(t, err == nil)
	err = svr.Run()
	test.Assert(t, err != nil)
	test.Assert(t, err.Error() == "method name [Echo1] is conflicted between services but no fallback service is specified")
}

func testMultiServiceWithRefuseTrafficWithoutServiceName(t *testing.T, ln net.Listener, opts ...server.Option) {
	svr := GetServer(ln, opts...)
	err := servicea.RegisterService(svr, new(ServiceAImpl))
	test.Assert(t, err == nil)
	err = servicec.RegisterService(svr, new(ServiceCImpl))
	test.Assert(t, err == nil)
	go svr.Run()
	defer svr.Stop()

	clientA, err := servicea.NewClient("ServiceA", client.WithHostPorts(ln.Addr().String()))
	test.Assert(t, err == nil, err)
	_, err = clientA.Echo1(context.Background(), &multi_service.Request{Message: "multi_service req"})
	test.Assert(t, err != nil)
}

func testMultiService(t *testing.T, ln net.Listener, opts ...server.Option) {
	ip := ln.Addr().String()
	svr := GetServer(ln, opts...)
	servicea.RegisterService(svr, new(ServiceAImpl))
	serviceb.RegisterService(svr, new(ServiceBImpl))
	servicec.RegisterService(svr, new(ServiceCImpl), server.WithFallbackService())
	go svr.Run()
	defer svr.Stop()

	req := &multi_service.Request{Message: "multi_service req"}

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
