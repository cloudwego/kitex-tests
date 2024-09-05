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

package serviceb

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/consts"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

func ServiceBMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		err = next(ctx, req, resp)
		//time.Sleep(500 * time.Millisecond)
		return err
	}
}

func InitServiceBServer() server.Server {
	cli, err := servicea.NewClient("serviceC",
		client.WithMiddleware(ServiceBMiddleware),
		client.WithHostPorts(consts.ServiceCAddr), client.WithTransportProtocol(transport.GRPC))
	if err != nil {
		panic(err)
	}
	addr, _ := net.ResolveTCPAddr("tcp", consts.ServiceBAddr)
	svr := servicea.NewServer(&ServiceBImpl{cli: cli}, server.WithServiceAddr(addr))
	return svr
}

type ServiceBImpl struct {
	cli servicea.Client
}

func (s ServiceBImpl) CallUnary(ctx context.Context, req *grpc_demo.Request) (res *grpc_demo.Reply, err error) {
	res, err = s.cli.CallUnary(ctx, req)
	fmt.Println(res, err)
	if dl, ok := ctx.Deadline(); ok {
		klog.Infof("dl=%v", dl)
	}
	return
}

func (s ServiceBImpl) CallClientStream(stream grpc_demo.ServiceA_CallClientStreamServer) (err error) {
	svcCStream, err := s.cli.CallClientStream(stream.Context())
	if err != nil {
		return err
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err = svcCStream.Send(msg); err != nil {
			return err
		}
	}
	reply, err := svcCStream.CloseAndRecv()
	if err != nil {
		return err
	}
	return stream.SendAndClose(reply)
}

func (s ServiceBImpl) CallServerStream(req *grpc_demo.Request, stream grpc_demo.ServiceA_CallServerStreamServer) (err error) {
	svcCStream, err := s.cli.CallServerStream(stream.Context(), req)
	if err != nil {
		return err
	}
	for {
		msg, err := svcCStream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err = stream.Send(msg); err != nil {
			return err
		}
	}
	return
}

func (s ServiceBImpl) CallBidiStream(stream grpc_demo.ServiceA_CallBidiStreamServer) error {
	svcCStream, err := s.cli.CallBidiStream(stream.Context())
	if err != nil {
		return err
	}
	for {
		msg, rErr := stream.Recv()
		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			return rErr
		}
		if sErr := svcCStream.Send(msg); sErr != nil {
			return sErr
		}
		reply, rErr := svcCStream.Recv()
		if rErr != nil {
			return rErr
		}
		if sErr := stream.Send(reply); sErr != nil {
			return sErr
		}
	}
	return nil
}
