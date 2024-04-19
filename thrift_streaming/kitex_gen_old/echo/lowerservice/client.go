// Code generated by Kitex v0.8.0. DO NOT EDIT.

package lowerservice

import (
	"context"
	echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_old/echo"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	EchoBidirectional(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error)
	EchoClient(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error)
	EchoServer(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error)
	EchoUnary(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error)
	EchoPingPong(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error)
}

// NewClient creates a client for the service defined in IDL.
func NewClient(destService string, opts ...client.Option) (Client, error) {
	var options []client.Option
	options = append(options, client.WithDestService(destService))

	options = append(options, opts...)

	kc, err := client.NewClient(serviceInfo(), options...)
	if err != nil {
		return nil, err
	}
	return &kLowerServiceClient{
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

type kLowerServiceClient struct {
	*kClient
}

func (p *kLowerServiceClient) EchoBidirectional(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.EchoBidirectional(ctx, req1)
}

func (p *kLowerServiceClient) EchoClient(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.EchoClient(ctx, req1)
}

func (p *kLowerServiceClient) EchoServer(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.EchoServer(ctx, req1)
}

func (p *kLowerServiceClient) EchoUnary(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.EchoUnary(ctx, req1)
}

func (p *kLowerServiceClient) EchoPingPong(ctx context.Context, req1 *echo.LowerRequest, callOptions ...callopt.Option) (r *echo.LowerResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.EchoPingPong(ctx, req1)
}
