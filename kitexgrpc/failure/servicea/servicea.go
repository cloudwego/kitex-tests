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

package servicea

import (
	"context"
	"fmt"
	"github.com/cloudwego/kitex-tests/kitexgrpc/failure/common"
	"github.com/cloudwego/kitex/pkg/streaming"
	"io"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/consts"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/transport"
)

func ServiceAMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		err = next(ctx, req, resp)
		return err
	}
}

func InitServiceAClient() (servicea.Client, error) {
	cli, err := servicea.NewClient("serviceB",
		client.WithRPCTimeout(1000*time.Millisecond),
		client.WithHostPorts(consts.ServiceBAddr), client.WithTransportProtocol(transport.GRPC),
		client.WithMiddleware(ServiceAMiddleware))
	return cli, err
}

func SendUnary(cli servicea.Client) error {
	// case 1: unary
	req := &grpc_demo.Request{Name: "service_a_CallUnary"}
	resp, err := cli.CallUnary(context.Background(), req)
	fmt.Println(resp, err)
	return err
}

func SendClientStreaming(cli servicea.Client) error {
	// case 2: client streaming
	req := &grpc_demo.Request{Name: "service_a_CallClientStream"}
	s, err := cli.CallClientStream(context.Background())
	if err != nil {
		return err
	}

	// use wrapStream to request
	ws := common.WrapClientStream[grpc_demo.Request, grpc_demo.Reply]{
		Mode: common.StreamingClient,
		CSC:  s,
	}

	for i := 0; i < 3; i++ {
		if err := ws.Send(context.Background(), req); err != nil {
			return err
		}
	}
	_, err = ws.CloseAndRecv(context.Background())
	return err
}

func SendServerStreaming(cli servicea.Client, injector common.Injector) error {
	// case 3: server streaming
	req := &grpc_demo.Request{Name: "service_a_CallServerStream"}
	s, err := cli.CallServerStream(context.Background(), req)
	if err != nil {
		return err
	}

	// use wrapStream to request
	ws := common.WrapClientStream[grpc_demo.Request, grpc_demo.Reply]{
		Mode:     common.StreamingServer,
		SSC:      s,
		Injector: injector,
	}

	ctx := streaming.NewCtxWithStream(s.Context(), s)
	for {
		reply, err := ws.Recv(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Println(reply)
	}
	return err
}

func SendBidiStreaming(cli servicea.Client) error {
	// case 4: bidi streaming
	req := &grpc_demo.Request{Name: "service_a_CallBidiStream"}
	s, err := cli.CallBidiStream(context.Background())
	if err != nil {
		return err
	}
	ws := common.WrapClientStream[grpc_demo.Request, grpc_demo.Reply]{
		Mode: common.StreamingBidirectional,
		BSC:  s,
	}

	for i := 0; i < 3; i++ {
		if err := ws.Send(context.Background(), req); err != nil {
			return err
		}
	}
	ws.Close(context.Background())
	for {
		reply, err := ws.Recv(context.Background())
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Println("serviceA", reply)
	}
	return err
}
