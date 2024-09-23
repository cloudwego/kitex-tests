// Copyright 2023 CloudWeGo Authors
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
	"io"
	"net"
	"testing"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service/servicea"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service/serviceb"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service/servicec"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service_2"
	servicea2 "github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service_2/servicea"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/clientutils"
)

type ServiceAImpl struct{}

func (s *ServiceAImpl) EchoA(stream grpc_multi_service.ServiceA_EchoAServer) error {
	for {
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		resp := &grpc_multi_service.ReplyA{}
		resp.Message = recv.Name
		err = stream.Send(resp)
		if err != nil {
			return err
		}
	}
	return nil
}

type ServiceBImpl struct{}

func (s *ServiceBImpl) EchoB(stream grpc_multi_service.ServiceB_EchoBServer) error {
	for {
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		resp := &grpc_multi_service.ReplyB{}
		resp.Message = recv.Name
		err = stream.Send(resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetServer(hostport string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", hostport)

	svr := server.NewServer(server.WithServiceAddr(addr))
	servicea.RegisterService(svr, new(ServiceAImpl))
	serviceb.RegisterService(svr, new(ServiceBImpl))

	return svr
}

func TestMultiService(t *testing.T) {
	hostport := "localhost:9898"
	svr := GetServer(hostport)
	go svr.Run()
	defer svr.Stop()

	clientA, err := servicea.NewClient("ServiceA", client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(hostport))
	test.Assert(t, err == nil, err)
	defer clientutils.CallClose(clientA)

	streamCliA, err := clientA.EchoA(context.Background())
	test.Assert(t, err == nil, err)
	streamCliA.Send(&grpc_multi_service.RequestA{Name: "ServiceA"})
	respA, err := streamCliA.Recv()
	test.Assert(t, err == nil)
	test.Assert(t, respA.Message == "ServiceA")
	streamCliA.Close()

	clientB, err := serviceb.NewClient("ServiceB", client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(hostport))
	test.Assert(t, err == nil, err)
	defer clientutils.CallClose(clientB)

	streamCliB, err := clientB.EchoB(context.Background())
	test.Assert(t, err == nil, err)
	streamCliB.Send(&grpc_multi_service.RequestB{Name: "ServiceB"})
	respB, err := streamCliB.Recv()
	test.Assert(t, err == nil)
	test.Assert(t, respB.Message == "ServiceB")
	streamCliB.Close()
}

func TestUnknownException(t *testing.T) {
	hostport := "localhost:9899"
	addr, _ := net.ResolveTCPAddr("tcp", hostport)
	svr := server.NewServer(server.WithServiceAddr(addr))
	servicea.RegisterService(svr, new(ServiceAImpl))
	go svr.Run()
	defer svr.Stop()

	clientC, err := servicec.NewClient("ServiceC", client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(hostport))
	test.Assert(t, err == nil, err)
	defer clientutils.CallClose(clientC)

	streamCliC, err := clientC.EchoC(context.Background())
	test.Assert(t, err == nil, err)
	streamCliC.Send(&grpc_multi_service.RequestC{Name: "ServiceC"})
	_, err = streamCliC.Recv()
	test.Assert(t, err != nil)
	test.DeepEqual(t, err.Error(), "rpc error: code = 20 desc = unknown service ServiceC")
}

func TestUnknownExceptionWithMultiService(t *testing.T) {
	hostport := "localhost:9900"
	svr := GetServer(hostport)
	go svr.Run()
	defer svr.Stop()

	// unknown service error
	clientC, err := servicec.NewClient("ServiceC", client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(hostport))
	test.Assert(t, err == nil, err)
	defer clientutils.CallClose(clientC)

	streamCliC, err := clientC.EchoC(context.Background())
	test.Assert(t, err == nil, err)
	streamCliC.Send(&grpc_multi_service.RequestC{Name: "ServiceC"})
	_, err = streamCliC.Recv()
	test.Assert(t, err != nil)
	test.DeepEqual(t, err.Error(), "rpc error: code = 20 desc = unknown service ServiceC")

	// unknown method error
	clientA, err := servicea2.NewClient("ServiceA", client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(hostport))
	test.Assert(t, err == nil, err)
	defer clientutils.CallClose(clientA)

	streamCliA, err := clientA.Echo(context.Background())
	test.Assert(t, err == nil, err)
	streamCliA.Send(&grpc_multi_service_2.Request{Name: "ServiceA"})
	_, err = streamCliA.Recv()
	test.Assert(t, err != nil)
	test.DeepEqual(t, err.Error(), "rpc error: code = 1 desc = unknown method Echo")
}
