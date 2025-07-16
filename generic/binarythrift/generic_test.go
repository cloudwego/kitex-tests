package test

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/server/genericserver"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/streamx/kitex_gen/echo"
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
		PingPongHandler: pingPongUnknownHandler, StreamingHandler: streamingUnknownHandler}, genericLn)

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

				args := echo.NewTestServicePingPongArgs()
				args.Req = &echo.EchoClientRequest{Message: "hello world"}
				buf := thrift.FastMarshal(args)
				res, err := genericClient.GenericCall(context.Background(), "PingPong", buf)
				test.Assert(t, err == nil)

				resp := echo.NewTestServicePingPongResult()
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil)
				test.Assert(t, resp.Success.Message == "hello world")
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
				req := &echo.EchoClientRequest{Message: "hello world"}
				buf := thrift.FastMarshal(req)
				stream, err := genericClient.ClientStreaming(context.Background(), "EchoClient")
				test.Assert(t, err == nil)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil)

				res, err := stream.CloseAndRecv(stream.Context())
				test.Assert(t, err == nil)

				resp := &echo.EchoClientResponse{}
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil)
				test.Assert(t, resp.Message == "hello world")
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
				req := &echo.EchoClientRequest{Message: "hello world"}
				buf := thrift.FastMarshal(req)
				stream, err := genericClient.ServerStreaming(context.Background(), "EchoServer", buf)
				test.Assert(t, err == nil)

				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil)

				resp := &echo.EchoClientResponse{}
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil)
				test.Assert(t, resp.Message == "hello world")
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
				req := &echo.EchoClientRequest{Message: "hello world"}
				buf := thrift.FastMarshal(req)
				stream, err := genericClient.BidirectionalStreaming(context.Background(), "EchoServer")
				test.Assert(t, err == nil)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil)

				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil)

				resp := &echo.EchoClientResponse{}
				err = thrift.FastUnmarshal(res.([]byte), resp)
				test.Assert(t, err == nil)
				test.Assert(t, resp.Message == "hello world")
			})
		}
	}
}
