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
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service/servicea"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service/serviceb"
	"github.com/cloudwego/kitex-tests/pkg/test"
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
	svr.RegisterService(servicea.NewServiceInfo(), new(ServiceAImpl))
	svr.RegisterService(serviceb.NewServiceInfo(), new(ServiceBImpl))

	return svr
}

func TestMultiService(t *testing.T) {
	ip := "localhost:9898"
	svr := GetServer(ip)
	go svr.Run()
	defer svr.Stop()

	time.Sleep(time.Second)
	clientA, err := servicea.NewClient("ServiceA", client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)

	_, err = clientA.EchoA(context.Background())
	test.Assert(t, err == nil, err)

	clientB, err := serviceb.NewClient("ServiceB", client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(ip))
	test.Assert(t, err == nil, err)

	_, err = clientB.EchoB(context.Background())
	test.Assert(t, err == nil, err)
}
