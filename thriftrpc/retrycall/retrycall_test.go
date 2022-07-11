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

package retrycall

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/circuitbreak"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/retry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/transport"
)

var retryChainStopStr = "chain stop retry"

func TestMain(m *testing.M) {
	svr1 := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9001",
	}, new(STServiceHandler))

	svr2 := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9002",
	}, new(STServiceMockResultHandler))

	time.Sleep(time.Second)
	m.Run()
	svr1.Stop()
	svr2.Stop()
}

func getKitexClient(p transport.Protocol, opts ...client.Option) stservice.Client {
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{":9001"},
		Protocol:          p,
		ConnMode:          thriftrpc.LongConnection,
	}, opts...)
}

func genCBKey(ri rpcinfo.RPCInfo) string {
	sss := []string{
		ri.From().ServiceName(),
		ri.To().ServiceName(),
		ri.To().Method(),
	}
	return strings.Join(sss, "/")
}

func TestRetryCB(t *testing.T) {
	atomic.StoreInt32(&testSTReqCount, -1)
	// retry config
	fp := retry.NewFailurePolicy()

	cli := getKitexClient(
		transport.PurePayload,
		client.WithFailureRetry(fp),
		client.WithRPCTimeout(20*time.Millisecond),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	cbCount := 0
	for i := 0; i < 300; i++ {
		stResp, err := cli.TestSTReq(ctx, stReq)
		if err != nil {
			test.Assert(t, strings.Contains(err.Error(), "retry circuit break"), err, i)
			cbCount++
		} else {
			test.Assert(t, err == nil, err, i)
			test.Assert(t, stReq.Str == stResp.Str)
		}
	}
	test.Assert(t, cbCount == 30, cbCount)
}

func TestNoCB(t *testing.T) {
	atomic.StoreInt32(&testSTReqCount, -1)
	// retry config
	fp := retry.NewFailurePolicy()
	fp.StopPolicy.CBPolicy.ErrorRate = 0.3
	var opts []client.Option
	opts = append(opts,
		client.WithFailureRetry(fp),
		client.WithCircuitBreaker(circuitbreak.NewCBSuite(genCBKey)),
	)
	cli := getKitexClient(transport.TTHeaderFramed, opts...)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	_, _ = cli.TestSTReq(ctx, stReq)
	for i := 0; i < 250; i++ {
		stResp, err := cli.TestSTReq(ctx, stReq)
		test.Assert(t, err == nil, err, i)
		test.Assert(t, stReq.Str == stResp.Str)
	}
}

func TestNoRetry(t *testing.T) {
	atomic.StoreInt32(&testSTReqCount, -1)
	// retry config
	fp := retry.NewFailurePolicy()
	fp.StopPolicy.CBPolicy.ErrorRate = 0.3
	var opts []client.Option
	opts = append(opts,
		client.WithFailureRetry(fp),
		client.WithRPCTimeout(20*time.Millisecond),
	)
	cli := getKitexClient(transport.Framed, opts...)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	// add a mark to avoid retry
	ctx = metainfo.WithPersistentValue(ctx, retry.TransitKey, strconv.Itoa(1))
	for i := 0; i < 250; i++ {
		stResp, err := cli.TestSTReq(ctx, stReq)
		if i%10 == 0 {
			test.Assert(t, err != nil)
			test.Assert(t, strings.Contains(err.Error(), retryChainStopStr), err)
		} else {
			test.Assert(t, err == nil, err, i)
			test.Assert(t, stReq.Str == stResp.Str)
		}
	}
}

func TestBackupRequest(t *testing.T) {
	atomic.StoreInt32(&testSTReqCount, -1)
	// retry config
	bp := retry.NewBackupPolicy(5)
	var opts []client.Option
	opts = append(opts,
		client.WithBackupRequest(bp),
		client.WithRPCTimeout(40*time.Millisecond),
	)
	cli := getKitexClient(transport.Framed, opts...)
	for i := 0; i < 300; i++ {
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		stReq.Int64 = int64(i)
		stResp, err := cli.TestSTReq(ctx, stReq)
		test.Assert(t, err == nil, err, i, testSTReqCount)
		test.Assert(t, stReq.Str == stResp.Str)
	}
}

func TestServiceCB(t *testing.T) {
	atomic.StoreInt32(&circuitBreakTestCount, -1)
	// retry config
	fp := retry.NewFailurePolicy()
	var opts []client.Option
	opts = append(opts, client.WithFailureRetry(fp))
	opts = append(opts, client.WithRPCTimeout(50*time.Millisecond))
	opts = append(opts, client.WithCircuitBreaker(circuitbreak.NewCBSuite(genCBKey)))
	cli := getKitexClient(transport.TTHeader, opts...)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	cbCount := 0
	for i := 0; i < 300; i++ {
		stResp, err := cli.CircuitBreakTest(ctx, stReq)
		if err != nil {
			test.Assert(t, strings.Contains(err.Error(), "retry circuit break") ||
				strings.Contains(err.Error(), "service circuitbreak"), err, i)
			cbCount++
		} else {
			test.Assert(t, stReq.Str == stResp.Str)
		}
	}
	test.Assert(t, cbCount == 200, cbCount)
}

func TestRetryWithSpecifiedResult(t *testing.T) {
	methodPolicy := make(map[string]bool)
	isErrorRetry := func(err error, ri rpcinfo.RPCInfo) bool {
		if ri.To().Method() == "testObjReq" {
			if te, ok := errors.Unwrap(err).(*remote.TransError); ok && te.TypeID() == 1000 && te.Error() == "retry [biz error]" {
				methodPolicy[ri.To().Method()] = true
				return true
			}
		}
		return false
	}
	isRespRetry := func(resp interface{}, ri rpcinfo.RPCInfo) bool {
		if ri.To().Method() == "testSTReq" {
			if respI, ok1 := resp.(interface{ GetResult() interface{} }); ok1 {
				if r, ok2 := respI.GetResult().(*stability.STResponse); ok2 && r.FlagMsg == retryMsg {
					methodPolicy[ri.To().Method()] = true
					return true
				}
			}
		} else if ri.To().Method() == "testException" {
			teResult := resp.(*stability.STServiceTestExceptionResult)
			if teResult.GetSuccess() != nil {
				teResult.SetStException(nil)
			} else if teResult.IsSetStException() && teResult.StException.Message == retryMsg {
				methodPolicy[ri.To().Method()] = true
				return true
			}
		}
		return false
	}
	isResultRetry := &retry.IsResultRetry{IsErrorRetry: isErrorRetry, IsRespRetry: isRespRetry}
	cli := getKitexClient(
		transport.PurePayload,
		client.WithFailureRetry(retry.NewFailurePolicy()),
		client.WithRetryMethodPolicies(map[string]retry.Policy{
			"testSTReq":        retry.BuildFailurePolicy(retry.NewFailurePolicy()),
			"testObjReq":       retry.BuildFailurePolicy(retry.NewFailurePolicy()),
			"circuitBreakTest": retry.BuildBackupRequest(retry.NewBackupPolicy(10))}),
		client.WithSpecifiedResultRetry(isResultRetry),
		client.WithTransportProtocol(transport.TTHeader),
		client.WithHostPorts(":9002"),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	_, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"])

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	_, err = cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	_, err = cli.TestException(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	_, err = cli.CircuitBreakTest(ctx, stReq)
	test.Assert(t, err == nil, err)

}

func BenchmarkRetryNoCB(b *testing.B) {
	atomic.StoreInt32(&testSTReqCount, -1)
	// retry config
	bp := retry.NewBackupPolicy(3)
	bp.StopPolicy.MaxRetryTimes = 2
	bp.StopPolicy.CBPolicy.ErrorRate = 0.3
	var opts []client.Option
	opts = append(opts, client.WithBackupRequest(bp))
	cli := getKitexClient(transport.PurePayload, opts...)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
			stResp, err := cli.TestSTReq(ctx, stReq)
			test.Assert(b, err == nil, err)
			test.Assert(b, stReq.Str == stResp.Str)
		}
	})
}

// go test -benchmem -bench=BenchmarkThriftCallParallel ./mocks/thriftrpc/normalcall/ | grep '^Benchmark'
func BenchmarkThriftCallParallel(b *testing.B) {
	atomic.StoreInt32(&testExceptionCount, -1)
	fp := retry.NewFailurePolicy()
	fp.StopPolicy.MaxRetryTimes = 5
	var opts []client.Option
	opts = append(opts, client.WithFailureRetry(fp))
	cli := getKitexClient(transport.TTHeader, opts...)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
			_, err := cli.TestException(ctx, stReq)
			test.Assert(b, err != nil)
			test.Assert(b, err.Error() == "STException({Message:mock exception})", err)
			nr, ok := err.(*stability.STException)
			test.Assert(b, ok)
			test.Assert(b, nr.Message == "mock exception")

			err = cli.VisitOneway(ctx, stReq)
			test.Assert(b, err == nil, err)
		}
	})
}
