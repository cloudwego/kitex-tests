// Code generated by Kitex v0.12.3. DO NOT EDIT.

package servicea

import (
	"context"
	multi_service "github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	Echo1(ctx context.Context, req *multi_service.Request, callOptions ...callopt.Option) (r *multi_service.Response, err error)
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
	return &kServiceAClient{
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

type kServiceAClient struct {
	*kClient
}

func (p *kServiceAClient) Echo1(ctx context.Context, req *multi_service.Request, callOptions ...callopt.Option) (r *multi_service.Response, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.Echo1(ctx, req)
}
