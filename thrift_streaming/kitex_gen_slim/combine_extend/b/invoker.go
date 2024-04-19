// Code generated by Kitex v0.8.0. DO NOT EDIT.

package b

import (
	combine_extend "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_slim/combine_extend"
	server "github.com/cloudwego/kitex/server"
)

// NewInvoker creates a server.Invoker with the given handler and options.
func NewInvoker(handler combine_extend.B, opts ...server.Option) server.Invoker {
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
