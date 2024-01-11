// Copyright 2023 CloudWeGo Authors
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

package thrift_streaming

import (
	"context"
	"errors"

	cross_echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_cross/echo"
	"github.com/cloudwego/kitex/pkg/klog"
)

type CrossEchoServiceImpl struct{}

func (e *CrossEchoServiceImpl) EchoPingPong(ctx context.Context, req1 *cross_echo.EchoRequest, req2 *cross_echo.EchoRequest) (r *cross_echo.EchoResponse, err error) {
	klog.Infof("CrossEchoServiceImpl.EchoPingPong: req1 = %v, req2 = %v", req1, req2)
	if err = GetError(ctx); err != nil {
		klog.Infof("EchoPingPong: context err = %v", err)
		return
	}
	r = &cross_echo.EchoResponse{
		Message: req1.Message + "-" + req2.Message,
	}
	klog.Infof("EchoPingPong: resp = %v", r)
	return r, nil
}

func (e *CrossEchoServiceImpl) EchoOneway(ctx context.Context, req1 *cross_echo.EchoRequest) (err error) {
	klog.Infof("CrossEchoServiceImpl.EchoOneway: req1 = %v", req1)
	expected := GetValue(ctx, KeyMessage, "")
	if req1.Message != expected {
		klog.Infof("EchoOneway: invalid message, req1 = %v, expected = %v", req1, expected)
		return errors.New("invalid message")
	}
	if err = GetError(ctx); err != nil {
		klog.Infof("EchoOneway: context err = %v", err)
	}
	return
}

func (e *CrossEchoServiceImpl) Ping(ctx context.Context) (err error) {
	err = GetError(ctx)
	klog.Infof("CrossEchoServiceImpl.Ping: context err = %v", err)
	return
}
