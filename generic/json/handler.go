// Copyright 2024 CloudWeGo Authors
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

package tests

import (
	"context"
	"sync/atomic"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
	"github.com/cloudwego/kitex/pkg/kerrors"
)

// EchoServiceImpl implements the last service interface defined in the IDL.
type EchoServiceImpl struct{}

// Echo implements the EchoServiceImpl interface.
func (s *EchoServiceImpl) Echo(ctx context.Context, req *tenant.EchoRequest) (r *tenant.EchoResponse, err error) {
	if err := assertRequest(req); err != nil {
		return nil, err
	}
	return &tenant.EchoResponse{
		Msg:       "world",
		I8:        1,
		I16:       1,
		I32:       1,
		I64:       1,
		Binary:    []byte("world"),
		Map:       map[string]string{"hello": "world"},
		Set:       []string{"hello", "world"},
		List:      []string{"hello", "world"},
		ErrorCode: tenant.ErrorCode_FAILURE,
		Info: &tenant.Info{
			Map: map[string]string{"hello": "world"},
			ID:  233333,
		},
	}, nil
}

var checkNum int32

func (s *EchoServiceImpl) EchoOneway(ctx context.Context, req *tenant.EchoRequest) error {
	if err := assertRequest(req); err != nil {
		return err
	}
	atomic.AddInt32(&checkNum, 1)
	return nil
}

type GenericServiceImpl struct{}

func (s *GenericServiceImpl) GenericCall(ctx context.Context, method string, request interface{}) (response interface{}, err error) {
	return nil, kerrors.NewBizStatusError(404, "not found")
}

func assertRequest(req *tenant.EchoRequest) error {
	if err := assert(req.GetMsg(), "hello"); err != nil {
		return err
	}
	if err := assert(req.GetI8(), int8(1)); err != nil {
		return err
	}
	if err := assert(req.GetI16(), int16(1)); err != nil {
		return err
	}
	if err := assert(req.GetI32(), int32(1)); err != nil {
		return err
	}
	if err := assert(req.GetI64(), int64(1)); err != nil {
		return err
	}
	if err := assert(req.GetMap(), map[string]string{
		"hello": "world",
	}); err != nil {
		return err
	}
	if err := assert(req.GetSet(), []string{"hello", "world"}); err != nil {
		return err
	}
	if err := assert(req.GetList(), []string{"hello", "world"}); err != nil {
		return err
	}
	if err := assert(req.GetErrorCode(), tenant.ErrorCode_FAILURE); err != nil {
		return err
	}
	if err := assert(req.GetInfo(), &tenant.Info{
		Map: map[string]string{"hello": "world"},
		ID:  232324,
	}); err != nil {
		return err
	}
	return nil
}
