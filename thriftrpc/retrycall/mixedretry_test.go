// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
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
	"strings"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/retry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thriftrpc"
)

// closeTo returns true if actual is within 20% of expected
func closeTo(actual, expected time.Duration) bool {
	diff := actual - expected
	if diff < 0 {
		diff = -diff
	}
	return diff < expected/5
}

// TestMockCaseWithDiffRetry compares retry strategies when req1 times out but req2 succeeds.
//
// Setup: req1=120ms, req2=80ms, rpcTimeout=100ms, backupDelay=60ms, maxRetry=2
//
// Results:
//   - Mixed:   Success in 140ms (backupDelay + req2)
//   - Failure: Success in 180ms (rpcTimeout + req2)
//   - Backup:  Timeout in 120ms (2 * backupDelay, both requests timeout)
func TestMockCaseWithDiffRetry(t *testing.T) {
	t.Parallel() // slow test, parallel for reducing test time

	// backupDelay < req2Sleep < rpcTimeout < req1Sleep
	rpcTimeout := 100 * time.Millisecond
	req1Sleep := rpcTimeout + 20*time.Millisecond  // 120ms
	req2Sleep := rpcTimeout - 20*time.Millisecond  // 80ms
	backupDelay := req2Sleep - 20*time.Millisecond // 60ms

	controlCostMW := func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, args, result interface{}) (err error) {
			_, exit := rpcinfo.GetRPCInfo(ctx).To().Tag(rpcinfo.RetryTag)
			if !exit {
				ctx = withSleepTime(ctx, req1Sleep)
			} else {
				ctx = withSleepTime(ctx, req2Sleep)
			}
			return next(ctx, args, result)
		}
	}

	t.Run("mixed retry", func(t *testing.T) {
		mp := retry.NewMixedPolicy(uint32(backupDelay / time.Millisecond))
		mp.WithMaxRetryTimes(2)
		rCli := getKitexClient(
			transport.TTHeader,
			client.WithMixedRetry(mp),
			client.WithRPCTimeout(rpcTimeout),
			client.WithMiddleware(controlCostMW),
		)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		start := time.Now()
		_, err := rCli.TestSTReq(ctx, stReq)
		cost := time.Since(start)
		test.Assert(t, err == nil, err)
		test.Assert(t, closeTo(cost, backupDelay+req2Sleep), cost)
	})

	t.Run("failure retry", func(t *testing.T) {
		fp := retry.NewFailurePolicy()
		fp.WithMaxRetryTimes(2)
		rCli := getKitexClient(
			transport.TTHeader,
			client.WithFailureRetry(fp),
			client.WithRPCTimeout(rpcTimeout),
			client.WithMiddleware(controlCostMW),
		)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		start := time.Now()
		_, err := rCli.TestSTReq(ctx, stReq)
		cost := time.Since(start)
		test.Assert(t, err == nil, err)
		test.Assert(t, closeTo(cost, rpcTimeout+req2Sleep), cost)
	})

	t.Run("backup request", func(t *testing.T) {
		bp := retry.NewBackupPolicy(uint32(backupDelay / time.Millisecond))
		bp.WithMaxRetryTimes(2)
		rCli := getKitexClient(
			transport.TTHeader,
			client.WithBackupRequest(bp),
			client.WithRPCTimeout(rpcTimeout),
			client.WithMiddleware(controlCostMW),
		)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		start := time.Now()
		_, err := rCli.TestSTReq(ctx, stReq)
		cost := time.Since(start)
		test.Assert(t, err != nil, err)
		test.Assert(t, closeTo(cost, 2*backupDelay), cost)
	})
}

// TestMockCaseWithTimeoutWithDiffRetry compares retry strategies when all requests timeout.
//
// Setup: reqSleep=130ms, rpcTimeout=100ms, backupDelay=70ms, maxRetry=2
//
// Results:
//   - Mixed:   Timeout in 200ms (2 * rpcTimeout)
//   - Failure: Timeout in 300ms (3 * rpcTimeout)
//   - Backup:  Timeout in 100ms (1 * rpcTimeout, no retry triggered)
func TestMockCaseWithTimeoutWithDiffRetry(t *testing.T) {
	t.Parallel() // slow test, parallel for reducing test time

	// backupDelay < rpcTimeout < reqSleep
	rpcTimeout := 100 * time.Millisecond
	backupDelay := rpcTimeout - 30*time.Millisecond // 70ms
	reqSleep := rpcTimeout + 30*time.Millisecond    // 130ms

	controlCostMW := func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, args, result interface{}) (err error) {
			ctx = withSleepTime(ctx, reqSleep)
			return next(ctx, args, result)
		}
	}
	rCli := getKitexClient(
		transport.TTHeader,
		client.WithRPCTimeout(rpcTimeout),
		client.WithMiddleware(controlCostMW),
	)

	t.Run("mixed retry", func(t *testing.T) {
		mp := retry.NewMixedPolicy(uint32(backupDelay / time.Millisecond))
		mp.WithMaxRetryTimes(2)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		start := time.Now()
		_, err := rCli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildMixedPolicy(mp)))
		cost := time.Since(start)
		test.Assert(t, err != nil, err)
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		test.Assert(t, closeTo(cost, 2*rpcTimeout), cost)
	})

	t.Run("failure retry", func(t *testing.T) {
		fp := retry.NewFailurePolicy()
		fp.WithMaxRetryTimes(2)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		start := time.Now()
		_, err := rCli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildFailurePolicy(fp)))
		cost := time.Since(start)
		test.Assert(t, err != nil, err)
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		test.Assert(t, closeTo(cost, 3*rpcTimeout), cost)
	})

	t.Run("backup request", func(t *testing.T) {
		bp := retry.NewBackupPolicy(100)
		bp.WithMaxRetryTimes(2)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		start := time.Now()
		_, err := rCli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildBackupRequest(bp)))
		cost := time.Since(start)
		test.Assert(t, err != nil, err)
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		test.Assert(t, closeTo(cost, rpcTimeout), cost)
	})
}

// TestMockCase4WithDiffRetry compares retry strategies with result-based retry.
//
// Setup: each request costs 250ms, backupDelay=100ms, maxRetry=3
// Server returns: req1="11111"(retry), req2="11112"(retry), req3="0"(success)
//
// Results:
//   - Mixed:   Success in 450ms (backup fires at 100ms/200ms, failure retry on bad result)
//   - Failure: Success in 750ms (sequential retries: 250ms * 3)
//   - Backup:  BizError in 250ms (no result-based retry, returns first response)
func TestMockCase4WithDiffRetry(t *testing.T) {
	t.Parallel()

	t.Run("mixed retry", func(t *testing.T) {
		mp := retry.NewMixedPolicy(100)
		mp.WithMaxRetryTimes(3)
		rCli := getKitexClient(
			transport.TTHeader,
			client.WithSpecifiedResultRetry(resultRetry),
			client.WithMixedRetry(mp),
			client.WithMiddleware(controlRespMW),
		)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		stReq.FlagMsg = mockTypeCustomizedResp
		start := time.Now()
		resp, err := rCli.TestSTReq(ctx, stReq)
		cost := time.Since(start)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.FlagMsg == "0", resp.FlagMsg)
		test.Assert(t, closeTo(cost, 450*time.Millisecond), cost)
	})

	t.Run("failure retry", func(t *testing.T) {
		fp := retry.NewFailurePolicy()
		fp.WithMaxRetryTimes(3)
		rCli := getKitexClient(
			transport.TTHeader,
			client.WithSpecifiedResultRetry(resultRetry),
			client.WithFailureRetry(fp),
			client.WithMiddleware(controlRespMW),
		)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		stReq.FlagMsg = mockTypeCustomizedResp
		start := time.Now()
		resp, err := rCli.TestSTReq(ctx, stReq)
		cost := time.Since(start)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.FlagMsg == "0", resp.FlagMsg)
		test.Assert(t, closeTo(cost, 750*time.Millisecond), cost)
	})

	t.Run("backup request", func(t *testing.T) {
		bp := retry.NewBackupPolicy(100)
		bp.WithMaxRetryTimes(2)
		rCli := getKitexClient(
			transport.TTHeader,
			// callopt policy overrides client options, so resultRetry won't take effect
			client.WithSpecifiedResultRetry(resultRetry),
			client.WithFailureRetry(retry.NewFailurePolicy()),
			client.WithMiddleware(controlRespMW),
		)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		stReq.FlagMsg = mockTypeCustomizedResp
		start := time.Now()
		resp, err := rCli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildBackupRequest(bp)))
		cost := time.Since(start)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.FlagMsg != "0")
		test.Assert(t, closeTo(cost, 250*time.Millisecond), cost)
	})
}

func TestMixedRetryWithDiffConfigurationMethod(t *testing.T) {
	t.Parallel()

	rpcCall := func(cli stservice.Client) {
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		stReq.FlagMsg = mockTypeCustomizedResp
		start := time.Now()
		resp, err := cli.TestSTReq(ctx, stReq)
		cost := time.Since(start)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.FlagMsg == "0", resp.FlagMsg)
		test.Assert(t, closeTo(cost, 450*time.Millisecond), cost)
	}

	t.Run("mixed retry with NewMixedPolicy", func(t *testing.T) {
		mp := retry.NewMixedPolicy(100)
		mp.WithMaxRetryTimes(3)

		rCli := getKitexClient(
			transport.TTHeader,
			client.WithMixedRetry(mp),
			client.WithSpecifiedResultRetry(resultRetry),
			client.WithMiddleware(controlRespMW),
		)
		rpcCall(rCli)
	})

	t.Run("mixed retry with NewMixedPolicyWithResultRetry", func(t *testing.T) {
		mp := retry.NewMixedPolicyWithResultRetry(100, resultRetry)
		mp.WithMaxRetryTimes(3)

		rCli := getKitexClient(
			transport.TTHeader,
			client.WithMixedRetry(mp),
			client.WithMiddleware(controlRespMW),
		)
		rpcCall(rCli)
	})

	t.Run("mixed retry with callop", func(t *testing.T) {
		mp := retry.NewMixedPolicyWithResultRetry(100, resultRetry)
		mp.WithMaxRetryTimes(3)

		rCli := getKitexClient(
			transport.TTHeader,
			client.WithMiddleware(controlRespMW),
		)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		stReq.FlagMsg = mockTypeCustomizedResp
		start := time.Now()
		resp, err := rCli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildMixedPolicy(mp)))
		cost := time.Since(start)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.FlagMsg == "0", resp.FlagMsg)
		test.Assert(t, closeTo(cost, 450*time.Millisecond), cost)
	})

	t.Run("mixed retry with NotifyPolicyChange", func(t *testing.T) {
		mp := retry.NewMixedPolicy(100)
		mp.WithMaxRetryTimes(3)
		rc := retry.NewRetryContainer()
		rc.NotifyPolicyChange("*", retry.BuildMixedPolicy(mp))

		rCli := getKitexClient(
			transport.TTHeader,
			client.WithRetryContainer(rc),
			client.WithSpecifiedResultRetry(resultRetry),
			client.WithMiddleware(controlRespMW),
		)
		rpcCall(rCli)
	})
}

func TestMixedRetryWithNotifyPolicyChange(t *testing.T) {
	t.Parallel()

	t.Run("mixed retry with result retry", func(t *testing.T) {
		mp := retry.NewMixedPolicy(100)
		mp.WithMaxRetryTimes(3)
		rc := retry.NewRetryContainer()
		rc.NotifyPolicyChange("*", retry.BuildMixedPolicy(mp))

		rCli := getKitexClient(
			transport.TTHeader,
			client.WithRetryContainer(rc),
			client.WithSpecifiedResultRetry(resultRetry),
			client.WithMiddleware(controlRespMW),
		)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		stReq.FlagMsg = mockTypeCustomizedResp
		start := time.Now()
		resp, err := rCli.TestSTReq(ctx, stReq)
		cost := time.Since(start)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.FlagMsg == "0", resp.FlagMsg)
		test.Assert(t, closeTo(cost, 450*time.Millisecond), cost)
	})

	t.Run("mixed retry with error retry", func(t *testing.T) {
		mp := retry.NewMixedPolicy(100)
		mp.WithMaxRetryTimes(3)
		rc := retry.NewRetryContainer()
		rc.NotifyPolicyChange("testSTReq", retry.BuildMixedPolicy(mp))

		rCli := getKitexClient(
			transport.TTHeader,
			client.WithRetryContainer(rc),
			client.WithSpecifiedResultRetry(errRetry),
			client.WithMiddleware(controlErrMW),
		)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		stReq.FlagMsg = mockTypeReturnTransErr
		resp, err := rCli.TestSTReq(ctx, stReq)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.FlagMsg == "0", resp.FlagMsg)
	})
}

func BenchmarkMixedRetry(b *testing.B) {
	mp := retry.NewMixedPolicyWithResultRetry(100, errRetry)
	mp.WithMaxRetryTimes(3)
	rCli := getKitexClient(
		transport.TTHeader,
		client.WithMixedRetry(mp),
		client.WithMiddleware(controlErrMW),
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
			stReq.FlagMsg = mockTypeReturnTransErr
			start := time.Now()
			resp, err := rCli.TestSTReq(ctx, stReq)
			cost := time.Since(start)
			if err != nil {
				// mock retry will trigger retry circuit break
				test.Assert(b, strings.Contains(err.Error(), "retry circuit break"))
			} else {
				test.Assert(b, err == nil, err)
				test.Assert(b, resp.FlagMsg == "0", resp.FlagMsg, cost.Milliseconds())
			}
		}
	})
}

// controlRespMW simulates server responses: req1="11111", req2="11112", req3="0"
var controlRespMW = func(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, args, result interface{}) (err error) {
		ctx = metainfo.WithValue(ctx, sleepTimeMsKey, "250")
		retryCount, exist := rpcinfo.GetRPCInfo(ctx).To().Tag(rpcinfo.RetryTag)
		if !exist {
			ctx = metainfo.WithValue(ctx, respFlagMsgKey, "11111")
		} else if retryCount == "2" {
			ctx = metainfo.WithValue(ctx, respFlagMsgKey, "0")
		} else {
			ctx = metainfo.WithValue(ctx, respFlagMsgKey, "11112")
		}
		return next(ctx, args, result)
	}
}

// resultRetry retries when response FlagMsg is "11111" or "11112"
var resultRetry = &retry.ShouldResultRetry{RespRetryWithCtx: func(ctx context.Context, resp interface{}, ri rpcinfo.RPCInfo) bool {
	if respI, ok1 := resp.(interface{ GetResult() interface{} }); ok1 {
		if r, ok2 := respI.GetResult().(*stability.STResponse); ok2 {
			return r.FlagMsg == "11111" || r.FlagMsg == "11112"
		}
	}
	return false
}}

// controlErrMW simulates server returning TransError until 3rd request
var controlErrMW = func(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, args, result interface{}) (err error) {
		ctx = metainfo.WithValue(ctx, sleepTimeMsKey, "250")
		retryCount, exist := rpcinfo.GetRPCInfo(ctx).To().Tag(rpcinfo.RetryTag)
		if !exist || retryCount != "2" {
			ctx = metainfo.WithValue(ctx, respFlagMsgKey, retryMsg)
		} else {
			ctx = metainfo.WithValue(ctx, respFlagMsgKey, "0")
		}
		return next(ctx, args, result)
	}
}

// errRetry retries when error is TransError with retryTransErrCode
var errRetry = &retry.ShouldResultRetry{ErrorRetryWithCtx: func(ctx context.Context, err error, ri rpcinfo.RPCInfo) bool {
	var te *remote.TransError
	if errors.As(err, &te) && te.TypeID() == retryTransErrCode {
		return true
	}
	return false
}}
