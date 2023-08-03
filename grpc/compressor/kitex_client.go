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
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/transport"
	"io"
	"strconv"
)

type ClientWrapper struct {
	client servicea.Client
}

func GetClient(hostport string, clientOptions ...client.Option) (*ClientWrapper, error) {
	opts := append(clientOptions, client.WithTransportProtocol(transport.GRPC), client.WithHostPorts(hostport))
	client, err := servicea.NewClient("kitex-test", opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWrapper{client: client}, nil
}

func (c *ClientWrapper) RunUnary(callOptions ...callopt.Option) (*grpc_demo.Reply, error) {
	req := &grpc_demo.Request{Name: "Kitex"}
	ctx := context.Background()
	return c.client.CallUnary(ctx, req, callOptions...)
}

func (c *ClientWrapper) RunClientStream(callOptions ...callopt.Option) (*grpc_demo.Reply, error) {

	ctx := context.Background()
	streamCli, err := c.client.CallClientStream(ctx)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 3; i++ {
		req := &grpc_demo.Request{Name: "kitex-" + strconv.Itoa(i)}
		err = streamCli.SendMsg(req)
		if err != nil {
			return nil, err
		}
	}
	return streamCli.CloseAndRecv()
}

func (c *ClientWrapper) RunServerStream(callOptions ...callopt.Option) ([]*grpc_demo.Reply, error) {
	ctx := context.Background()
	req := &grpc_demo.Request{Name: "kitex"}
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

func (c *ClientWrapper) RunBidiStream(callOptions ...callopt.Option) ([]*grpc_demo.Reply, error) {

	ctx := context.Background()
	stream, err := c.client.CallBidiStream(ctx)
	if err != nil {
		return nil, err
	}
	errChan := make(chan error)
	go func() {
		for i := 0; i < 3; i++ {
			req := &grpc_demo.Request{Name: "kitex-" + strconv.Itoa(i)}
			err = stream.Send(req)
			if err != nil {
				errChan <- err
				return
			}
		}
		er := stream.Close()
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
