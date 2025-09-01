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

package tests

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"

	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/kerrors"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant"
)

var returnedResponse = &tenant.EchoResponse{
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
}

func getReturnedResponseMap() map[string]interface{} {
	m := make(map[string]interface{})
	b, err := json.Marshal(returnedResponse)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

// EchoServiceImpl implements the last service interface defined in the IDL.
type EchoServiceImpl struct {
	tenant.EchoService
}

// Echo implements the EchoServiceImpl interface.
func (s *EchoServiceImpl) Echo(ctx context.Context, req *tenant.EchoRequest) (r *tenant.EchoResponse, err error) {
	if err := assertRequest(req); err != nil {
		return nil, err
	}
	b, err := json.Marshal(returnedResponse)
	if err != nil {
		return nil, err
	}
	var resp tenant.EchoResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
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
	if m := request.(map[string]interface{}); m["Msg"] == "biz_error" {
		return nil, kerrors.NewBizStatusError(404, "not found")
	}
	return getReturnedResponseMap(), nil
}

type GenericServiceImplV2 struct{}

func (s *GenericServiceImplV2) GenericCall(ctx context.Context, service, method string, request interface{}) (response interface{}, err error) {
	return getReturnedResponseMap(), nil
}

func (s *GenericServiceImplV2) ClientStreaming(ctx context.Context, service, method string, stream generic.ClientStreamingServer) (err error) {
	req, err := stream.Recv(ctx)
	if err != nil {
		return err
	}
	if req.(map[string]interface{})["Msg"] != "hello world" {
		return errors.New("request msg not match")
	}
	return stream.SendAndClose(ctx, getReturnedResponseMap())
}

func (s *GenericServiceImplV2) ServerStreaming(ctx context.Context, service, method string, request interface{}, stream generic.ServerStreamingServer) (err error) {
	if req := request.(map[string]interface{}); req["Msg"] != "hello world" {
		return errors.New("request msg not match")
	}
	return stream.Send(ctx, getReturnedResponseMap())
}

func (s *GenericServiceImplV2) BidiStreaming(ctx context.Context, service, method string, stream generic.BidiStreamingServer) (err error) {
	req, err := stream.Recv(ctx)
	if err != nil {
		return err
	}
	if req.(map[string]interface{})["Msg"] != "hello world" {
		return errors.New("request msg not match")
	}
	return stream.Send(ctx, getReturnedResponseMap())
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
	if err := assert(req.GetBinary(), []byte("hello")); err != nil {
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
