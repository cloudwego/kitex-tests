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
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/streamx/kitex_gen/echo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/server"
)

func TestClientStreaming(t *testing.T) {
	ctx := context.Background()
	ln := serverutils.Listen()
	svr := initMockTestServer(new(serviceImpl), ln)
	defer svr.Stop()

	cli := initStreamingClient(t, ctx, ln.Addr().String(), "./idl/echo.thrift")
	streamCli, err := cli.ClientStreaming(ctx, "EchoClient")
	test.Assert(t, err == nil, err)
	for i := 0; i < 3; i++ {
		req := `{"Message":"Hello from Client"}`
		err = streamCli.Send(streamCli.Context(), req)
		test.Assert(t, err == nil)
	}
	resp, err := streamCli.CloseAndRecv(streamCli.Context())
	test.Assert(t, err == nil)
	s, ok := resp.(string)
	test.Assert(t, ok)
	fmt.Printf("clientStreaming message received: %v\n", resp)
	test.Assert(t, s == `{"Message":"all message: Hello from Client, Hello from Client, Hello from Client"}`)
}

func TestServerStreaming(t *testing.T) {
	ctx := context.Background()
	ln := serverutils.Listen()

	svr := initMockTestServer(new(serviceImpl), ln)
	defer svr.Stop()

	cli := initStreamingClient(t, ctx, ln.Addr().String(), "./idl/echo.thrift")
	req := `{"Message":"Hello from Client"}`
	streamCli, err := cli.ServerStreaming(ctx, "EchoServer", req)
	test.Assert(t, err == nil, err)
	for {
		resp, err := streamCli.Recv(streamCli.Context())
		if err != nil {
			test.Assert(t, err == io.EOF)
			fmt.Println("serverStreaming message receive done")
			break
		} else {
			s, ok := resp.(string)
			test.Assert(t, ok)
			fmt.Printf("serverStreaming message received: %v\n", resp)
			test.Assert(t, s == `{"Message":"Hello from Client response"}`)
		}
	}
}

func TestBidirectionalStreaming(t *testing.T) {
	ctx := context.Background()
	ln := serverutils.Listen()

	svr := initMockTestServer(new(serviceImpl), ln)
	defer svr.Stop()

	cli := initStreamingClient(t, ctx, ln.Addr().String(), "./idl/echo.thrift")
	streamCli, err := cli.BidirectionalStreaming(ctx, "EchoBidi")
	test.Assert(t, err == nil)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer streamCli.CloseSend(streamCli.Context())
		for i := 0; i < 3; i++ {
			req := `{"Message":` + fmt.Sprintf(`"Hello from Client %d"}`, i)
			err = streamCli.Send(streamCli.Context(), req)
			test.Assert(t, err == nil)
			fmt.Printf("BidirectionalStreamingTest send: req = %v\n", req)
		}
	}()

	go func() {
		defer wg.Done()
		for {
			resp, err := streamCli.Recv(streamCli.Context())
			if err != nil {
				test.Assert(t, err == io.EOF)
				fmt.Println("bidirectionalStreaming message receive done")
				break
			} else {
				s, ok := resp.(string)
				test.Assert(t, ok)
				fmt.Printf("bidirectionalStreaming message received: %s\n", resp)
				test.Assert(t, s == `{"Message":"Hello from Server"}`)
			}
		}
	}()
	wg.Wait()
}

func TestUnary(t *testing.T) {
	ctx := context.Background()
	ln := serverutils.Listen()

	svr := initMockTestServer(new(serviceImpl), ln)
	defer svr.Stop()

	cli := initStreamingClient(t, ctx, ln.Addr().String(), "./idl/echo.thrift")
	req := `{"Message":"Hello from Client"}`
	resp, err := cli.GenericCall(ctx, "Unary", req)
	test.Assert(t, err == nil)
	s, ok := resp.(string)
	test.Assert(t, ok)
	fmt.Printf("unary message received: %v\n", resp)
	test.Assert(t, s == `{"Message":"hello Hello from Client"}`)
}

func initStreamingClient(t *testing.T, ctx context.Context, addr, idl string, cliOpts ...client.Option) genericclient.Client {
	p, err := generic.NewThriftFileProvider(idl)
	test.Assert(t, err == nil)
	g, err := generic.JSONThriftGeneric(p)
	test.Assert(t, err == nil)
	return newGenericClient(g, addr, cliOpts...)
}

func initMockTestServer(handler echo.TestService, ln net.Listener) server.Server {
	svr := newMockTestServer(handler, ln)
	return svr
}
