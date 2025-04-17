// Code generated by Kitex v0.12.3. DO NOT EDIT.

package serviceb

import (
	"context"
	grpc_multi_service "github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_multi_service"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
	streamcall "github.com/cloudwego/kitex/client/callopt/streamcall"
	streamclient "github.com/cloudwego/kitex/client/streamclient"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
	transport "github.com/cloudwego/kitex/transport"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	EchoB(ctx context.Context, callOptions ...callopt.Option) (stream ServiceB_EchoBClient, err error)
}

// StreamClient is designed to provide Interface for Streaming APIs.
type StreamClient interface {
	EchoB(ctx context.Context, callOptions ...streamcall.Option) (stream ServiceB_EchoBClient, err error)
}

type ServiceB_EchoBClient interface {
	streaming.Stream
	Send(*grpc_multi_service.RequestB) error
	Recv() (*grpc_multi_service.ReplyB, error)
}

// NewClient creates a client for the service defined in IDL.
func NewClient(destService string, opts ...client.Option) (Client, error) {
	var options []client.Option
	options = append(options, client.WithDestService(destService))

	options = append(options, client.WithTransportProtocol(transport.GRPC))

	options = append(options, opts...)

	kc, err := client.NewClient(serviceInfo(), options...)
	if err != nil {
		return nil, err
	}
	return &kServiceBClient{
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

type kServiceBClient struct {
	*kClient
}

func (p *kServiceBClient) EchoB(ctx context.Context, callOptions ...callopt.Option) (stream ServiceB_EchoBClient, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.EchoB(ctx)
}

// NewStreamClient creates a stream client for the service's streaming APIs defined in IDL.
func NewStreamClient(destService string, opts ...streamclient.Option) (StreamClient, error) {
	var options []client.Option
	options = append(options, client.WithDestService(destService))
	options = append(options, client.WithTransportProtocol(transport.GRPC))
	options = append(options, streamclient.GetClientOptions(opts)...)

	kc, err := client.NewClient(serviceInfoForStreamClient(), options...)
	if err != nil {
		return nil, err
	}
	return &kServiceBStreamClient{
		kClient: newServiceClient(kc),
	}, nil
}

// MustNewStreamClient creates a stream client for the service's streaming APIs defined in IDL.
// It panics if any error occurs.
func MustNewStreamClient(destService string, opts ...streamclient.Option) StreamClient {
	kc, err := NewStreamClient(destService, opts...)
	if err != nil {
		panic(err)
	}
	return kc
}

type kServiceBStreamClient struct {
	*kClient
}

func (p *kServiceBStreamClient) EchoB(ctx context.Context, callOptions ...streamcall.Option) (stream ServiceB_EchoBClient, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, streamcall.GetCallOptions(callOptions))
	return p.kClient.EchoB(ctx)
}
