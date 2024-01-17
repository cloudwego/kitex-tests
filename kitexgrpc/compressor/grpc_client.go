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
	"io"
	"strconv"

	grpc_demo "github.com/cloudwego/kitex-tests/grpc_gen/protobuf/grpc_demo_2"
	"google.golang.org/grpc"
)

type ClientWrapperGRPC struct {
	client grpc_demo.ServiceAClient
}

func GetGRPCClient(hostport string) (*ClientWrapperGRPC, error) {
	conn, err := grpc.Dial(hostport, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return &ClientWrapperGRPC{
		client: grpc_demo.NewServiceAClient(conn),
	}, nil
}

func (c *ClientWrapperGRPC) RunUnary(opts ...grpc.CallOption) (*grpc_demo.Reply, error) {
	return c.client.CallUnary(context.Background(), &grpc_demo.Request{Name: "Grpc"}, opts...)
}

func (c *ClientWrapperGRPC) RunClientStream(opts ...grpc.CallOption) (*grpc_demo.Reply, error) {
	ctx := context.Background()
	streamCli, err := c.client.CallClientStream(ctx)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 3; i++ {
		req := &grpc_demo.Request{Name: "kitexgrpc-" + strconv.Itoa(i)}
		err = streamCli.SendMsg(req)
		if err != nil {
			return nil, err
		}
	}
	return streamCli.CloseAndRecv()
}

func (c *ClientWrapperGRPC) RunServerStream(opts ...grpc.CallOption) ([]*grpc_demo.Reply, error) {
	ctx := context.Background()
	req := &grpc_demo.Request{Name: "kitexgrpc"}
	streamCli, err := c.client.CallServerStream(ctx, req)
	if err != nil {
		return nil, err
	}
	replies := []*grpc_demo.Reply{}
	for {
		reply, er := streamCli.Recv()
		if er != nil {
			if er != io.EOF {
				return nil, er
			}
			break
		}
		replies = append(replies, reply)
	}
	return replies, nil
}

func (c *ClientWrapperGRPC) RunBidiStream(opts ...grpc.CallOption) ([]*grpc_demo.Reply, error) {

	ctx := context.Background()
	stream, err := c.client.CallBidiStream(ctx)
	if err != nil {
		return nil, err
	}
	errChan := make(chan error)
	go func() {
		for i := 0; i < 3; i++ {
			req := &grpc_demo.Request{Name: "kitexgrpc-" + strconv.Itoa(i)}
			err = stream.Send(req)
			if err != nil {
				errChan <- err
				return
			}
		}
		er := stream.CloseSend()
		if er != nil {
			errChan <- er
			return
		}
		errChan <- nil
	}()
	repliesChan := make(chan []*grpc_demo.Reply)
	errRecvChan := make(chan error)
	go func() {
		replies := []*grpc_demo.Reply{}
		for {
			reply, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errRecvChan <- err
				return
			}
			replies = append(replies, reply)
		}
		repliesChan <- replies
	}()
	err = <-errChan
	if err != nil {
		return nil, err
	}
	select {
	case replies := <-repliesChan:
		return replies, nil
	case err = <-errRecvChan:
		return nil, err
	}
}
