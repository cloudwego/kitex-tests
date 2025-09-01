package test

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
)

var (
	genCodeAddr net.Addr
)

func TestMain(m *testing.M) {
	genCodeLn := serverutils.Listen()

	genCodeAddr = genCodeLn.Addr()

	newMockTestServer(&serviceImpl{}, genCodeLn)

	m.Run()
}

func TestUnknownGenericCall(t *testing.T) {
	protocols := []transport.Protocol{transport.Framed, transport.TTHeader, transport.GRPC, transport.GRPCStreaming | transport.TTHeader}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			genericClient := newGenericClient(generic.BinaryThriftGenericV2(unknownServiceName),
				genCodeAddr.String(), client.WithTransportProtocol(protocol))

			res, err := genericClient.GenericCall(context.Background(), unknownMethodName, []byte(unknownMessage))
			test.Assert(t, err == nil)
			test.Assert(t, string(res.([]byte)) == unknownMessage)
		})
	}
}

func TestUnknownClientStreaming(t *testing.T) {
	protocols := []transport.Protocol{transport.TTHeaderStreaming, transport.GRPCStreaming, transport.GRPC}
	for _, protocol := range protocols {
		for _, svcName := range []string{unknownServiceName, serviceName} {
			t.Run(protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryThriftGenericV2(svcName),
					genCodeAddr.String(), client.WithTransportProtocol(protocol))

				stream, err := genericClient.ClientStreaming(context.Background(), unknownMethodName)
				test.Assert(t, err == nil)
				err = stream.Send(stream.Context(), []byte(unknownMessage))
				test.Assert(t, err == nil)

				res, err := stream.CloseAndRecv(stream.Context())
				test.Assert(t, err == nil)
				test.Assert(t, string(res.([]byte)) == unknownMessage)
			})
		}
	}
}

func TestUnknownServerStreaming(t *testing.T) {
	protocols := []transport.Protocol{transport.TTHeaderStreaming, transport.GRPCStreaming, transport.GRPC}
	for _, protocol := range protocols {
		for _, svcName := range []string{unknownServiceName, serviceName} {
			t.Run(protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryThriftGenericV2(svcName),
					genCodeAddr.String(), client.WithTransportProtocol(protocol))

				stream, err := genericClient.ServerStreaming(context.Background(), unknownMethodName, []byte(unknownMessage))
				test.Assert(t, err == nil)

				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil)
				test.Assert(t, string(res.([]byte)) == unknownMessage)

				_, err = stream.Recv(stream.Context())
				test.Assert(t, err == io.EOF)
			})
		}
	}
}

func TestUnknownBidiStreaming(t *testing.T) {
	protocols := []transport.Protocol{transport.TTHeaderStreaming, transport.GRPCStreaming, transport.GRPC}
	for _, protocol := range protocols {
		for _, svcName := range []string{unknownServiceName, serviceName} {
			t.Run(protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryThriftGenericV2(svcName),
					genCodeAddr.String(), client.WithTransportProtocol(protocol))

				stream, err := genericClient.BidirectionalStreaming(context.Background(), unknownMethodName)
				test.Assert(t, err == nil)
				err = stream.Send(stream.Context(), []byte(unknownMessage))
				test.Assert(t, err == nil)

				err = stream.CloseSend(stream.Context())
				test.Assert(t, err == nil)

				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil)
				test.Assert(t, string(res.([]byte)) == unknownMessage)

				_, err = stream.Recv(stream.Context())
				test.Assert(t, err == io.EOF)
			})
		}
	}
}

func TestGenericCall(t *testing.T) {
	protocols := []transport.Protocol{transport.Framed, transport.TTHeader, transport.GRPC, transport.GRPCStreaming | transport.TTHeader}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			genericClient := newGenericClient(generic.BinaryThriftGenericV2(serviceName),
				genCodeAddr.String(), client.WithTransportProtocol(protocol))

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

func TestClientStreaming(t *testing.T) {
	protocols := []transport.Protocol{transport.TTHeaderStreaming, transport.GRPCStreaming, transport.GRPC}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			genericClient := newGenericClient(generic.BinaryThriftGenericV2(serviceName),
				genCodeAddr.String(), client.WithTransportProtocol(protocol))
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

func TestServerStreaming(t *testing.T) {
	protocols := []transport.Protocol{transport.TTHeaderStreaming, transport.GRPCStreaming, transport.GRPC}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			genericClient := newGenericClient(generic.BinaryThriftGenericV2(serviceName),
				genCodeAddr.String(), client.WithTransportProtocol(protocol))
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

func TestBidiStreaming(t *testing.T) {
	protocols := []transport.Protocol{transport.TTHeaderStreaming, transport.GRPCStreaming, transport.GRPC}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			genericClient := newGenericClient(generic.BinaryThriftGenericV2(serviceName),
				genCodeAddr.String(), client.WithTransportProtocol(protocol))
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
