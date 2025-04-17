// Code generated by Kitex v0.12.3. DO NOT EDIT.

package streamonlyservicechildchild

import (
	"context"
	echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_slim/echo"
	client "github.com/cloudwego/kitex/client"
	streamcall "github.com/cloudwego/kitex/client/callopt/streamcall"
	streamclient "github.com/cloudwego/kitex/client/streamclient"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
	transport "github.com/cloudwego/kitex/transport"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
}

// StreamClient is designed to provide Interface for Streaming APIs.
type StreamClient interface {
	EchoBidirectionalNew(ctx context.Context, callOptions ...streamcall.Option) (stream StreamOnlyService_EchoBidirectionalNewClient, err error)
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
	return &kStreamOnlyServiceChildChildClient{
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

type kStreamOnlyServiceChildChildClient struct {
	*kClient
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
	return &kStreamOnlyServiceChildChildStreamClient{
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

type kStreamOnlyServiceChildChildStreamClient struct {
	*kClient
}

func (p *kStreamOnlyServiceChildChildStreamClient) EchoBidirectionalNew(ctx context.Context, callOptions ...streamcall.Option) (stream StreamOnlyService_EchoBidirectionalNewClient, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, streamcall.GetCallOptions(callOptions))
	return p.kClient.EchoBidirectionalNew(ctx)
}
