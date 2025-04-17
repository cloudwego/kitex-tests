// Code generated by Kitex v0.12.3. DO NOT EDIT.

package abcservice

import (
	"context"
	c "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/a/b/c"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
	streamcall "github.com/cloudwego/kitex/client/callopt/streamcall"
	streamclient "github.com/cloudwego/kitex/client/streamclient"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
	transport "github.com/cloudwego/kitex/transport"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	Echo(ctx context.Context, req1 *c.Request, req2 *c.Request, callOptions ...callopt.Option) (r *c.Response, err error)
}

// StreamClient is designed to provide Interface for Streaming APIs.
type StreamClient interface {
	EchoBidirectional(ctx context.Context, callOptions ...streamcall.Option) (stream ABCService_EchoBidirectionalClient, err error)
	EchoServer(ctx context.Context, req1 *c.Request, callOptions ...streamcall.Option) (stream ABCService_EchoServerClient, err error)
	EchoClient(ctx context.Context, callOptions ...streamcall.Option) (stream ABCService_EchoClientClient, err error)
	EchoUnary(ctx context.Context, req1 *c.Request, callOptions ...streamcall.Option) (r *c.Response, err error)
}

type ABCService_EchoBidirectionalClient interface {
	streaming.Stream
	Send(*c.Request) error
	Recv() (*c.Response, error)
}

type ABCService_EchoServerClient interface {
	streaming.Stream
	Recv() (*c.Response, error)
}

type ABCService_EchoClientClient interface {
	streaming.Stream
	Send(*c.Request) error
	CloseAndRecv() (*c.Response, error)
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
	return &kABCServiceClient{
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

type kABCServiceClient struct {
	*kClient
}

func (p *kABCServiceClient) Echo(ctx context.Context, req1 *c.Request, req2 *c.Request, callOptions ...callopt.Option) (r *c.Response, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.Echo(ctx, req1, req2)
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
	return &kABCServiceStreamClient{
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

type kABCServiceStreamClient struct {
	*kClient
}

func (p *kABCServiceStreamClient) EchoBidirectional(ctx context.Context, callOptions ...streamcall.Option) (stream ABCService_EchoBidirectionalClient, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, streamcall.GetCallOptions(callOptions))
	return p.kClient.EchoBidirectional(ctx)
}

func (p *kABCServiceStreamClient) EchoServer(ctx context.Context, req1 *c.Request, callOptions ...streamcall.Option) (stream ABCService_EchoServerClient, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, streamcall.GetCallOptions(callOptions))
	return p.kClient.EchoServer(ctx, req1)
}

func (p *kABCServiceStreamClient) EchoClient(ctx context.Context, callOptions ...streamcall.Option) (stream ABCService_EchoClientClient, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, streamcall.GetCallOptions(callOptions))
	return p.kClient.EchoClient(ctx)
}

func (p *kABCServiceStreamClient) EchoUnary(ctx context.Context, req1 *c.Request, callOptions ...streamcall.Option) (r *c.Response, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, streamcall.GetCallOptions(callOptions))
	return p.kClient.EchoUnary(ctx, req1)
}
