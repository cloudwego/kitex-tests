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
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex-tests/kitexgrpc/failure/common"
	"github.com/cloudwego/kitex-tests/kitexgrpc/failure/consts"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/transport"
	"io"
	"time"
)

func ServiceAMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		err = next(ctx, req, resp)
		return err
	}
}

func InitServiceAClient(dialer ...remote.Dialer) (servicea.Client, error) {
	opt := []client.Option{
		client.WithRPCTimeout(1000 * time.Millisecond),
		client.WithHostPorts(consts.ServiceBAddr),
		client.WithTransportProtocol(transport.GRPC),
		client.WithMiddleware(ServiceAMiddleware),
	}
	if len(dialer) > 0 {
		opt = append(opt, client.WithDialer(dialer[0]))
	}
	cli, err := servicea.NewClient("serviceB", opt...)
	return cli, err
}

func SendUnary(cli servicea.Client) error {
	// case 1: unary
	req := &grpc_demo.Request{Name: "service_a_CallUnary"}
	resp, err := cli.CallUnary(context.Background(), req)
	fmt.Println(resp, err)
	return err
}

func SendClientStreaming(cli servicea.Client, ctx context.Context, injector common.Injector) (err error) {
	defer func() {
		if err != nil {
			fmt.Printf("[ServiceA] ClientStream, error=%v\n", err)
		}
	}()

	// case 2: client streaming
	req := &grpc_demo.Request{Name: "service_a_CallClientStream"}
	s, err := cli.CallClientStream(ctx)
	if err != nil {
		return err
	}

	// use wrapStream to request
	ws := common.WrapClientStream[grpc_demo.Request, grpc_demo.Reply]{
		Mode:     common.StreamingClient,
		CSC:      s,
		Injector: injector,
	}

	ctx = streaming.NewCtxWithStream(s.Context(), s)
	for i := 0; i < 10; i++ {
		if err := ws.Send(ctx, req); err != nil {
			return err
		}
	}
	ctx = context.WithValue(ctx, "cnt", 1)
	_, err = ws.CloseAndRecv(ctx)
	return err
}

func SendServerStreaming(cli servicea.Client, ctx context.Context, injector common.Injector) (err error) {
	defer func() {
		if err != nil {
			fmt.Printf("[ServiceA] ClientStream error=%v\n", err)
		}
	}()

	// case 3: server streaming
	req := &grpc_demo.Request{Name: "service_a_CallServerStream"}
	s, err := cli.CallServerStream(ctx, req)
	if err != nil {
		return err
	}

	// use wrapStream to request
	ws := common.WrapClientStream[grpc_demo.Request, grpc_demo.Reply]{
		Mode:     common.StreamingServer,
		SSC:      s,
		Injector: injector,
	}

	ctx = streaming.NewCtxWithStream(s.Context(), s)
	for i := 0; ; i++ {
		ctx = context.WithValue(ctx, "cnt", i)
		_, rErr := ws.Recv(ctx)
		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			err = rErr
			return err
		}
	}
	return nil
}

func SendBidiStreaming(cli servicea.Client, ctx context.Context, injector common.Injector) (err error) {
	defer func() {
		if err != nil {
			fmt.Printf("[ServiceA] BidiStream error=%v\n", err)
		}
	}()

	// case 4: bidi streaming
	req := &grpc_demo.Request{Name: "service_a_CallBidiStream"}
	s, err := cli.CallBidiStream(ctx)
	if err != nil {
		return err
	}
	ws := common.WrapClientStream[grpc_demo.Request, grpc_demo.Reply]{
		Mode:     common.StreamingBidirectional,
		BSC:      s,
		Injector: injector,
	}

	for i := 0; i < 20; i++ {
		if err := ws.Send(s.Context(), req); err != nil {
			return err
		}
	}
	//ws.Close(ctx)
	for i := 0; ; i++ {
		_, err := ws.Recv(context.WithValue(s.Context(), "cnt", i))
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return err
}
