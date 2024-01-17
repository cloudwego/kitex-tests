// Code generated by Kitex v0.8.0. DO NOT EDIT.

package combineservice

import (
	"context"
	echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	EchoPingPongNew(ctx context.Context, req1 *echo.EchoRequest, callOptions ...callopt.Option) (r *echo.EchoResponse, err error)
}

type StreamOnlyService_EchoBidirectionalNewClient interface {
	streaming.Stream
	Send(*echo.EchoRequest) error
	Recv() (*echo.EchoResponse, error)
}

// NewClient creates a client for the service defined in IDL.
func NewClient(destService string, opts ...client.Option) (Client, error) {
	var options []client.Option
	options = append(options, client.WithDestService(destService))

	options = append(options, opts...)

	kc, err := client.NewClient(serviceInfoForClient(), options...)
	if err != nil {
		return nil, err
	}
	return &kCombineServiceClient{
		kClient: newServiceClient(kc),
	}, nil
}

// MustNewClient creates a client for the service defined in IDL. It panics if any error occurs.
func MustNewClient(destService string, opts ...client.Option) Client {
	kc, err := NewClient(destService, opts...)
	if err != nil {
		panic(err)
	}
	return kc
}

type kCombineServiceClient struct {
	*kClient
}

func (p *kCombineServiceClient) EchoPingPongNew(ctx context.Context, req1 *echo.EchoRequest, callOptions ...callopt.Option) (r *echo.EchoResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.EchoPingPongNew(ctx, req1)
}
