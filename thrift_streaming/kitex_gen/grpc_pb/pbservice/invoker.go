// Code generated by Kitex v0.9.1. DO NOT EDIT.

package pbservice

import (
	grpc_pb "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb"
	server "github.com/cloudwego/kitex/server"
)

// NewInvoker creates a server.Invoker with the given handler and options.
func NewInvoker(handler grpc_pb.PBService, opts ...server.Option) server.Invoker {
	var options []server.Option

	options = append(options, opts...)

	s := server.NewInvoker(options...)
	if err := s.RegisterService(serviceInfo(), handler); err != nil {
		panic(err)
	}
	if err := s.Init(); err != nil {
		panic(err)
	}
	return s
}
