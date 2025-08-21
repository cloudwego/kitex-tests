package test

import (
	"context"
	"io"
	"net"
	"strconv"
	"testing"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/server/genericserver"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi/mock"
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

	newMockTestServer(&StreamingTestImpl{}, genCodeLn)
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
				genericClient := newGenericClient(generic.BinaryPbGeneric(serviceName, packageName),
					addr.String(), client.WithTransportProtocol(protocol))

				args := mock.UnaryTestArgs{Req: &pbapi.MockReq{Message: "hello world"}}
				buf, _ := args.Marshal(nil)
				res, err := genericClient.GenericCall(context.Background(), "UnaryTest", buf)
				test.Assert(t, err == nil)

				resp := &mock.UnaryTestResult{}
				err = resp.Unmarshal(res.([]byte))
				test.Assert(t, err == nil)
				test.Assert(t, resp.Success.Message == "hello world")
			})
		}
	}
}

func TestClientStreaming(t *testing.T) {
	addresses := []net.Addr{genericAddr, genCodeAddr}
	protocols := []transport.Protocol{transport.GRPCStreaming, transport.GRPC}
	for i, addr := range addresses {
		for _, protocol := range protocols {
			t.Run(strconv.Itoa(i)+protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryPbGeneric(serviceName, packageName),
					addr.String(), client.WithTransportProtocol(protocol))
				req := &pbapi.MockReq{Message: "hello world"}
				buf, _ := req.Marshal(nil)
				stream, err := genericClient.ClientStreaming(context.Background(), "ClientStreamingTest")
				test.Assert(t, err == nil)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil)

				res, err := stream.CloseAndRecv(stream.Context())
				test.Assert(t, err == nil)

				resp := &pbapi.MockResp{}
				err = resp.Unmarshal(res.([]byte))
				test.Assert(t, err == nil)
				test.Assert(t, resp.Message == "hello world")
			})
		}
	}
}

func TestServerStreaming(t *testing.T) {
	addresses := []net.Addr{genericAddr, genCodeAddr}
	protocols := []transport.Protocol{transport.GRPCStreaming, transport.GRPC}
	for i, addr := range addresses {
		for _, protocol := range protocols {
			t.Run(strconv.Itoa(i)+protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryPbGeneric(serviceName, packageName),
					addr.String(), client.WithTransportProtocol(protocol))
				req := &pbapi.MockReq{Message: "hello world"}
				buf, _ := req.Marshal(nil)
				stream, err := genericClient.ServerStreaming(context.Background(), "ServerStreamingTest", buf)
				test.Assert(t, err == nil)

				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil)

				resp := &pbapi.MockResp{}
				err = resp.Unmarshal(res.([]byte))
				test.Assert(t, err == nil)
				test.Assert(t, resp.Message == "hello world")

				_, err = stream.Recv(stream.Context())
				test.Assert(t, err == io.EOF)
			})
		}
	}
}

func TestBidiStreaming(t *testing.T) {
	addresses := []net.Addr{genericAddr, genCodeAddr}
	protocols := []transport.Protocol{transport.GRPCStreaming, transport.GRPC}
	for i, addr := range addresses {
		for _, protocol := range protocols {
			t.Run(strconv.Itoa(i)+protocol.String(), func(t *testing.T) {
				genericClient := newGenericClient(generic.BinaryPbGeneric(serviceName, packageName),
					addr.String(), client.WithTransportProtocol(protocol))
				req := &pbapi.MockReq{Message: "hello world"}
				buf, _ := req.Marshal(nil)
				stream, err := genericClient.BidirectionalStreaming(context.Background(), "BidirectionalStreamingTest")
				test.Assert(t, err == nil)
				err = stream.Send(stream.Context(), buf)
				test.Assert(t, err == nil)

				err = stream.CloseSend(stream.Context())
				test.Assert(t, err == nil)

				res, err := stream.Recv(stream.Context())
				test.Assert(t, err == nil)

				resp := &pbapi.MockResp{}
				err = resp.Unmarshal(res.([]byte))
				test.Assert(t, err == nil)
				test.Assert(t, resp.Message == "hello world")

				_, err = stream.Recv(stream.Context())
				test.Assert(t, err == io.EOF)
			})
		}
	}
}

func TestClientStreamingV1(t *testing.T) {
	genericClient := newGenericClient(generic.BinaryPbGeneric(serviceName, packageName),
		genCodeAddr.String(), client.WithTransportProtocol(transport.GRPC))
	req := &pbapi.MockReq{Message: "hello world"}
	buf, _ := req.Marshal(nil)
	stream, err := genericclient.NewClientStreaming(context.Background(), genericClient, "ClientStreamingTest")
	test.Assert(t, err == nil)
	err = stream.Send(buf)
	test.Assert(t, err == nil)

	res, err := stream.CloseAndRecv()
	test.Assert(t, err == nil)

	resp := &pbapi.MockResp{}
	err = resp.Unmarshal(res.([]byte))
	test.Assert(t, err == nil)
	test.Assert(t, resp.Message == "hello world")
}

func TestServerStreamingV1(t *testing.T) {
	genericClient := newGenericClient(generic.BinaryPbGeneric(serviceName, packageName),
		genCodeAddr.String(), client.WithTransportProtocol(transport.GRPC))
	req := &pbapi.MockReq{Message: "hello world"}
	buf, _ := req.Marshal(nil)
	stream, err := genericclient.NewServerStreaming(context.Background(), genericClient, "ServerStreamingTest", buf)
	test.Assert(t, err == nil)

	res, err := stream.Recv()
	test.Assert(t, err == nil)

	resp := &pbapi.MockResp{}
	err = resp.Unmarshal(res.([]byte))
	test.Assert(t, err == nil)
	test.Assert(t, resp.Message == "hello world")
}

func TestBidiStreamingV1(t *testing.T) {
	genericClient := newGenericClient(generic.BinaryPbGeneric(serviceName, packageName),
		genCodeAddr.String(), client.WithTransportProtocol(transport.GRPC))
	req := &pbapi.MockReq{Message: "hello world"}
	buf, _ := req.Marshal(nil)
	stream, err := genericclient.NewBidirectionalStreaming(context.Background(), genericClient, "BidirectionalStreamingTest")
	test.Assert(t, err == nil)
	err = stream.Send(buf)
	test.Assert(t, err == nil)

	res, err := stream.Recv()
	test.Assert(t, err == nil)

	resp := &pbapi.MockResp{}
	err = resp.Unmarshal(res.([]byte))
	test.Assert(t, err == nil)
	test.Assert(t, resp.Message == "hello world")
}
