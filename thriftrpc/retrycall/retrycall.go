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
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
)

var _ stability.STService = &STServiceHandler{}

var (
	testSTReqCount        int32
	testObjReqCount       int32
	testExceptionCount    int32
	circuitBreakTestCount int32
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
	if !skipCounterSleep(ctx) { // better not rely on this since it's a global variable
		if v := atomic.AddInt32(&testSTReqCount, 1); v%10 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	if sleepTime := getSleepTimeMS(ctx); sleepTime > 0 {
		time.Sleep(sleepTime)
	}
	return resp, nil
}

// TestObjReq .
func (h *STServiceHandler) TestObjReq(ctx context.Context, req *instparam.ObjReq) (r *instparam.ObjResp, err error) {
	resp := &instparam.ObjResp{
		Msg:     req.Msg,
		MsgSet:  req.MsgSet,
		MsgMap:  req.MsgMap,
		FlagMsg: req.FlagMsg,
	}

	if atomic.AddInt32(&testObjReqCount, 1)%5 == 0 {
		time.Sleep(200 * time.Millisecond)
	}
	return resp, nil
}

// TestException .
func (h *STServiceHandler) TestException(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	err = &stability.STException{Message: "mock exception"}

	if atomic.AddInt32(&testExceptionCount, 1)%30 == 0 {
		time.Sleep(50 * time.Millisecond)
	}
	return nil, err
}

// VisitOneway .
func (*STServiceHandler) VisitOneway(ctx context.Context, req *stability.STRequest) (err error) {
	// no nothing
	return nil
}

// CircuitBreakTest .
func (h *STServiceHandler) CircuitBreakTest(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	// force 50% of the responses to cost over 200ms
	if atomic.AddInt32(&circuitBreakTestCount, 1)%2 == 0 {
		time.Sleep(200 * time.Millisecond)
	}
	resp := &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}
