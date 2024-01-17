package servicea

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/consts"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/transport"
)

func myMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		err = next(ctx, req, resp)
		return err
	}
}

func InitServiceAClient() (servicea.Client, error) {
	cli, err := servicea.NewClient("serviceB",
		client.WithRPCTimeout(1000*time.Millisecond),
		client.WithHostPorts(consts.ServiceBAddr), client.WithTransportProtocol(transport.GRPC),
		client.WithMiddleware(myMiddleware))
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
	stream1, err := cli.CallClientStream(context.Background())
	if err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		if err := stream1.Send(req); err != nil {
			return err
		}
	}
	_, err = stream1.CloseAndRecv()
	return err
}

func SendServerStreaming(cli servicea.Client) error {
	// case 3: server streaming
	req := &grpc_demo.Request{Name: "service_a_CallServerStream"}
	stream2, err := cli.CallServerStream(context.Background(), req)
	if err != nil {
		return err
	}
	for {
		reply, err := stream2.Recv()
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
	stream3, err := cli.CallBidiStream(context.Background())
	if err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		if err := stream3.Send(req); err != nil {
			return err
		}
	}
	stream3.Close()
	for {
		reply, err := stream3.Recv()
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
