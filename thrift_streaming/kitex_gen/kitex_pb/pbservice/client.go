// Code generated by Kitex v0.9.1. DO NOT EDIT.

package pbservice

import (
	"context"
	kitex_pb "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	EchoPingPong(ctx context.Context, Req *kitex_pb.Request, callOptions ...callopt.Option) (r *kitex_pb.Response, err error)
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
	return &kPBServiceClient{
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

type kPBServiceClient struct {
	*kClient
}

func (p *kPBServiceClient) EchoPingPong(ctx context.Context, Req *kitex_pb.Request, callOptions ...callopt.Option) (r *kitex_pb.Response, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.EchoPingPong(ctx, Req)
}
