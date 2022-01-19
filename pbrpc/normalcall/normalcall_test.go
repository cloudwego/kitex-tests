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

package normalcall

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/stability/stservice"
	"github.com/cloudwego/kitex-tests/pbrpc"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/transport"
)

func TestMain(m *testing.M) {
	svr := pbrpc.RunServer(&pbrpc.ServerInitParam{
		Network: "tcp",
		Address: ":8001",
	}, nil)
	time.Sleep(time.Second)
	m.Run()
	svr.Stop()
}

func getKitexClient(p transport.Protocol) stservice.Client {
	return pbrpc.CreateKitexClient(&pbrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{":8001"},
		Protocol:          p,
		ConnMode:          pbrpc.LongConnection,
	})
}

func TestStTReq(t *testing.T) {
	cli := getKitexClient(transport.PurePayload)

	ctx, stReq := pbrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)

	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)

	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)
}

func TestStTReqWithTTHeader(t *testing.T) {
	cli := getKitexClient(transport.TTHeader)

	ctx, stReq := pbrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(time.Second))
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

func BenchmarkNormalCall(b *testing.B) {
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
		objReq, err := cli.TestObjReq(ctx, objReq)
		test.Assert(b, err == nil, err)
		test.Assert(b, objReq.FlagMsg == objReq.FlagMsg)
	}
}

func BenchmarkPBCallWithTTHeader(b *testing.B) {
	cli := getKitexClient(transport.TTHeader)
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

//  go test -benchmem -bench=BenchmarkTTHeaderParallel ./mocks/pbrpc/normalcall/ | grep '^Benchmark' | tee -a pbbench.txt
func BenchmarkTTHeaderParallel(b *testing.B) {
	cli := getKitexClient(transport.TTHeader)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ctx, stReq := pbrpc.CreateSTRequest(context.Background())
			stReq.FlagMsg = strconv.Itoa(i)
			stResp, err := cli.TestSTReq(ctx, stReq)
			test.Assert(b, err == nil, err)
			test.Assert(b, stReq.FlagMsg == stResp.FlagMsg)

			ctx, objReq := pbrpc.CreateObjReq(context.Background())
			objReq.FlagMsg = strconv.Itoa(i)
			objResp, err := cli.TestObjReq(ctx, objReq)
			test.Assert(b, err == nil, err)
			test.Assert(b, objReq.FlagMsg == objResp.FlagMsg)
			i++
		}
	})
}
