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
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/retry"
)

const (
	retryMsg            = "retry"
	sleepTimeMsKey      = "TIME_SLEEP_TIME_MS"
	skipCounterSleepKey = "TIME_SKIP_COUNTER_SLEEP"
	respFlagMsgKey      = "CUSTOMIZED_RESP_MSG"
)

var _ stability.STService = &STServiceHandler{}

var (
	testSTReqCount        int32
	testObjReqCount       int32
	testExceptionCount    int32
	circuitBreakTestCount int32
)

type mockType = string

const (
	mockTypeMockTimeout         mockType = "0"
	mockTypeNonRetryReturnError mockType = "1"
	mockTypeCustomizedResp      mockType = "2"
	mockTypeSleepWithMetainfo   mockType = "3"
	mockTypeBizStatus           mockType = "4"
	mockTypeReturnTransErr      mockType = "5"

	retryTransErrCode = 1000
)

// STServiceHandler .
type STServiceHandler struct{}

// TestSTReq .
func (h *STServiceHandler) TestSTReq(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	resp := &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	if req.FlagMsg == mockTypeCustomizedResp {
		// should use ttheader

		if respFM := getRespFlagMsg(ctx); respFM != "" {
			// test case1: set FlagMsg with the msg transmit through metainfo
			resp.FlagMsg = respFM
		} else {
			// test case2: first request(non retry request) will return the resp with retry message , then retry request return resp directly
			if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
				resp.FlagMsg = retryMsg
			} else {
				resp.FlagMsg = "success"
			}
		}
		if sleepTime := getSleepTimeMS(ctx); sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	} else if req.FlagMsg == mockTypeSleepWithMetainfo {
		if sleepTime := getSleepTimeMS(ctx); sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	} else if req.FlagMsg == mockTypeReturnTransErr {
		// should use ttheader

		if respFM := getRespFlagMsg(ctx); respFM == retryMsg {
			// test case1: set FlagMsg with the msg transmit through metainfo
			err = remote.NewTransErrorWithMsg(retryTransErrCode, "mock error")
		} else {
			resp.FlagMsg = respFM
			// test case2: first request(non retry request) will return the resp with retry message , then retry request return resp directly
			if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
				err = remote.NewTransErrorWithMsg(retryTransErrCode, "mock error")
			}
		}
		if sleepTime := getSleepTimeMS(ctx); sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	} else {
		if v := atomic.AddInt32(&testSTReqCount, 1); v%10 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return resp, err
}

// TestObjReq .
func (h *STServiceHandler) TestObjReq(ctx context.Context, req *instparam.ObjReq) (r *instparam.ObjResp, err error) {
	resp := &instparam.ObjResp{
		Msg:     req.Msg,
		MsgSet:  req.MsgSet,
		MsgMap:  req.MsgMap,
		FlagMsg: req.FlagMsg,
	}
	if req.FlagMsg == mockTypeNonRetryReturnError {
		// use ttheader
		// first request(non retry request) will return error, then retry request return success
		if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
			return nil, remote.NewTransErrorWithMsg(1000, retryMsg)
		}
	} else {
		if atomic.AddInt32(&testObjReqCount, 1)%5 == 0 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	return resp, nil
}

// TestException .
func (h *STServiceHandler) TestException(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	if req.FlagMsg == mockTypeNonRetryReturnError {
		// use ttheader
		// first request(non retry request) will return error, then retry request return success
		if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
			err = &stability.STException{Message: retryMsg}
			return nil, err
		} else {
			r = &stability.STResponse{
				Str:     req.Str,
				Mp:      req.StringMap,
				FlagMsg: req.FlagMsg,
			}
			return r, nil
		}
	} else {
		err = &stability.STException{Message: "mock exception"}
		if atomic.AddInt32(&testExceptionCount, 1)%30 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return nil, err
}

// VisitOneway .
func (*STServiceHandler) VisitOneway(ctx context.Context, req *stability.STRequest) (err error) {
	// no nothing
	return nil
}

type CircuitBreak = string

const (
	CircuitBreak50PCT      CircuitBreak = "0"
	CircuitBreakRetrySleep CircuitBreak = "1"
)

// CircuitBreakTest .
func (h *STServiceHandler) CircuitBreakTest(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	if req.FlagMsg == CircuitBreakRetrySleep {
		// use ttheader
		// first request(non retry request) will sleep to mock timeout, then retry request return directly
		if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
			sleepTime := 200 * time.Millisecond
			if v := getSleepTimeMS(ctx); v > 0 {
				sleepTime = v
			}
			time.Sleep(sleepTime)
		}

	} else if req.FlagMsg == CircuitBreak50PCT {
		// force 50% of the responses to cost over 200ms
		if atomic.AddInt32(&circuitBreakTestCount, 1)%2 == 0 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	resp := &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}

func setSkipCounterSleep(ctx context.Context) context.Context {
	return metainfo.WithPersistentValue(ctx, skipCounterSleepKey, "1")
}

func skipCounterSleep(ctx context.Context) bool {
	if _, exist := metainfo.GetValue(ctx, skipCounterSleepKey); exist {
		return true
	}
	if _, exist := metainfo.GetPersistentValue(ctx, skipCounterSleepKey); exist {
		return true
	}
	return false
}

func getSleepTimeMS(ctx context.Context) time.Duration {
	if value, exist := metainfo.GetPersistentValue(ctx, sleepTimeMsKey); exist {
		if sleepTimeMS, err := strconv.Atoi(value); err == nil && sleepTimeMS > 0 {
			return time.Duration(sleepTimeMS) * time.Millisecond
		}
	}
	if value, exist := metainfo.GetValue(ctx, sleepTimeMsKey); exist {
		if sleepTimeMS, err := strconv.Atoi(value); err == nil && sleepTimeMS > 0 {
			return time.Duration(sleepTimeMS) * time.Millisecond
		}
	}
	return 0
}

func getRespFlagMsg(ctx context.Context) string {
	if value, exist := metainfo.GetPersistentValue(ctx, respFlagMsgKey); exist {
		return value
	}
	if value, exist := metainfo.GetValue(ctx, respFlagMsgKey); exist {
		return value
	}
	return ""
}
