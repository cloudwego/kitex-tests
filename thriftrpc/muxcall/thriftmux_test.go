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

package muxcall

import (
	"context"
	"strconv"
	"testing"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/thriftrpc"
)

var testaddr string

func TestMain(m *testing.M) {
	testaddr = serverutils.NextListenAddr()
	svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network:  "tcp",
		Address:  testaddr,
		ConnMode: thriftrpc.ConnectionMultiplexed,
	}, nil)
	serverutils.Wait(testaddr)
	m.Run()
	svr.Stop()
}

func getKitexMuxClient(opts ...client.Option) stservice.Client {
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{testaddr},
		ConnMode:          thriftrpc.ConnectionMultiplexed,
	}, opts...)
}

func TestStTReq(t *testing.T) {
	cli := getKitexMuxClient()

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)
}

func TestObjReq(t *testing.T) {
	cli := getKitexMuxClient()

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = "ObjReq"
	objResp, err := cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, objReq.FlagMsg == objResp.FlagMsg)
}

func TestException(t *testing.T) {
	cli := getKitexMuxClient()

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	_, err := cli.TestException(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, err.Error() == "STException({Message:mock exception})", err)
	nr, ok := err.(*stability.STException)
	test.Assert(t, ok)
	test.Assert(t, nr.Message == "mock exception")
}

func TestVisitOneway(t *testing.T) {
	cli := getKitexMuxClient()
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	err := cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)
}

func TestDisableRPCInfoReuse(t *testing.T) {
	backupState := rpcinfo.PoolEnabled()
	defer rpcinfo.EnablePool(backupState)

	addr := serverutils.NextListenAddr()
	var ri rpcinfo.RPCInfo
	svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network:  "tcp",
		Address:  addr,
		ConnMode: thriftrpc.ConnectionMultiplexed,
	}, nil, server.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) error {
			ri = rpcinfo.GetRPCInfo(ctx)
			return next(ctx, req, resp)
		}
	}))
	defer svr.Stop()
	serverutils.Wait(addr)

	cli := getKitexMuxClient(client.WithTransportProtocol(transport.TTHeaderFramed))
	ctx, stReq := thriftrpc.CreateSTRequest(metainfo.WithBackwardValues(context.Background()))

	t.Run("reuse", func(t *testing.T) {
		_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(addr))
		test.Assert(t, err == nil, err)
		test.Assert(t, ri.Invocation().MethodName() == "", ri.Invocation().MethodName()) // zeroed
	})

	t.Run("disable reuse", func(t *testing.T) {
		rpcinfo.EnablePool(false)
		_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(addr))
		test.Assert(t, err == nil, err)
		test.Assert(t, ri.Invocation().MethodName() != "", ri.Invocation().MethodName())
	})
}

func BenchmarkMuxCall(b *testing.B) {
	cli := getKitexMuxClient()

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stReq.FlagMsg = strconv.Itoa(i)
		stResp, err := cli.TestSTReq(ctx, stReq)
		test.Assert(b, err == nil, err)
		test.Assert(b, stReq.FlagMsg == stResp.FlagMsg)

		objReq.FlagMsg = strconv.Itoa(i)
		objResp, err := cli.TestObjReq(ctx, objReq)
		test.Assert(b, err == nil, err)
		test.Assert(b, objReq.FlagMsg == objResp.FlagMsg)
	}
}

func BenchmarkMuxCallParallel(b *testing.B) {
	cli := getKitexMuxClient()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		_, err := cli.TestException(ctx, stReq)
		test.Assert(b, err != nil)
		test.Assert(b, err.Error() == "STException({Message:mock exception})")
		nr, ok := err.(*stability.STException)
		test.Assert(b, ok)
		test.Assert(b, nr.Message == "mock exception")

		err = cli.VisitOneway(ctx, stReq)
		test.Assert(b, err == nil, err)
	}
}
