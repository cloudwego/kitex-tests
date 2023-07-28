// Code generated by Kitex v0.5.2. DO NOT EDIT.

package servicea

import (
	unknown_handler "github.com/cloudwego/kitex-tests/kitex_gen/protobuf/unknown_handler"
	server "github.com/cloudwego/kitex/server"
)

// NewInvoker creates a server.Invoker with the given handler and options.
func NewInvoker(handler unknown_handler.ServiceA, opts ...server.Option) server.Invoker {
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
