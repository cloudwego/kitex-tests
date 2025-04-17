// Code generated by Kitex v0.12.3. DO NOT EDIT.
package lowerservice

import (
	echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_slim/echo"
	server "github.com/cloudwego/kitex/server"
)

// NewServer creates a server.Server with the given handler and options.
func NewServer(handler echo.LowerService, opts ...server.Option) server.Server {
	var options []server.Option

	options = append(options, opts...)
	options = append(options, server.WithCompatibleMiddlewareForUnary())

	svr := server.NewServer(options...)
	if err := svr.RegisterService(serviceInfo(), handler); err != nil {
		panic(err)
	}
	return svr
}

func RegisterService(svr server.Server, handler echo.LowerService, opts ...server.RegisterOption) error {
	return svr.RegisterService(serviceInfo(), handler, opts...)
}
