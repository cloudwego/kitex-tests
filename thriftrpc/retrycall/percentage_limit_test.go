// Copyright 2023 CloudWeGo Authors
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
	"sync/atomic"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/retry"
	"github.com/cloudwego/kitex/transport"
)

func TestPercentageLimit(t *testing.T) {
	t.Parallel()

	backupPolicy := &retry.BackupPolicy{
		RetryDelayMS: 10,
		StopPolicy: retry.StopPolicy{
			MaxRetryTimes:    1,
			DisableChainStop: false,
			CBPolicy: retry.CBPolicy{
				ErrorRate: 0.1,
			},
		},
	}

	rc := retry.NewRetryContainerWithPercentageLimit()
	rc.NotifyPolicyChange("circuitBreakTest", retry.BuildBackupRequest(backupPolicy))

	totalCnt, retryCnt := int32(0), int32(0)
	cli := getKitexClient(
		transport.TTHeader,
		client.WithRetryContainer(rc),
		client.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				atomic.AddInt32(&totalCnt, 1)
				if retry.IsLocalRetryRequest(ctx) {
					atomic.AddInt32(&retryCnt, 1)
				}
				return next(ctx, req, resp)
			}
		}),
	)
	ctx, stReq := thriftrpc.CreateSTRequest(counterNamespace(t))
	stReq.FlagMsg = circuitBreakRetrySleep
	ctx = withSleepTime(ctx, defaultSleepTime)

	for i := 0; i < 5; i++ {
		// CircuitBreakTest will 50% sleep defaultSleepTime
		_, err := cli.CircuitBreakTest(ctx, stReq, callopt.WithRPCTimeout(defaultRPCTimeout))
		// first 10 requests will always success
		test.Assert(t, err == nil, err)
	}
	t.Logf("total = %v, retry = %v", atomic.LoadInt32(&totalCnt), atomic.LoadInt32(&retryCnt))
	test.Assert(t, 10 == atomic.LoadInt32(&totalCnt), "should have sent 10 requests (including retry)")
	test.Assert(t, 5 == atomic.LoadInt32(&retryCnt), "should have sent 5 retry requests")

	for i := 0; i < 40; i++ {
		_, _ = cli.CircuitBreakTest(ctx, stReq, callopt.WithRPCTimeout(time.Second))
	}
	t.Logf("total = %v, retry = %v", atomic.LoadInt32(&totalCnt), atomic.LoadInt32(&retryCnt))
	test.Assert(t, 50 == atomic.LoadInt32(&totalCnt), "should have sent 10 requests (including retry)")
	test.Assert(t, 5 == atomic.LoadInt32(&retryCnt), "shouldn't have sent more retry requests (<10%)")

	// 46
	_, _ = cli.CircuitBreakTest(ctx, stReq, callopt.WithRPCTimeout(time.Second))
	t.Logf("total = %v, retry = %v", atomic.LoadInt32(&totalCnt), atomic.LoadInt32(&retryCnt))
	test.Assert(t, 52 == atomic.LoadInt32(&totalCnt), "should have sent 52 requests (including retry)")
	test.Assert(t, 6 == atomic.LoadInt32(&retryCnt), "should have sent 6 retry requests")

	// 47
	_, _ = cli.CircuitBreakTest(ctx, stReq, callopt.WithRPCTimeout(time.Second))
	t.Logf("total = %v, retry = %v", atomic.LoadInt32(&totalCnt), atomic.LoadInt32(&retryCnt))
	test.Assert(t, 53 == atomic.LoadInt32(&totalCnt), "should have sent 53 requests (including retry)")
	test.Assert(t, 6 == atomic.LoadInt32(&retryCnt), "should have sent 6 retry requests")
}

func TestPercentageLimitAfterUpdate(t *testing.T) {
	t.Parallel()

	backupPolicy := &retry.BackupPolicy{
		RetryDelayMS: 10,
		StopPolicy: retry.StopPolicy{
			MaxRetryTimes:    1,
			DisableChainStop: false,
			CBPolicy: retry.CBPolicy{
				ErrorRate: 0.1,
			},
		},
	}

	rc := retry.NewRetryContainerWithPercentageLimit()
	rc.NotifyPolicyChange("circuitBreakTest", retry.BuildBackupRequest(backupPolicy))

	totalCnt, retryCnt := int32(0), int32(0)
	cli := getKitexClient(
		transport.TTHeader,
		client.WithRetryContainer(rc),
		client.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				atomic.AddInt32(&totalCnt, 1)
				if retry.IsLocalRetryRequest(ctx) {
					atomic.AddInt32(&retryCnt, 1)
				}
				return next(ctx, req, resp)
			}
		}),
	)
	ctx, stReq := thriftrpc.CreateSTRequest(counterNamespace(t))
	stReq.FlagMsg = circuitBreakRetrySleep
	ctx = metainfo.WithPersistentValue(ctx, sleepTimeMsKey, "20") //ms

	for i := 0; i < 46; i++ {
		_, _ = cli.CircuitBreakTest(ctx, stReq, callopt.WithRPCTimeout(time.Second))
	}
	t.Logf("total = %v, retry = %v", atomic.LoadInt32(&totalCnt), atomic.LoadInt32(&retryCnt))
	test.Assert(t, 52 == atomic.LoadInt32(&totalCnt), "should have sent 52 requests (including retry)")
	test.Assert(t, 6 == atomic.LoadInt32(&retryCnt), "should have sent 6 retry requests")

	backupPolicyNew := *backupPolicy
	backupPolicyNew.StopPolicy.CBPolicy.ErrorRate = 0.3
	rc.NotifyPolicyChange("circuitBreakTest", retry.BuildBackupRequest(&backupPolicyNew))

	_, _ = cli.CircuitBreakTest(ctx, stReq, callopt.WithRPCTimeout(time.Second))
	t.Logf("total = %v, retry = %v", atomic.LoadInt32(&totalCnt), atomic.LoadInt32(&retryCnt))
	test.Assert(t, 54 == atomic.LoadInt32(&totalCnt), "should have sent 54 requests (including retry)")
	test.Assert(t, 7 == atomic.LoadInt32(&retryCnt), "should have sent 7 retry requests")
}
