// Copyright 2024 CloudWeGo Authors
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

package tests

import (
	"context"
	"encoding/json"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/generic/thrift"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
)

var (
	testaddr      string
	genericAddr   string
	genericV2Addr string
	req           = map[string]interface{}{
		"Msg": "hello",
		"I8":  int8(1),
		"I16": int16(1),
		"I32": int32(1),
		"I64": int64(1),
		"Map": map[string]interface{}{
			"hello": "world",
		},
		"Set":       []interface{}{"hello", "world"},
		"List":      []interface{}{"hello", "world"},
		"ErrorCode": int32(1),
		"Info": map[string]interface{}{
			"Map": map[string]interface{}{
				"hello": "world",
			},
			"ID": int64(232324),
		},
	}
	reqStr, _ = json.Marshal(req)
)

func TestMain(m *testing.M) {
	ln1 := serverutils.Listen()
	testaddr = ln1.Addr().String()
	ln2 := serverutils.Listen()
	genericAddr = ln2.Addr().String()
	ln3 := serverutils.Listen()
	genericV2Addr = ln3.Addr().String()
	svc := runServer(ln1)
	gsvc := runGenericServer(ln2)
	gsvc2 := runGenericServerV2(ln3)
	m.Run()
	svc.Stop()
	gsvc.Stop()
	gsvc2.Stop()
}

func newGenericClient(destService string, g generic.Generic, opts ...client.Option) (genericclient.Client, error) {
	opts = append(opts, client.WithMetaHandler(transmeta.ClientTTHeaderHandler))
	opts = append(opts, client.WithMetaHandler(transmeta.ClientHTTP2Handler))
	return genericclient.NewClient(destService, g, opts...)
}

func TestPingPong(t *testing.T) {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	test.Assert(t, err == nil)
	g, err := generic.JSONThriftGeneric(p)
	test.Assert(t, err == nil)

	cli, err := newGenericClient("a.b.c", g, client.WithHostPorts(testaddr))
	test.Assert(t, err == nil)

	_, err = cli.GenericCall(context.Background(), "Echo", string(reqStr))
	test.Assert(t, err == nil)
}

func TestOneway(t *testing.T) {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	test.Assert(t, err == nil)
	g, err := generic.JSONThriftGeneric(p)
	test.Assert(t, err == nil)

	cli, err := newGenericClient("a.b.c", g, client.WithHostPorts(testaddr))
	test.Assert(t, err == nil)

	req := map[string]interface{}{
		"Msg": "hello",
		"I8":  int8(1),
		"I16": int16(1),
		"I32": int32(1),
		"I64": int64(1),
		"Map": map[string]interface{}{
			"hello": "world",
		},
		"Set":       []interface{}{"hello", "world"},
		"List":      []interface{}{"hello", "world"},
		"ErrorCode": int32(1),
		"Info": map[string]interface{}{
			"Map": map[string]interface{}{
				"hello": "world",
			},
			"ID": int64(232324),
		},
	}
	reqStr, _ := json.Marshal(req)
	num := 10
	for i := 0; i < num; i++ {
		_, err = cli.GenericCall(context.Background(), "EchoOneway", string(reqStr))
		test.Assert(t, err == nil)
	}
	// wait for request received
	time.Sleep(200 * time.Millisecond)
	test.Assert(t, atomic.LoadInt32(&checkNum) == int32(num))
}

func TestBizErr(t *testing.T) {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	test.Assert(t, err == nil)
	g, err := generic.JSONThriftGeneric(p)
	test.Assert(t, err == nil)

	cli, err := newGenericClient("a.b.c", g,
		client.WithHostPorts(genericAddr),
		client.WithTransportProtocol(transport.TTHeader))
	test.Assert(t, err == nil)
	_, err = cli.GenericCall(context.Background(), "Echo", `{"Msg":"biz_error"}`)
	bizerr, ok := kerrors.FromBizStatusError(err)
	test.Assert(t, ok)
	test.Assert(t, bizerr.BizStatusCode() == 404)
	test.Assert(t, bizerr.BizMessage() == "not found")
}

func TestCombinedServicesParseMode(t *testing.T) {
	p, err := generic.NewThriftFileProviderWithOption("../../idl/tenant.thrift", []generic.ThriftIDLProviderOption{generic.WithParseMode(thrift.CombineServices)})
	test.Assert(t, err == nil)
	g, err := generic.JSONThriftGeneric(p)
	test.Assert(t, err == nil)

	cli, err := newGenericClient("a.b.c", g, client.WithHostPorts(testaddr))
	test.Assert(t, err == nil)

	_, err = cli.GenericCall(context.Background(), "Echo", string(reqStr))
	test.Assert(t, err == nil)
}

func TestGenericServiceImplV2_ClientStreaming(t *testing.T) {
	protocols := []transport.Protocol{transport.Framed, transport.TTHeader, transport.GRPC, transport.GRPCStreaming | transport.TTHeader}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
			test.Assert(t, err == nil)
			g, err := generic.JSONThriftGeneric(p)
			test.Assert(t, err == nil)

			cli, err := newGenericClient("a.b.c", g, client.WithHostPorts(genericV2Addr), client.WithTransportProtocol(protocol))
			test.Assert(t, err == nil)

			stream, err := cli.ClientStreaming(context.Background(), "EchoClient")
			test.Assert(t, err == nil)

			err = stream.Send(stream.Context(), `{"Msg":"hello world"}`)
			test.Assert(t, err == nil)

			resp, err := stream.CloseAndRecv(stream.Context())
			test.Assert(t, err == nil)
			var response tenant.EchoResponse
			err = json.Unmarshal([]byte(resp.(string)), &response)
			test.Assert(t, err == nil)
			test.Assert(t, response.Msg == "world")
		})
	}
}

func TestGenericServiceImplV2_ServerStreaming(t *testing.T) {
	protocols := []transport.Protocol{transport.Framed, transport.TTHeader, transport.GRPC, transport.GRPCStreaming | transport.TTHeader}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
			test.Assert(t, err == nil)
			g, err := generic.JSONThriftGeneric(p)
			test.Assert(t, err == nil)

			cli, err := newGenericClient("a.b.c", g, client.WithHostPorts(genericV2Addr), client.WithTransportProtocol(protocol))
			test.Assert(t, err == nil)

			stream, err := cli.ServerStreaming(context.Background(), "EchoServer", `{"Msg":"hello world"}`)
			test.Assert(t, err == nil)

			resp, err := stream.Recv(stream.Context())
			test.Assert(t, err == nil)
			var response tenant.EchoResponse
			err = json.Unmarshal([]byte(resp.(string)), &response)
			test.Assert(t, err == nil)
			test.Assert(t, response.Msg == "world")

			_, err = stream.Recv(stream.Context())
			test.Assert(t, err == io.EOF)
		})
	}
}

func TestGenericServiceImplV2_BidiStreaming(t *testing.T) {
	protocols := []transport.Protocol{transport.Framed, transport.TTHeader, transport.GRPC, transport.GRPCStreaming | transport.TTHeader}
	for _, protocol := range protocols {
		t.Run(protocol.String(), func(t *testing.T) {
			p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
			test.Assert(t, err == nil)
			g, err := generic.JSONThriftGeneric(p)
			test.Assert(t, err == nil)

			cli, err := newGenericClient("a.b.c", g, client.WithHostPorts(genericV2Addr), client.WithTransportProtocol(protocol))
			test.Assert(t, err == nil)

			stream, err := cli.BidirectionalStreaming(context.Background(), "EchoBidi")
			test.Assert(t, err == nil)

			err = stream.Send(stream.Context(), `{"Msg":"hello world"}`)
			test.Assert(t, err == nil)

			err = stream.CloseSend(stream.Context())
			test.Assert(t, err == nil)

			resp, err := stream.Recv(stream.Context())
			test.Assert(t, err == nil)

			var response tenant.EchoResponse
			err = json.Unmarshal([]byte(resp.(string)), &response)
			test.Assert(t, err == nil)
			test.Assert(t, response.Msg == "world")

			_, err = stream.Recv(stream.Context())
			test.Assert(t, err == io.EOF)
		})
	}
}
