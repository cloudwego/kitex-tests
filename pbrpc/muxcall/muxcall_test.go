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
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/stability/stservice"
	"github.com/cloudwego/kitex-tests/pbrpc"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex/transport"
)

func TestMain(m *testing.M) {
	svr := pbrpc.RunServer(&pbrpc.ServerInitParam{
		Network: "tcp",
		Address: ":8002",
	}, nil)
	time.Sleep(2 * time.Second)
	m.Run()
	time.Sleep(2 * time.Second)
	svr.Stop()
}

func getKitexClient(p transport.Protocol) stservice.Client {
	return pbrpc.CreateKitexClient(&pbrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{":8002"},
		Protocol:          p,
		ConnMode:          pbrpc.ConnectionMultiplexed,
	})
}

func TestStTReq(t *testing.T) {
	cli := getKitexClient(transport.PurePayload)

	ctx, stReq := pbrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)
}

func TestObjReq(t *testing.T) {
	cli := getKitexClient(transport.PurePayload)

	ctx, objReq := pbrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = "ObjReq"
	objResp, err := cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, objReq.FlagMsg == objResp.FlagMsg)
}

func BenchmarkMuxCall(b *testing.B) {
	cli := getKitexClient(transport.PurePayload)

	ctx, stReq := pbrpc.CreateSTRequest(context.Background())
	ctx, objReq := pbrpc.CreateObjReq(context.Background())
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
	cli := getKitexClient(transport.PurePayload)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, stReq := pbrpc.CreateSTRequest(context.Background())
		stResp, err := cli.TestSTReq(ctx, stReq)
		test.Assert(b, err == nil, err)
		test.Assert(b, stReq.FlagMsg == stResp.FlagMsg)
	}
}
