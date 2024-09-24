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
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/retry"
)

var (
	defaultRPCTimeout = 40 * time.Millisecond
	defaultSleepTime  = 60 * time.Millisecond // must > defaultRPCTimeout
)

const (
	retryMsg            = "retry"
	counterNamespaceKey = "counterNs"
	sleepTimeMsKey      = "TIME_SLEEP_TIME_MS"
	skipCounterSleepKey = "TIME_SKIP_COUNTER_SLEEP"
	respFlagMsgKey      = "CUSTOMIZED_RESP_MSG"
)

var (
	testObjReqCount    int32
	testExceptionCount int32
)

type mockType = string

const (
	mockTypeMockTimeout         mockType = "mockTypeMockTimeout"
	mockTypeNonRetryReturnError mockType = "mockTypeNonRetryReturnError"
	mockTypeCustomizedResp      mockType = "mockTypeCustomizedResp"
	mockTypeBizStatus           mockType = "mockTypeBizStatus"
	mockTypeReturnTransErr      mockType = "mockTypeReturnTransErr"
	mockType10PctSleep          mockType = "mockType10PctSleep"

	retryTransErrCode = 1000
)

type reqCounters struct {
	STReq        int32
	CircuitBreak int32

	// TODO: need to refactor the code if tests can't run with t.Parallel()
	// ObjReq           int32
	// Exception        int32
}

// STServiceHandler .
type STServiceHandler struct{}

var _ stability.STService = &STServiceHandler{}

var counters sync.Map

func resetCounters(ns string) {
	counters.Store(ns, &reqCounters{})
}

func getCounter(ns string) *reqCounters {
	v, ok := counters.Load(ns)
	if !ok {
		return nil
	}
	return v.(*reqCounters)
}

type testingT interface {
	Name() string
}

func getTestCounters(t testingT) *reqCounters {
	v, ok := counters.Load(t.Name())
	if !ok {
		return nil
	}
	return v.(*reqCounters)
}

func counterNamespace(t testingT) context.Context {
	ns := t.Name()
	resetCounters(ns)
	return metainfo.WithValue(context.Background(), counterNamespaceKey, ns)
}

func getCounterNamespace(ctx context.Context) string {
	s, _ := metainfo.GetPersistentValue(ctx, counterNamespaceKey)
	if s == "" {
		s, _ = metainfo.GetValue(ctx, counterNamespaceKey)
	}
	return s
}

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
	} else if req.FlagMsg == mockType10PctSleep {
		v := getCounter(getCounterNamespace(ctx))
		if (atomic.AddInt32(&v.STReq, 1)-1)%10 == 0 {
			time.Sleep(defaultSleepTime)
		}
	}

	if sleepTime := getSleepTime(ctx); sleepTime > 0 {
		time.Sleep(sleepTime)
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
			time.Sleep(defaultSleepTime)
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
			time.Sleep(defaultSleepTime)
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
	circuitBreak50PCT      CircuitBreak = "0"
	circuitBreakRetrySleep CircuitBreak = "1"
)

// CircuitBreakTest .
func (h *STServiceHandler) CircuitBreakTest(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	if req.FlagMsg == circuitBreakRetrySleep {
		// use ttheader
		// first request(non retry request) will sleep to mock timeout, then retry request return directly
		if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
			sleepTime := defaultSleepTime
			if v := getSleepTime(ctx); v > 0 {
				sleepTime = v
			}
			time.Sleep(sleepTime)
		}

	} else if req.FlagMsg == circuitBreak50PCT {
		// force 50% of the responses to cost defaultSleepTime
		v := getCounter(getCounterNamespace(ctx))
		if (atomic.AddInt32(&v.CircuitBreak, 1)-1)%2 == 0 {
			time.Sleep(defaultSleepTime)
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

func withSleepTime(ctx context.Context, d time.Duration) context.Context {
	ms := int(d / time.Millisecond)
	return metainfo.WithValue(ctx, sleepTimeMsKey, strconv.Itoa(ms))
}

func getSleepTime(ctx context.Context) time.Duration {
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
