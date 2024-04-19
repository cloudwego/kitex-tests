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

package compressor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex/server"
)

type ServiceAImpl struct{}

func (s *ServiceAImpl) CallUnary(ctx context.Context, req *grpc_demo.Request) (res *grpc_demo.Reply, err error) {
	res = &grpc_demo.Reply{Message: req.Name + " Hello!"}
	return
}

func (s *ServiceAImpl) CallClientStream(stream grpc_demo.ServiceA_CallClientStreamServer) error {
	var msgs []string
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		msgs = append(msgs, req.Name)
	}
	return stream.SendAndClose(&grpc_demo.Reply{Message: "all message: " + strings.Join(msgs, ", ")})
}

func (s *ServiceAImpl) CallServerStream(req *grpc_demo.Request, stream grpc_demo.ServiceA_CallServerStreamServer) error {
	resp := &grpc_demo.Reply{}
	for i := 0; i < 3; i++ {
		resp.Message = fmt.Sprintf("%v-%d", req.Name, i)
		err := stream.Send(resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ServiceAImpl) CallBidiStream(stream grpc_demo.ServiceA_CallBidiStreamServer) error {
	for {
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		resp := &grpc_demo.Reply{}
		resp.Message = recv.Name
		err = stream.Send(resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func RunServer(hostport string) {

	addr, _ := net.ResolveTCPAddr("tcp", hostport)
	svr := servicea.NewServer(new(ServiceAImpl), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
