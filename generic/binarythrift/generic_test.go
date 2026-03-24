/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"context"
	"io"
	"net"
	"strconv"
	"testing"

	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/client/callopt/streamcall"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/server/genericserver"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
)

var (
	genCodeAddr, genericAddr net.Addr
)

func TestMain(m *testing.M) {
	genCodeLn := serverutils.Listen()
	genericLn := serverutils.Listen()

	genCodeAddr = genCodeLn.Addr()
	genericAddr = genericLn.Addr()

	newMockTestServer(&serviceImpl{}, genCodeLn)
	newGenericServer(&genericserver.UnknownServiceOrMethodHandler{
		DefaultHandler: pingPongUnknownHandler, StreamingHandler: streamingUnknownHandler}, genericLn)

	m.Run()
}

func TestGenericCall(t *testing.T) {
	addresses := []net.Addr{genericAddr, genCodeAddr}
	protocols := []transport.Protocol{transport.Framed, transport.TTHeader, transport.GRPC, transport.GRPCStreaming | transport.TTHeader}
	for i, addr := range addresses {
		for _, protocol := range protocols {
			t.Run(strconv.Itoa(i)+protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryThriftGenericV2(serviceName),
					addr.String(), client.WithTransportProtocol(protocol))

				args := tenant.NewEchoServiceEchoArgs()
				args.Request = &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(args)
				res, err := genericClient.GenericCall(context.Background(), "Echo", buf)
				test.Assert(t, err == nil)

				resp := tenant.NewEchoServiceEchoResult()
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil)
				test.Assert(t, resp.Success.Msg == "hello world")
			})
		}
	}
}

func TestClientStreaming(t *testing.T) {
	addresses := []net.Addr{genericAddr, genCodeAddr}
	protocols := []transport.Protocol{transport.TTHeaderStreaming, transport.GRPCStreaming, transport.GRPC}
	for i, addr := range addresses {
		for _, protocol := range protocols {
			t.Run(strconv.Itoa(i)+protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryThriftGenericV2(serviceName),
					addr.String(), client.WithTransportProtocol(protocol))
				req := &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(req)
				stream, err := genericClient.ClientStreaming(context.Background(), "EchoClient")
				test.Assert(t, err == nil)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil)

				res, err := stream.CloseAndRecv(stream.Context())
				test.Assert(t, err == nil)

				resp := &tenant.EchoResponse{}
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil)
				test.Assert(t, resp.Msg == "hello world")
			})
		}
	}
}

func TestServerStreaming(t *testing.T) {
	addresses := []net.Addr{genericAddr, genCodeAddr}
	protocols := []transport.Protocol{transport.TTHeaderStreaming, transport.GRPCStreaming, transport.GRPC}
	for i, addr := range addresses {
		for _, protocol := range protocols {
			t.Run(strconv.Itoa(i)+protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryThriftGenericV2(serviceName),
					addr.String(), client.WithTransportProtocol(protocol))
				req := &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(req)
				stream, err := genericClient.ServerStreaming(context.Background(), "EchoServer", buf)
				test.Assert(t, err == nil)

				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil)

				resp := &tenant.EchoResponse{}
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil)
				test.Assert(t, resp.Msg == "hello world")

				_, err = stream.Recv(stream.Context())
				test.Assert(t, err == io.EOF)
			})
		}
	}
}

func TestBidiStreaming(t *testing.T) {
	addresses := []net.Addr{genericAddr, genCodeAddr}
	protocols := []transport.Protocol{transport.TTHeaderStreaming, transport.GRPCStreaming, transport.GRPC}
	for i, addr := range addresses {
		for _, protocol := range protocols {
			t.Run(strconv.Itoa(i)+protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryThriftGenericV2(serviceName),
					addr.String(), client.WithTransportProtocol(protocol))
				req := &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(req)
				stream, err := genericClient.BidirectionalStreaming(context.Background(), "EchoBidi")
				test.Assert(t, err == nil)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil)

				err = stream.CloseSend(stream.Context())
				test.Assert(t, err == nil)

				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil)

				resp := &tenant.EchoResponse{}
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil)
				test.Assert(t, resp.Msg == "hello world")

				_, err = stream.Recv(stream.Context())
				test.Assert(t, err == io.EOF)
			})
		}
	}
}

func TestGenericCallWithIDLService(t *testing.T) {
	protocols := []transport.Protocol{transport.TTHeader, transport.GRPC, transport.GRPCStreaming | transport.TTHeader, transport.TTHeaderStreaming | transport.TTHeader}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			// client created with an invalid IDL service name
			genericClient := newGenericClient(generic.BinaryThriftGenericV2("ClientSideIDLService"),
				genericAddr.String(), client.WithTransportProtocol(protocol))

			args := tenant.NewEchoServiceEchoArgs()
			args.Request = &tenant.EchoRequest{Msg: "hello world"}
			buf := thrift.FastMarshal(args)

			// without WithBinaryGenericIDLService: service name is "ClientSideIDLService", should fail
			_, err := genericClient.GenericCall(context.Background(), "Echo", buf)
			test.Assert(t, err != nil)

			// with WithBinaryGenericIDLService: service name overridden to "EchoService", should succeed
			res, err := genericClient.GenericCall(context.Background(), "Echo", buf,
				callopt.WithBinaryGenericIDLService(serviceName))
			test.Assert(t, err == nil, err)

			resp := tenant.NewEchoServiceEchoResult()
			err = thrift.FastUnmarshal(res.([]byte), resp)
			test.Assert(t, err == nil)
			test.Assert(t, resp.Success.Msg == "hello world", resp)
		})
	}
}

func TestStreamingWithIDLService(t *testing.T) {
	protocols := []transport.Protocol{transport.TTHeaderStreaming | transport.TTHeader, transport.GRPCStreaming | transport.TTHeader, transport.GRPC}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			// client created with an invalid IDL service name
			genericClient := newGenericClient(generic.BinaryThriftGenericV2("ClientSideIDLService"),
				genericAddr.String(), client.WithTransportProtocol(protocol))

			t.Run("BidiStreamingWithoutSpecifying", func(t *testing.T) {
				stream, err := genericClient.BidirectionalStreaming(context.Background(), "EchoBidi")
				test.Assert(t, err == nil, err)
				req := &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(req)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil, err)
				err = stream.CloseSend(stream.Context())
				test.Assert(t, err == nil, err)
				_, err = stream.Recv(stream.Context())
				test.Assert(t, err != nil)
			})
			t.Run("BidiStreamingWithSpecifying", func(t *testing.T) {
				stream, err := genericClient.BidirectionalStreaming(context.Background(), "EchoBidi",
					streamcall.WithBinaryGenericIDLService(serviceName))
				test.Assert(t, err == nil, err)
				req := &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(req)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil, err)
				err = stream.CloseSend(stream.Context())
				test.Assert(t, err == nil, err)
				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil, err)
				resp := &tenant.EchoResponse{}
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Msg == "hello world", resp)
				_, err = stream.Recv(stream.Context())
				test.Assert(t, err == io.EOF, err)
			})
			t.Run("ClientStreamingWithoutSpecifying", func(t *testing.T) {
				stream, err := genericClient.ClientStreaming(context.Background(), "EchoClient")
				test.Assert(t, err == nil, err)
				req := &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(req)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil, err)
				_, err = stream.CloseAndRecv(stream.Context())
				test.Assert(t, err != nil)
			})
			t.Run("ClientStreamingWithSpecifying", func(t *testing.T) {
				stream, err := genericClient.ClientStreaming(context.Background(), "EchoClient",
					streamcall.WithBinaryGenericIDLService(serviceName))
				test.Assert(t, err == nil, err)
				req := &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(req)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil, err)
				res, err := stream.CloseAndRecv(stream.Context())
				test.Assert(t, err == nil, err)
				resp := &tenant.EchoResponse{}
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Msg == "hello world", resp)
			})
			t.Run("ServerStreamingWithoutSpecifying", func(t *testing.T) {
				req := &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(req)
				stream, err := genericClient.ServerStreaming(context.Background(), "EchoServer", buf)
				test.Assert(t, err == nil, err)
				_, err = stream.Recv(stream.Context())
				test.Assert(t, err != nil)
			})
			t.Run("ServerStreamingWithSpecifying", func(t *testing.T) {
				req := &tenant.EchoRequest{Msg: "hello world"}
				buf := thrift.FastMarshal(req)
				stream, err := genericClient.ServerStreaming(context.Background(), "EchoServer", buf,
					streamcall.WithBinaryGenericIDLService(serviceName))
				test.Assert(t, err == nil, err)
				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil, err)
				resp := &tenant.EchoResponse{}
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Msg == "hello world", resp)
				_, err = stream.Recv(stream.Context())
				test.Assert(t, err == io.EOF, err)
			})
		})
	}
}
