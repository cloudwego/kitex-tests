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
	"math"
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

func durAbs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// Assuming:
// - the first request costs 60ms (req1Sleep)
// - the second request costs 40ms (req2Sleep)
// Configuration: MaxRetryTimes=2, rpcTimeout=50ms, backupDelay=30ms
// - Mixed Retry: Success, cost backupDelay+req2Sleep
// - Failure Retry: Success, cost rpcTimeout+req2Sleep
// - Backup Retry: Failure, cost 2*backupDelay
func TestMockCaseWithDiffRetry(t *testing.T) {
	// backupDelay < req2Sleep < rpcTimeout < req1Sleep
	rpcTimeout := 50 * time.Millisecond
	req1Sleep := rpcTimeout + 10*time.Millisecond  // 60ms
	req2Sleep := rpcTimeout - 10*time.Millisecond  // 40ms
	backupDelay := req2Sleep - 10*time.Millisecond // 30ms

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

	// mixed retry will success, latency is lowest
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
		test.Assert(t, durAbs(cost-(backupDelay+req2Sleep)) < 10*time.Millisecond, cost)
	})

	// failure retry will success, but latency is more than mixed retry
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
		test.Assert(t, durAbs(cost-(rpcTimeout+req2Sleep)) < 10*time.Millisecond, cost)
	})

	// backup request will failure
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
		test.Assert(t, durAbs(cost-2*backupDelay) < 10*time.Millisecond, cost)
	})
}

// Assuming all request timeout
// Configuration: rpcTimeout=100ms、MaxRetryTimes=2 BackupDelay=100ms
// - Mixed Retry: Failure, cost 2*rpcTimeout
// - Failure Retry: Failure, cost 3*rpcTimeout
// - Backup Retry: Failure, cost 1*rpcTimeout
func TestMockCaseWithTimeoutWithDiffRetry(t *testing.T) {
	// backupDelay < rpcTimeout < reqSleep
	rpcTimeout := 30 * time.Millisecond
	backupDelay := rpcTimeout - 10*time.Millisecond // 20ms
	reqSleep := rpcTimeout + 10*time.Millisecond    // 40ms

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

	// mixed retry will success, cost is least
	t.Run("mixed retry", func(t *testing.T) {
		mp := retry.NewMixedPolicy(uint32(backupDelay / time.Millisecond))
		mp.WithMaxRetryTimes(2)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		start := time.Now()
		_, err := rCli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildMixedPolicy(mp)))
		cost := time.Since(start)
		test.Assert(t, err != nil, err)
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		test.Assert(t, durAbs(cost-2*rpcTimeout) < 10*time.Millisecond, cost)
	})

	// failure retry will success, but cost is more than mixed retry
	t.Run("failure retry", func(t *testing.T) {
		fp := retry.NewFailurePolicy()
		fp.WithMaxRetryTimes(2)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		start := time.Now()
		_, err := rCli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildFailurePolicy(fp)))
		cost := time.Since(start)
		test.Assert(t, err != nil, err)
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		test.Assert(t, durAbs(cost-3*rpcTimeout) < 10*time.Millisecond, cost)
	})

	// backup request will failure
	t.Run("backup request", func(t *testing.T) {
		bp := retry.NewBackupPolicy(100)
		bp.WithMaxRetryTimes(2)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		start := time.Now()
		_, err := rCli.TestSTReq(ctx, stReq, callopt.WithRetryPolicy(retry.BuildBackupRequest(bp)))
		cost := time.Since(start)
		test.Assert(t, err != nil, err)
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		test.Assert(t, durAbs(cost-rpcTimeout) < 10*time.Millisecond, cost)
	})
}

// Assuming resp.FlagMsg=11111/11112 needs to be retried,
//
//	the first reply is resp.FlagMsg=11111, it costs 250ms,
//	the second reply is resp.FlagMsg=11112, it costs 250ms,
//	the third reply is resp.FlagMsg=0, it costs 250ms,
//
// Configuration: MaxRetryTimes=3 BackupDelay=100ms
// - Mixed Retry: Success, cost 450ms, two backup retry, and one failure retry
// - Failure Retry: Success, cost 750ms
// - Backup Retry: Biz Error, cost 250ms
func TestMockCase4WithDiffRetry(t *testing.T) {
	t.Parallel()

	// mixed retry will success, cost is least
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
		test.Assert(t, math.Abs(float64(cost.Milliseconds())-450.0) < 50.0, cost.Milliseconds())
	})

	// failure retry will success, but cost is more than mixed retry
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
		test.Assert(t, math.Abs(float64(cost.Milliseconds())-750.0) < 50.0, cost.Milliseconds())
	})

	// backup request will failure
	t.Run("backup request", func(t *testing.T) {
		bp := retry.NewBackupPolicy(100)
		bp.WithMaxRetryTimes(2)
		rCli := getKitexClient(
			transport.TTHeader,
			// this won't take effect, because callopt policy is high priority
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
		test.Assert(t, math.Abs(float64(cost.Milliseconds())-250.0) < 50.0, cost.Milliseconds())
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
		test.Assert(t, math.Abs(float64(cost.Milliseconds())-450.0) < 50.0, cost.Milliseconds())
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
		test.Assert(t, math.Abs(float64(cost.Milliseconds())-450.0) < 50.0, cost.Milliseconds())
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
		test.Assert(t, math.Abs(float64(cost.Milliseconds())-450.0) < 50.0, cost.Milliseconds())
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

var controlRespMW = func(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, args, result interface{}) (err error) {
		ctx = metainfo.WithValue(ctx, sleepTimeMsKey, "250") //ms
		retryCount, exit := rpcinfo.GetRPCInfo(ctx).To().Tag(rpcinfo.RetryTag)
		if !exit {
			ctx = metainfo.WithValue(ctx, respFlagMsgKey, "11111") //ms
		} else {
			if retryCount == "2" {
				ctx = metainfo.WithValue(ctx, respFlagMsgKey, "0") //ms
			} else {
				ctx = metainfo.WithValue(ctx, respFlagMsgKey, "11112") //ms
			}
		}
		err = next(ctx, args, result)
		return
	}
}

var resultRetry = &retry.ShouldResultRetry{RespRetryWithCtx: func(ctx context.Context, resp interface{}, ri rpcinfo.RPCInfo) bool {
	if respI, ok1 := resp.(interface{ GetResult() interface{} }); ok1 {
		if r, ok2 := respI.GetResult().(*stability.STResponse); ok2 {
			if r.FlagMsg == "11111" || r.FlagMsg == "11112" {
				return true
			}
		}
	}
	return false
}}

var controlErrMW = func(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, args, result interface{}) (err error) {
		ctx = metainfo.WithValue(ctx, sleepTimeMsKey, "250") //ms
		retryCount, exit := rpcinfo.GetRPCInfo(ctx).To().Tag(rpcinfo.RetryTag)
		if !exit {
			ctx = metainfo.WithValue(ctx, respFlagMsgKey, retryMsg)
		} else {
			if retryCount == "2" {
				ctx = metainfo.WithValue(ctx, respFlagMsgKey, "0")
			} else {
				ctx = metainfo.WithValue(ctx, respFlagMsgKey, retryMsg)
			}
		}
		err = next(ctx, args, result)
		return
	}
}

var errRetry = &retry.ShouldResultRetry{ErrorRetryWithCtx: func(ctx context.Context, err error, ri rpcinfo.RPCInfo) bool {
	var te *remote.TransError
	ok := errors.As(err, &te)
	if ok && te.TypeID() == retryTransErrCode {
		return true
	}
	return false
}}
