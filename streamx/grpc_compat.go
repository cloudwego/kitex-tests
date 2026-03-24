// Copyright 2026 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package streamx

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/endpoint/sep"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi/mock"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/echoservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
)

type oldStreamMWKey struct{}

// wrappedOldStream simulates a user-wrapped streaming.Stream in old-style endpoint middleware.
type wrappedOldStream struct {
	streaming.Stream
}

type grpcCompatThriftImpl struct {
	tenant.EchoService
}

func (s *grpcCompatThriftImpl) Echo(ctx context.Context, req *tenant.EchoRequest) (*tenant.EchoResponse, error) {
	return &tenant.EchoResponse{Msg: req.Msg}, nil
}

func (s *grpcCompatThriftImpl) EchoBidi(ctx context.Context, stream tenant.EchoService_EchoBidiServer) error {
	if ctx.Value(oldStreamMWKey{}) != true {
		return errors.New("old stream middleware wrapping not preserved in EchoBidi")
	}
	for {
		req, err := stream.Recv(ctx)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err = stream.Send(ctx, &tenant.EchoResponse{Msg: req.Msg}); err != nil {
			return err
		}
	}
}

func (s *grpcCompatThriftImpl) EchoClient(ctx context.Context, stream tenant.EchoService_EchoClientServer) error {
	if ctx.Value(oldStreamMWKey{}) != true {
		return errors.New("old stream middleware wrapping not preserved in EchoClient")
	}
	var lastMsg string
	for {
		req, err := stream.Recv(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		lastMsg = req.Msg
	}
	return stream.SendAndClose(ctx, &tenant.EchoResponse{Msg: lastMsg})
}

func (s *grpcCompatThriftImpl) EchoServer(ctx context.Context, req *tenant.EchoRequest, stream tenant.EchoService_EchoServerServer) error {
	if ctx.Value(oldStreamMWKey{}) != true {
		return errors.New("old stream middleware wrapping not preserved in EchoServer")
	}
	return stream.Send(ctx, &tenant.EchoResponse{Msg: req.Msg})
}

type grpcCompatPbImpl struct{}

func (s *grpcCompatPbImpl) UnaryTest(ctx context.Context, req *pbapi.MockReq) (*pbapi.MockResp, error) {
	return &pbapi.MockResp{Message: req.Message}, nil
}

func (s *grpcCompatPbImpl) BidirectionalStreamingTest(ctx context.Context, stream pbapi.Mock_BidirectionalStreamingTestServer) error {
	if ctx.Value(oldStreamMWKey{}) != true {
		return errors.New("old stream middleware wrapping not preserved")
	}
	for {
		req, err := stream.Recv(ctx)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err = stream.Send(ctx, &pbapi.MockResp{Message: req.Message}); err != nil {
			return err
		}
	}
}

func (s *grpcCompatPbImpl) ClientStreamingTest(ctx context.Context, stream pbapi.Mock_ClientStreamingTestServer) error {
	if ctx.Value(oldStreamMWKey{}) != true {
		return errors.New("old stream middleware wrapping not preserved")
	}
	var lastMsg string
	for {
		req, err := stream.Recv(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		lastMsg = req.Message
	}
	return stream.SendAndClose(ctx, &pbapi.MockResp{Message: lastMsg})
}

func (s *grpcCompatPbImpl) ServerStreamingTest(ctx context.Context, req *pbapi.MockReq, stream pbapi.Mock_ServerStreamingTestServer) error {
	if ctx.Value(oldStreamMWKey{}) != true {
		return errors.New("old stream middleware wrapping not preserved")
	}
	return stream.Send(ctx, &pbapi.MockResp{Message: req.Message})
}

// RunGRPCCompatServer creates a server with old-style endpoint middleware that wraps
// streaming.Args.Stream. This tests the fix that the wrapped stream is preserved
// (not discarded) when passed to the stream handler endpoint.
func RunGRPCCompatServer(listenAddr string) server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", listenAddr)
	svr := echoservice.NewServer(&grpcCompatThriftImpl{},
		server.WithServiceAddr(addr),
		server.WithExitWaitTime(1*time.Second),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler),
		// old-style endpoint middleware: wraps streaming.Args.Stream
		server.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) error {
				if args, ok := req.(*streaming.Args); ok && args.Stream != nil {
					args.Stream = &wrappedOldStream{Stream: args.Stream}
				}
				return next(ctx, req, resp)
			}
		}),
		// stream middleware: verifies the wrapped stream is preserved via GRPCStreamGetter
		server.WithStreamOptions(server.WithStreamMiddleware(func(next sep.StreamEndpoint) sep.StreamEndpoint {
			return func(ctx context.Context, stream streaming.ServerStream) error {
				if getter, ok := stream.(streaming.GRPCStreamGetter); ok {
					st := getter.GetGRPCStream()
					if _, ok := st.(*wrappedOldStream); ok {
						ctx = context.WithValue(ctx, oldStreamMWKey{}, true)
					}
				}
				return next(ctx, stream)
			}
		})),
	)
	if err := mock.RegisterService(svr, &grpcCompatPbImpl{}); err != nil {
		panic(err)
	}
	go func() {
		if err := svr.Run(); err != nil {
			println(err)
		}
	}()
	return svr
}

// TestThriftGRPCCompat tests that old-style endpoint middleware wrapping streaming.Args.Stream
// is preserved and accessible in the stream middleware chain for thrift streaming calls.
func TestThriftGRPCCompat(t *testing.T, addr string) {
	testcases := []struct {
		desc string
		prot transport.Protocol
	}{
		{"GRPCStreaming", transport.TTHeader | transport.GRPCStreaming},
		{"GRPC", transport.GRPC},
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			cli := echoservice.MustNewClient("service",
				client.WithHostPorts(addr),
				client.WithTransportProtocol(tc.prot),
				client.WithMetaHandler(transmeta.ClientHTTP2Handler),
				client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
			)
			t.Run("BidiStreaming", func(t *testing.T) {
				ctx := context.Background()
				stream, err := cli.EchoBidi(ctx)
				test.Assert(t, err == nil, err)
				err = stream.Send(ctx, &tenant.EchoRequest{Msg: "ping"})
				test.Assert(t, err == nil, err)
				err = stream.CloseSend(ctx)
				test.Assert(t, err == nil, err)
				resp, err := stream.Recv(ctx)
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Msg == "ping", resp.Msg)
				_, err = stream.Recv(ctx)
				test.Assert(t, err == io.EOF)
			})
			t.Run("ClientStreaming", func(t *testing.T) {
				ctx := context.Background()
				stream, err := cli.EchoClient(ctx)
				test.Assert(t, err == nil, err)
				err = stream.Send(ctx, &tenant.EchoRequest{Msg: "ping"})
				test.Assert(t, err == nil, err)
				resp, err := stream.CloseAndRecv(ctx)
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Msg == "ping", resp.Msg)
			})
			t.Run("ServerStreaming", func(t *testing.T) {
				ctx := context.Background()
				stream, err := cli.EchoServer(ctx, &tenant.EchoRequest{Msg: "ping"})
				test.Assert(t, err == nil, err)
				resp, err := stream.Recv(ctx)
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Msg == "ping", resp.Msg)
				_, err = stream.Recv(ctx)
				test.Assert(t, err == io.EOF)
			})
		})
	}
}

// TestPbGRPCCompat tests that old-style endpoint middleware wrapping streaming.Args.Stream
// is preserved and accessible in the stream middleware chain for pb streaming calls.
func TestPbGRPCCompat(t *testing.T, addr string) {
	testcases := []struct {
		desc string
		prot transport.Protocol
	}{
		{"GRPCStreaming", transport.TTHeader | transport.GRPCStreaming},
		{"GRPC", transport.GRPC},
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			cli := mock.MustNewClient("service",
				client.WithHostPorts(addr),
				client.WithTransportProtocol(tc.prot),
				client.WithMetaHandler(transmeta.ClientHTTP2Handler),
				client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
			)
			t.Run("BidiStreaming", func(t *testing.T) {
				ctx := context.Background()
				stream, err := cli.BidirectionalStreamingTest(ctx)
				test.Assert(t, err == nil, err)
				err = stream.Send(ctx, &pbapi.MockReq{Message: "ping"})
				test.Assert(t, err == nil, err)
				err = stream.CloseSend(ctx)
				test.Assert(t, err == nil, err)
				resp, err := stream.Recv(ctx)
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Message == "ping", resp.Message)
				_, err = stream.Recv(ctx)
				test.Assert(t, err == io.EOF)
			})
			t.Run("ClientStreaming", func(t *testing.T) {
				ctx := context.Background()
				stream, err := cli.ClientStreamingTest(ctx)
				test.Assert(t, err == nil, err)
				err = stream.Send(ctx, &pbapi.MockReq{Message: "ping"})
				test.Assert(t, err == nil, err)
				resp, err := stream.CloseAndRecv(ctx)
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Message == "ping", resp.Message)
			})
			t.Run("ServerStreaming", func(t *testing.T) {
				ctx := context.Background()
				stream, err := cli.ServerStreamingTest(ctx, &pbapi.MockReq{Message: "ping"})
				test.Assert(t, err == nil, err)
				resp, err := stream.Recv(ctx)
				test.Assert(t, err == nil, err)
				test.Assert(t, resp.Message == "ping", resp.Message)
				_, err = stream.Recv(ctx)
				test.Assert(t, err == io.EOF)
			})
		})
	}
}
