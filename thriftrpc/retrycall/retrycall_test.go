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
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/circuitbreak"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/retry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/thriftrpc"
)

var retryChainStopStr = "chain stop retry"

var testaddr string

func TestMain(m *testing.M) {
	testaddr = serverutils.NextListenAddr()
	svr1 := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: testaddr,
	}, new(STServiceHandler))
	serverutils.Wait(testaddr)
	m.Run()
	svr1.Stop()
}

func getKitexClient(p transport.Protocol, opts ...client.Option) stservice.Client {
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{testaddr},
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
	t.Parallel()

	reqCount := int32(0)
	cli := getKitexClient(
		transport.TTHeader,
		client.WithRetryContainer(retry.NewRetryContainer()), // use default circuit breaker
		client.WithFailureRetry(retry.NewFailurePolicy()),
		client.WithRPCTimeout(defaultRPCTimeout),
		client.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				count := atomic.AddInt32(&reqCount, 1)
				if count%10 == 0 {
					time.Sleep(defaultSleepTime)
				}
				return nil // no need to send the real request
			}
		}),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	cbCount := 0
	for i := 0; i < 300; i++ {
		_, err := cli.TestSTReq(ctx, stReq)
		if err != nil && strings.Contains(err.Error(), "retry circuit break") {
			cbCount++
		}
	}
	test.Assert(t, reqCount == 300, reqCount)
	test.Assert(t, cbCount == 30, cbCount)
}

func TestRetryCBPercentageLimit(t *testing.T) {
	t.Parallel()

	reqCount := int32(0)
	cli := getKitexClient(
		transport.TTHeaderFramed,
		client.WithRetryContainer(retry.NewRetryContainerWithPercentageLimit()),
		client.WithFailureRetry(retry.NewFailurePolicy()), // cb threshold = 10%
		client.WithRPCTimeout(defaultRPCTimeout),
		client.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				atomic.AddInt32(&reqCount, 1)
				if tm := getSleepTime(ctx); tm > 0 {
					time.Sleep(tm)
				}
				return nil // no need to send a real request
			}
		}),
	)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	ctx = setSkipCounterSleep(ctx)
	cbCount := 0
	for i := 0; i < 300; i++ {
		reqCtx := ctx
		if i%5 == 0 {
			reqCtx = withSleepTime(ctx, defaultSleepTime)
		}
		_, err := cli.TestSTReq(reqCtx, stReq)
		if err != nil && strings.Contains(err.Error(), "retry circuit break") {
			cbCount++
		}
	}
	test.Assert(t, reqCount == 333, reqCount)
	test.Assert(t, cbCount == 58, cbCount)
}

// TestRetryRespOpIsolation fixed in https://github.com/cloudwego/kitex/pull/1194
func TestRetryRespOpIsolation(t *testing.T) {
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	ctx = setSkipCounterSleep(ctx)

	innerCli := getKitexClient(transport.TTHeaderFramed)
	cli := getKitexClient(
		transport.TTHeaderFramed,
		client.WithRetryContainer(retry.NewRetryContainerWithPercentageLimit()),
		client.WithBackupRequest(retry.NewBackupPolicy(20)),
		client.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				ctx = metainfo.WithPersistentValue(ctx, sleepTimeMsKey, "0")
				_, err = innerCli.TestSTReq(ctx, stReq)
				if err != nil {
					t.Errorf("mw resp err: %v", err)
				}
				return next(ctx, req, resp)
			}
		}),
	)

	timer := time.NewTimer(100 * time.Millisecond)
	result := make(chan error, 1)
	go func() {
		ctx = metainfo.WithPersistentValue(ctx, sleepTimeMsKey, "50")
		_, err := cli.TestSTReq(ctx, stReq)
		fmt.Printf("err: %v\n", err)
		result <- err
	}()
	select {
	case <-timer.C:
		t.Errorf("timeout")
	case err := <-result:
		test.Assert(t, err == nil, err)
	}
}

func TestNoCB(t *testing.T) {
	t.Parallel()

	// retry config
	fp := retry.NewFailurePolicy()
	fp.StopPolicy.CBPolicy.ErrorRate = 0.3
	var opts []client.Option
	opts = append(opts,
		client.WithFailureRetry(fp),
		client.WithCircuitBreaker(circuitbreak.NewCBSuite(genCBKey)),
	)
	cli := getKitexClient(transport.TTHeaderFramed, opts...)
	ctx, stReq := thriftrpc.CreateSTRequest(counterNamespace(t))
	stReq.FlagMsg = mockType10PctSleep
	_, _ = cli.TestSTReq(ctx, stReq)
	for i := 0; i < 250; i++ {
		stResp, err := cli.TestSTReq(ctx, stReq)
		test.Assert(t, err == nil, err, i)
		test.Assert(t, stReq.Str == stResp.Str)
	}
}

func TestRetry(t *testing.T) {
	t.Parallel()

	t.Run("TestNoRetry", func(t *testing.T) {
		// retry config
		fp := retry.NewFailurePolicy()
		fp.StopPolicy.CBPolicy.ErrorRate = 0.3
		var opts []client.Option
		opts = append(opts,
			client.WithFailureRetry(fp),
			client.WithRPCTimeout(defaultRPCTimeout),
		)
		cli := getKitexClient(transport.TTHeader, opts...)
		ctx, stReq := thriftrpc.CreateSTRequest(counterNamespace(t))
		stReq.FlagMsg = mockType10PctSleep

		// add a mark to avoid retry
		ctx = metainfo.WithPersistentValue(ctx, retry.TransitKey, "1")
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
	})

	t.Run("TestBackupRequest", func(t *testing.T) {
		// retry config
		bp := retry.NewBackupPolicy(5)
		var opts []client.Option
		opts = append(opts,
			client.WithBackupRequest(bp),
			client.WithRPCTimeout(defaultRPCTimeout),
		)
		basectx := counterNamespace(t)
		cli := getKitexClient(transport.TTHeader, opts...)
		for i := 0; i < 300; i++ {
			ctx, stReq := thriftrpc.CreateSTRequest(basectx)
			stReq.Int64 = int64(i)
			stReq.FlagMsg = mockType10PctSleep
			stResp, err := cli.TestSTReq(ctx, stReq)
			test.Assert(t, err == nil, err, i, getTestCounters(t).STReq)
			test.Assert(t, stReq.Str == stResp.Str)
		}
	})

	t.Run("TestServiceCB", func(t *testing.T) {
		// retry config
		fp := retry.NewFailurePolicy()
		var opts []client.Option
		opts = append(opts, client.WithFailureRetry(fp))
		opts = append(opts, client.WithRPCTimeout(defaultRPCTimeout))
		opts = append(opts, client.WithCircuitBreaker(circuitbreak.NewCBSuite(genCBKey)))
		cli := getKitexClient(transport.TTHeader, opts...)

		ctx, stReq := thriftrpc.CreateSTRequest(counterNamespace(t))
		stReq.FlagMsg = circuitBreak50PCT
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
	})
}

func TestRetryWithSpecifiedResultRetry(t *testing.T) {
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
	isResultRetry := &retry.ShouldResultRetry{ErrorRetry: isErrorRetry, RespRetry: isRespRetry}
	cli := getKitexClient(
		transport.TTHeader,
		client.WithFailureRetry(retry.NewFailurePolicy()),
		client.WithRetryMethodPolicies(map[string]retry.Policy{
			"testSTReq":        retry.BuildFailurePolicy(retry.NewFailurePolicy()),
			"testObjReq":       retry.BuildFailurePolicy(retry.NewFailurePolicy()),
			"circuitBreakTest": retry.BuildBackupRequest(retry.NewBackupPolicy(10)),
		}),
		client.WithSpecifiedResultRetry(isResultRetry),
		client.WithRPCTimeout(defaultRPCTimeout),
	)

	basectx := counterNamespace(t)

	ctx, stReq := thriftrpc.CreateSTRequest(basectx)
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"])

	ctx, objReq := thriftrpc.CreateObjReq(basectx)
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(basectx)
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])

	ctx, stReq = thriftrpc.CreateSTRequest(basectx)
	stReq.FlagMsg = circuitBreakRetrySleep
	_, err = cli.CircuitBreakTest(ctx, stReq)
	test.Assert(t, err == nil, err)
}

// TestFailureRetryWithSpecifiedResult setup the SpecifiedResultRetry with option `NewFailurePolicyWithResultRetry`
func TestFailureRetryWithSpecifiedResultRetry(t *testing.T) {
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
	isResultRetry := &retry.ShouldResultRetry{ErrorRetry: isErrorRetry, RespRetry: isRespRetry}
	cli := getKitexClient(
		transport.TTHeader,
		client.WithFailureRetry(retry.NewFailurePolicy()),
		client.WithRetryMethodPolicies(map[string]retry.Policy{
			"testSTReq":        retry.BuildFailurePolicy(retry.NewFailurePolicy()),
			"testObjReq":       retry.BuildFailurePolicy(retry.NewFailurePolicy()),
			"circuitBreakTest": retry.BuildBackupRequest(retry.NewBackupPolicy(10)),
		}),
		client.WithSpecifiedResultRetry(isResultRetry),
		client.WithRPCTimeout(defaultRPCTimeout),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"])

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = circuitBreakRetrySleep
	_, err = cli.CircuitBreakTest(ctx, stReq)
	test.Assert(t, err == nil, err)
}

func TestRetryWithSpecifiedResultRetryWithCtx(t *testing.T) {
	ctxKeyVal := "ctxKeyVal"
	methodPolicy := make(map[string]bool)
	isErrorRetry := func(ctx context.Context, err error, ri rpcinfo.RPCInfo) bool {
		if ri.To().Method() == "testObjReq" && ctx.Value(ctxKeyVal) == ctxKeyVal {
			if te, ok := errors.Unwrap(err).(*remote.TransError); ok && te.TypeID() == 1000 && te.Error() == "retry [biz error]" {
				methodPolicy[ri.To().Method()] = true
				return true
			}
		}
		return false
	}
	isRespRetry := func(ctx context.Context, resp interface{}, ri rpcinfo.RPCInfo) bool {
		if ri.To().Method() == "testSTReq" && ctx.Value(ctxKeyVal) == ctxKeyVal {
			if respI, ok1 := resp.(interface{ GetResult() interface{} }); ok1 {
				if r, ok2 := respI.GetResult().(*stability.STResponse); ok2 && r.FlagMsg == retryMsg {
					methodPolicy[ri.To().Method()] = true
					return true
				}
			}
		} else if ri.To().Method() == "testException" && ctx.Value(ctxKeyVal) == ctxKeyVal {
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
	isResultRetry := &retry.ShouldResultRetry{ErrorRetryWithCtx: isErrorRetry, RespRetryWithCtx: isRespRetry}
	cli := getKitexClient(
		transport.TTHeader,
		client.WithFailureRetry(retry.NewFailurePolicy()),
		client.WithRetryMethodPolicies(map[string]retry.Policy{
			"testSTReq":        retry.BuildFailurePolicy(retry.NewFailurePolicy()),
			"testObjReq":       retry.BuildFailurePolicy(retry.NewFailurePolicy()),
			"circuitBreakTest": retry.BuildBackupRequest(retry.NewBackupPolicy(10)),
		}),
		client.WithSpecifiedResultRetry(isResultRetry),
		client.WithRPCTimeout(defaultRPCTimeout),
	)

	ctx := context.WithValue(context.Background(), ctxKeyVal, ctxKeyVal)
	ctx, stReq := thriftrpc.CreateSTRequest(ctx)
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"])

	ctx, objReq := thriftrpc.CreateObjReq(ctx)
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(ctx)
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])

	ctx, stReq = thriftrpc.CreateSTRequest(ctx)
	stReq.FlagMsg = circuitBreakRetrySleep
	_, err = cli.CircuitBreakTest(ctx, stReq)
	test.Assert(t, err == nil, err)
}

func TestRetryWithSpecifiedResultRetryWithOldAndNew(t *testing.T) {
	isErrorRetryNew := func(ctx context.Context, err error, ri rpcinfo.RPCInfo) bool {
		return true
	}
	isErrorRetryOld := func(err error, ri rpcinfo.RPCInfo) bool {
		return false
	}
	isRespRetryNew := func(ctx context.Context, resp interface{}, ri rpcinfo.RPCInfo) bool {
		return true
	}
	isRespRetryOld := func(resp interface{}, ri rpcinfo.RPCInfo) bool {
		return false
	}
	isResultRetry := &retry.ShouldResultRetry{
		ErrorRetryWithCtx: isErrorRetryNew, RespRetryWithCtx: isRespRetryNew,
		ErrorRetry: isErrorRetryOld, RespRetry: isRespRetryOld}
	cli := getKitexClient(
		transport.TTHeader,
		client.WithFailureRetry(retry.NewFailurePolicy()),
		client.WithSpecifiedResultRetry(isResultRetry),
		client.WithTransportProtocol(transport.TTHeader),
		client.WithRPCTimeout(defaultRPCTimeout),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeCustomizedResp
	resp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, resp.FlagMsg == "success")

	ctx, req := thriftrpc.CreateObjReq(context.Background())
	req.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, req)
	test.Assert(t, err == nil, err)
}

// TestFailureRetryWithSpecifiedResult setup the SpecifiedResultRetry with option `NewFailurePolicyWithResultRetry`
func TestFailureRetryWithSpecifiedResultRetryWithCtx(t *testing.T) {
	ctxKeyVal := "ctxKeyVal"
	methodPolicy := make(map[string]bool)
	isErrorRetry := func(ctx context.Context, err error, ri rpcinfo.RPCInfo) bool {
		if ri.To().Method() == "testObjReq" && ctx.Value(ctxKeyVal) == ctxKeyVal {
			if te, ok := errors.Unwrap(err).(*remote.TransError); ok && te.TypeID() == 1000 && te.Error() == "retry [biz error]" {
				methodPolicy[ri.To().Method()] = true
				return true
			}
		}
		return false
	}
	isRespRetry := func(ctx context.Context, resp interface{}, ri rpcinfo.RPCInfo) bool {
		if ri.To().Method() == "testSTReq" && ctx.Value(ctxKeyVal) == ctxKeyVal {
			if respI, ok1 := resp.(interface{ GetResult() interface{} }); ok1 {
				if r, ok2 := respI.GetResult().(*stability.STResponse); ok2 && r.FlagMsg == retryMsg {
					methodPolicy[ri.To().Method()] = true
					return true
				}
			}
		} else if ri.To().Method() == "testException" && ctx.Value(ctxKeyVal) == ctxKeyVal {
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
	isResultRetry := &retry.ShouldResultRetry{ErrorRetryWithCtx: isErrorRetry, RespRetryWithCtx: isRespRetry}
	cli := getKitexClient(
		transport.TTHeader,
		client.WithFailureRetry(retry.NewFailurePolicyWithResultRetry(isResultRetry)),
		client.WithRetryMethodPolicies(map[string]retry.Policy{
			"testSTReq":        retry.BuildFailurePolicy(retry.NewFailurePolicyWithResultRetry(isResultRetry)),
			"testObjReq":       retry.BuildFailurePolicy(retry.NewFailurePolicyWithResultRetry(isResultRetry)),
			"circuitBreakTest": retry.BuildBackupRequest(retry.NewBackupPolicy(10)),
		}),
		client.WithRPCTimeout(defaultRPCTimeout),
	)

	ctx := context.WithValue(context.Background(), ctxKeyVal, ctxKeyVal)
	ctx, stReq := thriftrpc.CreateSTRequest(ctx)
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"])

	ctx, objReq := thriftrpc.CreateObjReq(ctx)
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(ctx)
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])

	ctx, stReq = thriftrpc.CreateSTRequest(ctx)
	stReq.FlagMsg = circuitBreakRetrySleep
	_, err = cli.CircuitBreakTest(ctx, stReq)
	test.Assert(t, err == nil, err)
}

func TestRetryNotifyHasOldResultRetry(t *testing.T) {
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
	isResultRetry := &retry.ShouldResultRetry{ErrorRetry: isErrorRetry, RespRetry: isRespRetry}
	rc := retry.NewRetryContainer()
	rc.NotifyPolicyChange("*", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("testSTReq", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("testObjReq", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("circuitBreakTest", retry.BuildBackupRequest(retry.NewBackupPolicy(10)))

	cli := getKitexClient(
		transport.TTHeader,
		client.WithRetryContainer(rc),
		client.WithSpecifiedResultRetry(isResultRetry),
		client.WithRPCTimeout(defaultRPCTimeout),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"], methodPolicy)

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = circuitBreakRetrySleep
	_, err = cli.CircuitBreakTest(ctx, stReq)
	test.Assert(t, err == nil, err)

	// modify the retry policy, but the result retry is still effective
	rc.NotifyPolicyChange("*", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("testSTReq", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("testObjReq", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	methodPolicy["testSTReq"] = false
	methodPolicy["testObjReq"] = false
	methodPolicy["testException"] = false

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"], methodPolicy)

	ctx, objReq = thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])
}

func TestRetryNotifyHasNewResultRetry(t *testing.T) {
	methodPolicy := make(map[string]bool)
	isErrorRetry := func(ctx context.Context, err error, ri rpcinfo.RPCInfo) bool {
		if ri.To().Method() == "testObjReq" {
			if te, ok := errors.Unwrap(err).(*remote.TransError); ok && te.TypeID() == 1000 && te.Error() == "retry [biz error]" {
				methodPolicy[ri.To().Method()] = true
				return true
			}
		}
		return false
	}
	isRespRetry := func(ctx context.Context, resp interface{}, ri rpcinfo.RPCInfo) bool {
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
	isResultRetry := &retry.ShouldResultRetry{ErrorRetryWithCtx: isErrorRetry, RespRetryWithCtx: isRespRetry}
	rc := retry.NewRetryContainer()
	rc.NotifyPolicyChange("*", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("testSTReq", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("testObjReq", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("circuitBreakTest", retry.BuildBackupRequest(retry.NewBackupPolicy(10)))

	cli := getKitexClient(
		transport.TTHeader,
		client.WithRetryContainer(rc),
		client.WithSpecifiedResultRetry(isResultRetry),
		client.WithRPCTimeout(defaultRPCTimeout),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"], methodPolicy)

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = circuitBreakRetrySleep
	_, err = cli.CircuitBreakTest(ctx, stReq)
	test.Assert(t, err == nil, err)

	// modify the retry policy, but the result retry is still effective
	rc.NotifyPolicyChange("*", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("testSTReq", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	rc.NotifyPolicyChange("testObjReq", retry.BuildFailurePolicy(retry.NewFailurePolicy()))
	methodPolicy["testSTReq"] = false
	methodPolicy["testObjReq"] = false
	methodPolicy["testException"] = false

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"], methodPolicy)

	ctx, objReq = thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])
}

// TestRetryWithCallOptHasOldResultRetry setup the retry policy with callopt.
func TestRetryWithCallOptHasOldResultRetry(t *testing.T) {
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
	isResultRetry := &retry.ShouldResultRetry{ErrorRetry: isErrorRetry, RespRetry: isRespRetry}
	cli := getKitexClient(
		transport.TTHeader,
		// setup WithBackupRequest is to check the callopt priority is higher
		client.WithBackupRequest(retry.NewBackupPolicy(10)),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err := cli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildFailurePolicy(retry.NewFailurePolicyWithResultRetry(isResultRetry))))
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"])

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq, callopt.WithRetryPolicy(retry.BuildFailurePolicy(retry.NewFailurePolicyWithResultRetry(isResultRetry))))
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq, callopt.WithRetryPolicy(retry.BuildFailurePolicy(retry.NewFailurePolicyWithResultRetry(isResultRetry))))
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = circuitBreakRetrySleep
	_, err = cli.CircuitBreakTest(ctx, stReq, callopt.WithRetryPolicy(retry.BuildBackupRequest(retry.NewBackupPolicy(10))))
	test.Assert(t, err == nil, err)
}

// TestRetryWithCallOptHasNewResultRetry setup the retry policy with callopt.
// test logic same with TestRetryWithCallOptHasOldResultRetry
func TestRetryWithCallOptHasNewResultRetry(t *testing.T) {
	methodPolicy := make(map[string]bool)
	isErrorRetry := func(ctx context.Context, err error, ri rpcinfo.RPCInfo) bool {
		if ri.To().Method() == "testObjReq" {
			if te, ok := errors.Unwrap(err).(*remote.TransError); ok && te.TypeID() == 1000 && te.Error() == "retry [biz error]" {
				methodPolicy[ri.To().Method()] = true
				return true
			}
		}
		return false
	}
	isRespRetry := func(ctx context.Context, resp interface{}, ri rpcinfo.RPCInfo) bool {
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
	isResultRetry := &retry.ShouldResultRetry{ErrorRetryWithCtx: isErrorRetry, RespRetryWithCtx: isRespRetry}
	cli := getKitexClient(
		transport.TTHeader,
		// setup WithBackupRequest is to check the callopt priority is higher
		client.WithBackupRequest(retry.NewBackupPolicy(10)),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeCustomizedResp
	_, err := cli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildFailurePolicy(retry.NewFailurePolicyWithResultRetry(isResultRetry))))
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testSTReq"])

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestObjReq(ctx, objReq, callopt.WithRetryPolicy(retry.BuildFailurePolicy(retry.NewFailurePolicyWithResultRetry(isResultRetry))))
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testObjReq"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = mockTypeNonRetryReturnError
	_, err = cli.TestException(ctx, stReq, callopt.WithRetryPolicy(retry.BuildFailurePolicy(retry.NewFailurePolicyWithResultRetry(isResultRetry))))
	test.Assert(t, err == nil, err)
	test.Assert(t, methodPolicy["testException"])

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	stReq.FlagMsg = circuitBreakRetrySleep
	_, err = cli.CircuitBreakTest(ctx, stReq, callopt.WithRetryPolicy(retry.BuildBackupRequest(retry.NewBackupPolicy(10))))
	test.Assert(t, err == nil, err)
}

func TestRetryForTimeout(t *testing.T) {
	isErrorRetry := func(err error, ri rpcinfo.RPCInfo) bool {
		if ri.To().Method() == "testSTReq" {
			if te, ok := errors.Unwrap(err).(*remote.TransError); ok && te.TypeID() == 1000 && te.Error() == "retry [biz error]" {
				return true
			}
		}
		return false
	}
	mockTimeoutMW := func(endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
				time.Sleep(defaultSleepTime)
			}
			return
		}
	}
	// case 1: timeout retry is effective by default
	isResultRetry := &retry.ShouldResultRetry{ErrorRetry: isErrorRetry}
	cli := getKitexClient(
		transport.PurePayload,
		client.WithMiddleware(mockTimeoutMW),
		client.WithRPCTimeout(defaultRPCTimeout),
		client.WithFailureRetry(retry.NewFailurePolicyWithResultRetry(isResultRetry)),
	)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	_, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)

	// case 2: timeout retry is not effective if set NotRetryForTimeout as true
	// callopt will reset the client-level option
	isResultRetry = &retry.ShouldResultRetry{ErrorRetry: isErrorRetry, NotRetryForTimeout: true}
	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	_, err = cli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildFailurePolicy(retry.NewFailurePolicyWithResultRetry(isResultRetry))))
	test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout), err)

	// case 3: use client-level option, timeout retry is not effective if set NotRetryForTimeout as true
	isResultRetry = &retry.ShouldResultRetry{ErrorRetry: isErrorRetry, NotRetryForTimeout: true}
	cli = getKitexClient(
		transport.PurePayload,
		client.WithMiddleware(mockTimeoutMW),
		client.WithRPCTimeout(defaultRPCTimeout),
		client.WithFailureRetry(retry.NewFailurePolicyWithResultRetry(isResultRetry)),
	)
	isResultRetry = &retry.ShouldResultRetry{ErrorRetry: isErrorRetry, NotRetryForTimeout: true}
	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	_, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout), err)

}

func BenchmarkRetryNoCB(b *testing.B) {
	// retry config
	bp := retry.NewBackupPolicy(3)
	bp.StopPolicy.MaxRetryTimes = 2
	bp.StopPolicy.CBPolicy.ErrorRate = 0.3
	var opts []client.Option
	opts = append(opts, client.WithBackupRequest(bp))
	cli := getKitexClient(transport.PurePayload, opts...)
	basectx := counterNamespace(b)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, stReq := thriftrpc.CreateSTRequest(basectx)
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
