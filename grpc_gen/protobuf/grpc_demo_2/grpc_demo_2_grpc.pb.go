//
// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: idl/grpc_demo_2.proto

package grpc_demo_2

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ServiceA_CallUnary_FullMethodName        = "/grpc_demo_2.ServiceA/CallUnary"
	ServiceA_CallClientStream_FullMethodName = "/grpc_demo_2.ServiceA/CallClientStream"
	ServiceA_CallServerStream_FullMethodName = "/grpc_demo_2.ServiceA/CallServerStream"
	ServiceA_CallBidiStream_FullMethodName   = "/grpc_demo_2.ServiceA/CallBidiStream"
)

// ServiceAClient is the client API for ServiceA service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServiceAClient interface {
	CallUnary(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Reply, error)
	CallClientStream(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[Request, Reply], error)
	CallServerStream(ctx context.Context, in *Request, opts ...grpc.CallOption) (grpc.ServerStreamingClient[Reply], error)
	CallBidiStream(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[Request, Reply], error)
}

type serviceAClient struct {
	cc grpc.ClientConnInterface
}

func NewServiceAClient(cc grpc.ClientConnInterface) ServiceAClient {
	return &serviceAClient{cc}
}

func (c *serviceAClient) CallUnary(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Reply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Reply)
	err := c.cc.Invoke(ctx, ServiceA_CallUnary_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceAClient) CallClientStream(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[Request, Reply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &ServiceA_ServiceDesc.Streams[0], ServiceA_CallClientStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[Request, Reply]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ServiceA_CallClientStreamClient = grpc.ClientStreamingClient[Request, Reply]

func (c *serviceAClient) CallServerStream(ctx context.Context, in *Request, opts ...grpc.CallOption) (grpc.ServerStreamingClient[Reply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &ServiceA_ServiceDesc.Streams[1], ServiceA_CallServerStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[Request, Reply]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ServiceA_CallServerStreamClient = grpc.ServerStreamingClient[Reply]

func (c *serviceAClient) CallBidiStream(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[Request, Reply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &ServiceA_ServiceDesc.Streams[2], ServiceA_CallBidiStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[Request, Reply]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ServiceA_CallBidiStreamClient = grpc.BidiStreamingClient[Request, Reply]

// ServiceAServer is the server API for ServiceA service.
// All implementations must embed UnimplementedServiceAServer
// for forward compatibility.
type ServiceAServer interface {
	CallUnary(context.Context, *Request) (*Reply, error)
	CallClientStream(grpc.ClientStreamingServer[Request, Reply]) error
	CallServerStream(*Request, grpc.ServerStreamingServer[Reply]) error
	CallBidiStream(grpc.BidiStreamingServer[Request, Reply]) error
	mustEmbedUnimplementedServiceAServer()
}

// UnimplementedServiceAServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedServiceAServer struct{}

func (UnimplementedServiceAServer) CallUnary(context.Context, *Request) (*Reply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CallUnary not implemented")
}
func (UnimplementedServiceAServer) CallClientStream(grpc.ClientStreamingServer[Request, Reply]) error {
	return status.Errorf(codes.Unimplemented, "method CallClientStream not implemented")
}
func (UnimplementedServiceAServer) CallServerStream(*Request, grpc.ServerStreamingServer[Reply]) error {
	return status.Errorf(codes.Unimplemented, "method CallServerStream not implemented")
}
func (UnimplementedServiceAServer) CallBidiStream(grpc.BidiStreamingServer[Request, Reply]) error {
	return status.Errorf(codes.Unimplemented, "method CallBidiStream not implemented")
}
func (UnimplementedServiceAServer) mustEmbedUnimplementedServiceAServer() {}
func (UnimplementedServiceAServer) testEmbeddedByValue()                  {}

// UnsafeServiceAServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServiceAServer will
// result in compilation errors.
type UnsafeServiceAServer interface {
	mustEmbedUnimplementedServiceAServer()
}

func RegisterServiceAServer(s grpc.ServiceRegistrar, srv ServiceAServer) {
	// If the following call pancis, it indicates UnimplementedServiceAServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ServiceA_ServiceDesc, srv)
}

func _ServiceA_CallUnary_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceAServer).CallUnary(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ServiceA_CallUnary_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceAServer).CallUnary(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _ServiceA_CallClientStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ServiceAServer).CallClientStream(&grpc.GenericServerStream[Request, Reply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ServiceA_CallClientStreamServer = grpc.ClientStreamingServer[Request, Reply]

func _ServiceA_CallServerStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Request)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ServiceAServer).CallServerStream(m, &grpc.GenericServerStream[Request, Reply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ServiceA_CallServerStreamServer = grpc.ServerStreamingServer[Reply]

func _ServiceA_CallBidiStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ServiceAServer).CallBidiStream(&grpc.GenericServerStream[Request, Reply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ServiceA_CallBidiStreamServer = grpc.BidiStreamingServer[Request, Reply]

// ServiceA_ServiceDesc is the grpc.ServiceDesc for ServiceA service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ServiceA_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc_demo_2.ServiceA",
	HandlerType: (*ServiceAServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CallUnary",
			Handler:    _ServiceA_CallUnary_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CallClientStream",
			Handler:       _ServiceA_CallClientStream_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "CallServerStream",
			Handler:       _ServiceA_CallServerStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "CallBidiStream",
			Handler:       _ServiceA_CallBidiStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "idl/grpc_demo_2.proto",
}

// ServiceBClient is the client API for ServiceB service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServiceBClient interface {
}

type serviceBClient struct {
	cc grpc.ClientConnInterface
}

func NewServiceBClient(cc grpc.ClientConnInterface) ServiceBClient {
	return &serviceBClient{cc}
}

// ServiceBServer is the server API for ServiceB service.
// All implementations must embed UnimplementedServiceBServer
// for forward compatibility.
type ServiceBServer interface {
	mustEmbedUnimplementedServiceBServer()
}

// UnimplementedServiceBServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedServiceBServer struct{}

func (UnimplementedServiceBServer) mustEmbedUnimplementedServiceBServer() {}
func (UnimplementedServiceBServer) testEmbeddedByValue()                  {}

// UnsafeServiceBServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServiceBServer will
// result in compilation errors.
type UnsafeServiceBServer interface {
	mustEmbedUnimplementedServiceBServer()
}

func RegisterServiceBServer(s grpc.ServiceRegistrar, srv ServiceBServer) {
	// If the following call pancis, it indicates UnimplementedServiceBServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ServiceB_ServiceDesc, srv)
}

// ServiceB_ServiceDesc is the grpc.ServiceDesc for ServiceB service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ServiceB_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc_demo_2.ServiceB",
	HandlerType: (*ServiceBServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "idl/grpc_demo_2.proto",
}
