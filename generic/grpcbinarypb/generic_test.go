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

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/pbapi"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/server"
)

func TestClientStreaming(t *testing.T) {
	ctx := context.Background()
	ln := serverutils.Listen()
	svr := initMockTestServer(new(StreamingTestImpl), ln)
	defer svr.Stop()

	cli := initStreamingClient(t, ctx, ln.Addr().String(), "", "Mock", "pbapi")
	streamCli, err := genericclient.NewClientStreaming(ctx, cli, "ClientStreamingTest")
	test.Assert(t, err == nil, err)
	for i := 0; i < 3; i++ {
		req := &pbapi.MockReq{Message: "Hello from Client"}
		buf := make([]byte, 0)
		buf, _ = req.Marshal(buf)
		err = streamCli.Send(buf)
		test.Assert(t, err == nil)
	}
	buf, err := streamCli.CloseAndRecv()
	test.Assert(t, err == nil)
	_, ok := buf.([]byte)
	test.Assert(t, ok)
	resp := &pbapi.MockResp{}
	err = resp.Unmarshal(buf.([]byte))
	test.Assert(t, err == nil)
	fmt.Printf("clientStreaming message received: %v\n", resp)
	test.Assert(t, resp.Message == "all message: Hello from Client, Hello from Client, Hello from Client")
}

func TestServerStreaming(t *testing.T) {
	ctx := context.Background()
	ln := serverutils.Listen()

	svr := initMockTestServer(new(StreamingTestImpl), ln)
	defer svr.Stop()

	cli := initStreamingClient(t, ctx, ln.Addr().String(), "", "Mock", "pbapi")
	req := &pbapi.MockReq{Message: "Hello from Client"}
	buf := make([]byte, 0)
	buf, _ = req.Marshal(buf)
	streamCli, err := genericclient.NewServerStreaming(ctx, cli, "ServerStreamingTest", buf)
	test.Assert(t, err == nil, err)
	for {
		buf, err := streamCli.Recv()
		if err != nil {
			test.Assert(t, err == io.EOF)
			fmt.Println("serverStreaming message receive done")
			break
		} else {
			_, ok := buf.([]byte)
			test.Assert(t, ok)
			resp := &pbapi.MockResp{}
			err = resp.Unmarshal(buf.([]byte))
			test.Assert(t, err == nil)
			fmt.Printf("serverStreaming message received: %v\n", resp)
			test.Assert(t, resp.Message == "Hello from Client -> response")
		}
	}
}

func TestBidirectionalStreaming(t *testing.T) {
	ctx := context.Background()
	ln := serverutils.Listen()

	svr := initMockTestServer(new(StreamingTestImpl), ln)
	defer svr.Stop()

	cli := initStreamingClient(t, ctx, ln.Addr().String(), "", "Mock", "pbapi")
	streamCli, err := genericclient.NewBidirectionalStreaming(ctx, cli, "BidirectionalStreamingTest")
	test.Assert(t, err == nil)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer streamCli.Close()
		for i := 0; i < 3; i++ {
			buf := make([]byte, 0)
			req := &pbapi.MockReq{Message: fmt.Sprintf("Hello from Client %d", i)}
			buf, _ = req.Marshal(buf)
			err = streamCli.Send(buf)
			test.Assert(t, err == nil)
			fmt.Printf("BidirectionalStreamingTest send: req = %v\n", req)
		}
	}()

	go func() {
		defer wg.Done()
		for {
			buf, err := streamCli.Recv()
			if err != nil {
				test.Assert(t, err == io.EOF)
				fmt.Println("bidirectionalStreaming message receive done")
				break
			} else {
				_, ok := buf.([]byte)
				test.Assert(t, ok)
				resp := &pbapi.MockResp{}
				err = resp.Unmarshal(buf.([]byte))
				test.Assert(t, err == nil)
				fmt.Printf("bidirectionalStreaming message received: %s\n", resp)
				test.Assert(t, resp.Message == "Hello from Server")
			}
		}
	}()
	wg.Wait()
}

func TestUnary(t *testing.T) {
	ctx := context.Background()
	ln := serverutils.Listen()

	svr := initMockTestServer(new(StreamingTestImpl), ln)
	defer svr.Stop()

	cli := initStreamingClient(t, ctx, ln.Addr().String(), "", "Mock", "pbapi")
	req := &pbapi.MockReq{Message: "Hello from Client"}
	buf := make([]byte, 0)
	buf, _ = req.Marshal(buf)
	res, err := cli.GenericCall(ctx, "UnaryTest", buf)
	test.Assert(t, err == nil)
	_, ok := res.([]byte)
	test.Assert(t, ok)
	resp := &pbapi.MockResp{}
	err = resp.Unmarshal(res.([]byte))
	test.Assert(t, err == nil)
	fmt.Printf("unary message received: %v\n", resp)
	test.Assert(t, resp.Message == "hello Hello from Client")
}

func initStreamingClient(t *testing.T, ctx context.Context, addr, idl, serviceName, packageName string, cliOpts ...client.Option) genericclient.Client {
	g := generic.BinaryPbGeneric(serviceName, packageName)
	return newGenericClient(g, addr, cliOpts...)
}

func initMockTestServer(handler pbapi.Mock, ln net.Listener) server.Server {
	svr := newMockTestServer(handler, ln)
	return svr
}
