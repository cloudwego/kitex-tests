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

package error_handler

import (
	"context"
	"errors"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/status"
)

var (
	normalErr     = errors.New("mock handler normal err")
	kitexTransErr = remote.NewTransErrorWithMsg(1900, "mock handler TransError")
	grpcStatus    = status.Errorf(1900, "mock handler StatusError")
	bizErr        = kerrors.NewBizStatusErrorWithExtra(502, "bad gateway", map[string]string{"version": "v1.0.0"})
	panicStr      = "panic"
	timeout       = "timeout"
)

// STServiceHandler .
type STServiceHandler struct{}

// TestSTReq .
func (*STServiceHandler) TestSTReq(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	r = &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	switch req.Name {
	case normalErr.Error():
		return nil, normalErr
	case kitexTransErr.Error():
		return nil, kitexTransErr
	case grpcStatus.Error():
		return nil, grpcStatus
	case bizErr.Error():
		return nil, bizErr
	case panicStr:
		panic("mock handler panic")
	case timeout:
		if d, e := time.ParseDuration(*req.MockCost); e != nil {
			time.Sleep(time.Second)
		} else {
			time.Sleep(d)
		}
	}
	return r, nil
}

// TestObjReq .
func (*STServiceHandler) TestObjReq(ctx context.Context, req *instparam.ObjReq) (r *instparam.ObjResp, err error) {
	// not called, do nothing for this test
	return
}

// TestException .
func (*STServiceHandler) TestException(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	// not called, do nothing for this test
	return
}

func (*STServiceHandler) VisitOneway(ctx context.Context, req *stability.STRequest) (err error) {
	// not called, do nothing for this test
	return
}

func (h *STServiceHandler) CircuitBreakTest(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	// not called, do nothing for this test
	return
}
