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
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"

	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_slim/echo"
)

type SlimEchoServiceImpl struct{}

func (e *SlimEchoServiceImpl) EchoBidirectional(stream echo.EchoService_EchoBidirectionalServer) (err error) {
	klog.Infof("EchoBidirectional: start")
	count := GetInt(stream.Context(), KeyCount, 0)
	var req *echo.EchoRequest
	var resp *echo.EchoResponse
	for i := 0; i < count; i++ {
		req, err = stream.Recv()
		if err != nil {
			klog.Infof("EchoBidirectional: recv error = %v", err)
			return
		}
		klog.Infof("EchoBidirectional: recv req = %v", req)
		resp = &echo.EchoResponse{
			Message: req.Message,
		}
		if err = stream.Send(resp); err != nil {
			klog.Infof("EchoBidirectional: send error = %v", err)
			return
		}
		klog.Infof("EchoBidirectional: send resp = %v", resp)
	}
	return
}

func (e *SlimEchoServiceImpl) EchoClient(stream echo.EchoService_EchoClientServer) (err error) {
	klog.Infof("EchoClient: start")
	count := GetInt(stream.Context(), KeyCount, 0)
	var req *echo.EchoRequest
	for i := 0; i < count; i++ {
		req, err = stream.Recv()
		if err != nil {
			klog.Infof("EchoClient: recv error = %v", err)
			return
		}
		klog.Infof("EchoClient: recv req = %v", req)
	}
	resp := &echo.EchoResponse{
		Message: strconv.Itoa(count),
	}
	if err = stream.SendAndClose(resp); err != nil {
		klog.Infof("EchoClient: send&close error = %v", err)
		return
	}
	klog.Infof("EchoClient: send&close resp = %v", resp)
	return
}

func (e *SlimEchoServiceImpl) EchoServer(req *echo.EchoRequest, stream echo.EchoService_EchoServerServer) (err error) {
	klog.Infof("EchoServer: recv req = %v", req)
	count := GetInt(stream.Context(), KeyCount, 0)
	for i := 0; i < count; i++ {
		resp := &echo.EchoResponse{
			Message: req.Message + "-" + strconv.Itoa(i),
		}
		if err = stream.Send(resp); err != nil {
			klog.Infof("EchoServer: send error = %v", err)
			return
		}
		klog.Infof("EchoServer: send resp = %v", resp)
	}
	return nil
}

func (e *SlimEchoServiceImpl) EchoUnary(ctx context.Context, req1 *echo.EchoRequest) (r *echo.EchoResponse, err error) {
	klog.Infof("EchoUnary: req1 = %v", req1)
	if err = GetError(ctx); err != nil {
		klog.Infof("EchoPingPong: context err = %v", err)
		return
	}
	r = &echo.EchoResponse{
		Message: req1.Message,
	}
	klog.Infof("EchoUnary: resp = %v", r)
	return r, nil
}

func (e *SlimEchoServiceImpl) EchoPingPong(ctx context.Context, req1, req2 *echo.EchoRequest) (r *echo.EchoResponse, err error) {
	klog.Infof("EchoPingPong: req1 = %v, req2 = %v", req1, req2)
	if err = GetError(ctx); err != nil {
		klog.Infof("EchoPingPong: context err = %v", err)
		return
	}
	r = &echo.EchoResponse{
		Message: req1.Message + "-" + req2.Message,
	}
	klog.Infof("EchoPingPong: resp = %v", r)
	return r, nil
}

func (e *SlimEchoServiceImpl) EchoOneway(ctx context.Context, req1 *echo.EchoRequest) (err error) {
	klog.Infof("EchoOneway: req1 = %v", req1)
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

func (e *SlimEchoServiceImpl) Ping(ctx context.Context) (err error) {
	err = GetError(ctx)
	klog.Infof("Ping: context err = %v", err)
	return
}
