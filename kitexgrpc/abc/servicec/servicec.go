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

package servicec

import (
	"context"
	"io"
	"net"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/consts"
	"github.com/cloudwego/kitex/server"
)

func InitServiceCServer() server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", consts.ServiceCAddr)
	svr := servicea.NewServer(new(ServiceCImpl), server.WithServiceAddr(addr))
	return svr
}

type ServiceCImpl struct {
}

func (s ServiceCImpl) CallUnary(ctx context.Context, req *grpc_demo.Request) (res *grpc_demo.Reply, err error) {
	res = &grpc_demo.Reply{Message: req.Name}
	return
}

func (s ServiceCImpl) CallClientStream(stream grpc_demo.ServiceA_CallClientStreamServer) (err error) {
	replyMsg := ""
	for {
		msg, err := stream.Recv()
		if err == nil {
			replyMsg = msg.Name
		} else if err == io.EOF {
			break
		} else {
			return err
		}
	}
	return stream.SendAndClose(&grpc_demo.Reply{Message: replyMsg})
}

func (s ServiceCImpl) CallServerStream(req *grpc_demo.Request, stream grpc_demo.ServiceA_CallServerStreamServer) (err error) {
	for i := 0; i < 3; i++ {
		if err = stream.Send(&grpc_demo.Reply{Message: req.Name}); err != nil {
			return err
		}
	}
	return
}

func (s ServiceCImpl) CallBidiStream(stream grpc_demo.ServiceA_CallBidiStreamServer) (err error) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		err = stream.Send(&grpc_demo.Reply{Message: msg.Name})
		if err != nil {
			return err
		}
	}
}
