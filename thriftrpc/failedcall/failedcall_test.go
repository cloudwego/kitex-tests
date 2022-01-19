// Copyright 2021 CloudWeGo Authors
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

package failedcall

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/codec"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

type mockedCodec struct {
	remote.Codec
}

func (mc *mockedCodec) Decode(ctx context.Context, msg remote.Message, in remote.ByteBuffer) error {
	mc.Codec.Decode(ctx, msg, in)
	return remote.NewTransError(remote.ProtocolError, thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, "mocked error"))
}

func (mc *mockedCodec) Name() string {
	return "mockedCodec"
}

func TestMain(m *testing.M) {
	svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9001",
	}, nil, server.WithCodec(&mockedCodec{
		Codec: codec.NewDefaultCodec(),
	}))
	time.Sleep(time.Second)
	m.Run()
	svr.Stop()
}

func getKitexClient(p transport.Protocol) stservice.Client {
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{":9001"},
		Protocol:          p,
		ConnMode:          thriftrpc.LongConnection,
	})
}

// TestSTReq method mock STRequest param read failed in server
func TestStTReq(t *testing.T) {
	cli := getKitexClient(transport.PurePayload)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	var opts []callopt.Option
	opts = nil

	stResp, err := cli.TestSTReq(ctx, stReq, opts...)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	test.Assert(t, strings.Contains(err.Error(), "mocked error"), err.Error())
}

func TestStTReqWithTTHeader(t *testing.T) {
	cli := getKitexClient(transport.TTHeader)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	test.Assert(t, strings.Contains(err.Error(), "mocked error"), err.Error())
}

func TestStTReqWithFramed(t *testing.T) {
	cli := getKitexClient(transport.Framed)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	test.Assert(t, strings.Contains(err.Error(), "mocked error"), err.Error())
}

func TestStTReqWithTTHeaderFramed(t *testing.T) {
	cli := getKitexClient(transport.TTHeaderFramed)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	test.Assert(t, strings.Contains(err.Error(), "mocked error"), err.Error())
}

// TestObjReq method mock ObjResp read failed in client
func TestObjReq(t *testing.T) {
	cli := getKitexClient(transport.PurePayload)

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = "ObjReq"
	objResp, err := cli.TestObjReq(ctx, objReq)
	test.Assert(t, err != nil)
	test.Assert(t, objResp == nil)
	test.Assert(t, strings.Contains(err.Error(), "mocked error"), err.Error())
}

// oneway cannot read failed of server
func TestVisitOneway(t *testing.T) {
	cli := getKitexClient(transport.TTHeaderFramed)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	err := cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)

	time.Sleep(time.Second / 2) // wait for the TCP close signal from server
	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	err = cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)

	time.Sleep(time.Second / 2) // wait for the TCP close signal from server
	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	err = cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)
}
