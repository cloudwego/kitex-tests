// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package thrift_streaming

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_old/echo"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_old/echo/echoservice"
)

func TestOldKitexThriftBidirectional(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr))
	t.Run("normal", func(t *testing.T) {
		n := 3
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(n))
		_, err := cli.EchoBidirectional(ctx, &echo.EchoRequest{Message: "bidirectional"})
		test.Assert(t, err != nil, err)
		test.Assert(t, strings.Contains(err.Error(), "please call with Kitex StreamClient"), err)
	})
}

func TestOldKitexThriftClient(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr))
	t.Run("normal", func(t *testing.T) {
		n := 3
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(n))
		_, err := cli.EchoClient(ctx, &echo.EchoRequest{Message: "client"})
		test.Assert(t, err != nil, err)
		test.Assert(t, strings.Contains(err.Error(), "please call with Kitex StreamClient"), err)
	})
}

func TestOldKitexThriftServer(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr),
	)
	t.Run("normal", func(t *testing.T) {
		n := 3
		ctx := metainfo.WithValue(context.Background(), KeyCount, strconv.Itoa(n))
		_, err := cli.EchoServer(ctx, &echo.EchoRequest{Message: "server"})
		test.Assert(t, err != nil, err)
		test.Assert(t, strings.Contains(err.Error(), "please call with Kitex StreamClient"), err)
	})
}

func TestOldKitexThriftEchoUnary(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr))
	t.Run("normal", func(t *testing.T) {
		_, err := cli.EchoUnary(context.Background(), &echo.EchoRequest{Message: "unary"})
		test.Assert(t, err != nil, err)
		test.Assert(t, strings.Contains(err.Error(), "please call with Kitex StreamClient"), err)
	})
}

func TestOldKitexThriftEchoPingPong(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr),
		client.WithTransportProtocol(transport.TTHeaderFramed))

	t.Run("normal", func(t *testing.T) {
		req1 := &echo.EchoRequest{Message: "hello"}
		req2 := &echo.EchoRequest{Message: "world"}
		resp, err := cli.EchoPingPong(context.Background(), req1, req2)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "hello-world", resp.Message)
	})

	t.Run("exception", func(t *testing.T) {
		ctx := metainfo.WithValue(context.Background(), KeyError, "pingpong")
		_, err := cli.EchoPingPong(ctx, &echo.EchoRequest{Message: "hello"}, &echo.EchoRequest{Message: "world"})
		test.Assert(t, err != nil, err)
	})
}

func TestOldKitexThriftEchoOneway(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr),
		client.WithTransportProtocol(transport.TTHeaderFramed))

	req := &echo.EchoRequest{Message: "oneway"}

	t.Run("normal", func(t *testing.T) {
		ctx := metainfo.WithValue(context.Background(), KeyMessage, req.Message)
		err := cli.EchoOneway(ctx, req)
		test.Assert(t, err == nil, err)
	})
	t.Run("normal", func(t *testing.T) {
		err := cli.EchoOneway(context.Background(), req)
		test.Assert(t, err != nil, err)
	})
}

func TestOldKitexThriftPing(t *testing.T) {
	cli := echoservice.MustNewClient(
		"service",
		client.WithHostPorts(thriftAddr),
		client.WithTransportProtocol(transport.TTHeaderFramed))
	t.Run("normal", func(t *testing.T) {
		err := cli.Ping(context.Background())
		test.Assert(t, err == nil, err)
	})
	t.Run("exception", func(t *testing.T) {
		ctx := metainfo.WithValue(context.Background(), KeyError, "ping")
		err := cli.Ping(ctx)
		test.Assert(t, err != nil, err)
	})
}
