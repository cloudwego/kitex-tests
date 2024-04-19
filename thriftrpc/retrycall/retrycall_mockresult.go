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
)

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

var _ stability.STService = &STServiceMockResultHandler{}

// STServiceMockResultHandler .
type STServiceMockResultHandler struct{}

// TestSTReq mock return resp with 'retry' msg
func (h *STServiceMockResultHandler) TestSTReq(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	resp := &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	// use ttheader
	if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
		resp.FlagMsg = retryMsg
	}
	return resp, nil
}

// TestObjReq mock return error with 'retry' info
func (h *STServiceMockResultHandler) TestObjReq(ctx context.Context, req *instparam.ObjReq) (r *instparam.ObjResp, err error) {
	// use ttheader
	if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
		return nil, remote.NewTransErrorWithMsg(1000, retryMsg)
	}
	resp := &instparam.ObjResp{
		Msg:     req.Msg,
		MsgSet:  req.MsgSet,
		MsgMap:  req.MsgMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}

// TestException mock return Exception to do retry
func (h *STServiceMockResultHandler) TestException(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	// use ttheader
	if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
		err = &stability.STException{Message: retryMsg}
		return nil, err
	}
	r = &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	return r, nil
}

// VisitOneway .
func (*STServiceMockResultHandler) VisitOneway(ctx context.Context, req *stability.STRequest) (err error) {
	// no nothing
	return nil
}

// CircuitBreakTest mock sleep which lead request do backup request
func (h *STServiceMockResultHandler) CircuitBreakTest(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	// use ttheader
	if _, exist := metainfo.GetPersistentValue(ctx, retry.TransitKey); !exist {
		sleepTime := 200 * time.Millisecond
		if v := getSleepTimeMS(ctx); v > 0 {
			sleepTime = v
		}
		time.Sleep(sleepTime)
	}
	resp := &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}
