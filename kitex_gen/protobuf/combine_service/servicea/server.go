// Code generated by Kitex v0.12.3. DO NOT EDIT.
package servicea

import (
	combine_service "github.com/cloudwego/kitex-tests/kitex_gen/protobuf/combine_service"
	server "github.com/cloudwego/kitex/server"
)

// NewServer creates a server.Server with the given handler and options.
func NewServer(handler combine_service.ServiceA, opts ...server.Option) server.Server {
	var options []server.Option

	options = append(options, opts...)

	svr := server.NewServer(options...)
	if err := svr.RegisterService(serviceInfo(), handler); err != nil {
		panic(err)
	}
	return svr
}

func RegisterService(svr server.Server, handler combine_service.ServiceA, opts ...server.RegisterOption) error {
	return svr.RegisterService(serviceInfo(), handler, opts...)
}
