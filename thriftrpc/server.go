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

package thriftrpc

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/trans"
	"github.com/cloudwego/kitex/pkg/remote/trans/netpoll"
	"github.com/cloudwego/kitex/pkg/remote/trans/shmipc"
	"github.com/cloudwego/kitex/server"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	instparam_apache_codec "github.com/cloudwego/kitex-tests/kitex_gen_apache_codec/thrift/instparam"
	stability_apache_codec "github.com/cloudwego/kitex-tests/kitex_gen_apache_codec/thrift/stability"
	stservice_apache_codec "github.com/cloudwego/kitex-tests/kitex_gen_apache_codec/thrift/stability/stservice"
	instparam_slim "github.com/cloudwego/kitex-tests/kitex_gen_slim/thrift/instparam"
	stability_slim "github.com/cloudwego/kitex-tests/kitex_gen_slim/thrift/stability"
	stservice_slim "github.com/cloudwego/kitex-tests/kitex_gen_slim/thrift/stability/stservice"
)

var (
	_ stability.STService      = &STServiceHandler{}
	_ stability_slim.STService = &STServiceSlimHandler{}
)

// ServerInitParam .
type ServerInitParam struct {
	Listener net.Listener
	ConnMode ConnectionMode
}

// RunServer .
func RunServer(param *ServerInitParam, handler stability.STService, opts ...server.Option) server.Server {
	opts = generateServerOptionsFromParam(param, opts...)
	if handler == nil {
		handler = new(STServiceHandler)
	}
	svr := stservice.NewServer(handler, opts...)

	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	return svr
}

// RunSlimServer .
func RunSlimServer(param *ServerInitParam, handler stability_slim.STService, opts ...server.Option) server.Server {
	opts = generateServerOptionsFromParam(param, opts...)
	if handler == nil {
		handler = new(STServiceSlimHandler)
	}
	svr := stservice_slim.NewServer(handler, opts...)

	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	return svr
}

// RunApacheCodecServer .
func RunApacheCodecServer(param *ServerInitParam, handler stability_apache_codec.STService, opts ...server.Option) server.Server {
	opts = generateServerOptionsFromParam(param, opts...)
	if handler == nil {
		handler = new(STServiceApacheCodecHandler)
	}
	svr := stservice_apache_codec.NewServer(handler, opts...)

	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	return svr
}

// RunEmptyServer .
func RunEmptyServer(param *ServerInitParam, handler stability.STService, opts ...server.Option) server.Server {
	opts = generateServerOptionsFromParam(param, opts...)
	if handler == nil {
		handler = new(STServiceHandler)
	}
	svr := server.NewServer(opts...)
	svcInfo := stservice.NewServiceInfo()
	delete(svcInfo.Methods, "testSTReq")
	if err := svr.RegisterService(svcInfo, handler); err != nil {
		panic(err)
	}

	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	return svr
}

// generateServerOptionsFromParam process ServerInitParam and add server.Option
func generateServerOptionsFromParam(param *ServerInitParam, opts ...server.Option) []server.Option {
	opts = append(opts, server.WithExitWaitTime(time.Millisecond*10))
	opts = append(opts, server.WithListener(param.Listener))
	opts = append(opts, server.WithLimit(&limit.Option{
		MaxConnections: 30000, MaxQPS: 300000, UpdateControl: func(u limit.Updater) {},
	}))
	if param.ConnMode == ConnectionMultiplexed {
		opts = append(opts, server.WithMuxTransport())
	}

	return opts
}

// STServiceHandler .
type STServiceHandler struct{}

// TestSTReq .
func (*STServiceHandler) TestSTReq(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	resp := stability.NewSTResponse()
	resp.Str = req.Str
	resp.Mp = req.StringMap
	resp.FlagMsg = req.FlagMsg
	resp.DefaultValue = req.DefaultValue
	if req.MockCost != nil {
		if mockSleep, err := time.ParseDuration(*req.MockCost); err != nil {
			return nil, err
		} else {
			time.Sleep(mockSleep)
		}
	}
	return resp, nil
}

// TestObjReq .
func (*STServiceHandler) TestObjReq(ctx context.Context, req *instparam.ObjReq) (r *instparam.ObjResp, err error) {
	resp := instparam.NewObjResp()
	resp.Msg = req.Msg
	resp.MsgSet = req.MsgSet
	resp.MsgMap = req.MsgMap
	resp.FlagMsg = req.FlagMsg
	if req.MockCost != nil {
		if mockSleep, err := time.ParseDuration(*req.MockCost); err != nil {
			return nil, err
		} else {
			time.Sleep(mockSleep)
		}
	}
	return resp, nil
}

// TestException .
func (*STServiceHandler) TestException(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	err = &stability.STException{Message: "mock exception"}
	return nil, err
}

// VisitOneway .
var CheckNum int32

func (*STServiceHandler) VisitOneway(ctx context.Context, req *stability.STRequest) (err error) {
	atomic.AddInt32(&CheckNum, 1)
	return nil
}

var countFlag int32

// CircuitBreakTest .
func (h *STServiceHandler) CircuitBreakTest(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	// force 50% of the responses to cost over 200ms
	if atomic.AddInt32(&countFlag, 1)%2 == 0 {
		time.Sleep(200 * time.Millisecond)
	}
	resp := &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}

type STServiceSlimHandler struct{}

func (S *STServiceSlimHandler) VisitOneway(ctx context.Context, req *stability_slim.STRequest) (err error) {
	// TODO implement me
	panic("implement me")
}

func (S *STServiceSlimHandler) TestSTReq(ctx context.Context, req *stability_slim.STRequest) (r *stability_slim.STResponse, err error) {
	resp := &stability_slim.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	if req.MockCost != nil {
		if mockSleep, err := time.ParseDuration(*req.MockCost); err != nil {
			return nil, err
		} else {
			time.Sleep(mockSleep)
		}
	}
	return resp, nil
}

func (S *STServiceSlimHandler) TestObjReq(ctx context.Context, req *instparam_slim.ObjReq) (r *instparam_slim.ObjResp, err error) {
	resp := &instparam_slim.ObjResp{
		Msg:     req.Msg,
		MsgSet:  req.MsgSet,
		MsgMap:  req.MsgMap,
		FlagMsg: req.FlagMsg,
	}
	if req.MockCost != nil {
		if mockSleep, err := time.ParseDuration(*req.MockCost); err != nil {
			return nil, err
		} else {
			time.Sleep(mockSleep)
		}
	}
	return resp, nil
}

func (S *STServiceSlimHandler) TestException(ctx context.Context, req *stability_slim.STRequest) (r *stability_slim.STResponse, err error) {
	err = &stability_slim.STException{Message: "mock exception"}
	return nil, err
}

func (S *STServiceSlimHandler) CircuitBreakTest(ctx context.Context, req *stability_slim.STRequest) (r *stability_slim.STResponse, err error) {
	// force 50% of the responses to cost over 200ms
	if atomic.AddInt32(&countFlag, 1)%2 == 0 {
		time.Sleep(200 * time.Millisecond)
	}
	resp := &stability_slim.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}

type STServiceApacheCodecHandler struct{}

func (S *STServiceApacheCodecHandler) VisitOneway(ctx context.Context, req *stability_apache_codec.STRequest) (err error) {
	// TODO implement me
	panic("implement me")
}

func (S *STServiceApacheCodecHandler) TestSTReq(ctx context.Context, req *stability_apache_codec.STRequest) (r *stability_apache_codec.STResponse, err error) {
	resp := &stability_apache_codec.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	if req.MockCost != nil {
		if mockSleep, err := time.ParseDuration(*req.MockCost); err != nil {
			return nil, err
		} else {
			time.Sleep(mockSleep)
		}
	}
	return resp, nil
}

func (S *STServiceApacheCodecHandler) TestObjReq(ctx context.Context, req *instparam_apache_codec.ObjReq) (r *instparam_apache_codec.ObjResp, err error) {
	resp := &instparam_apache_codec.ObjResp{
		Msg:     req.Msg,
		MsgSet:  req.MsgSet,
		MsgMap:  req.MsgMap,
		FlagMsg: req.FlagMsg,
	}
	if req.MockCost != nil {
		if mockSleep, err := time.ParseDuration(*req.MockCost); err != nil {
			return nil, err
		} else {
			time.Sleep(mockSleep)
		}
	}
	return resp, nil
}

func (S *STServiceApacheCodecHandler) TestException(ctx context.Context, req *stability_apache_codec.STRequest) (r *stability_apache_codec.STResponse, err error) {
	err = &stability_apache_codec.STException{Message: "mock exception"}
	return nil, err
}

func (S *STServiceApacheCodecHandler) CircuitBreakTest(ctx context.Context, req *stability_apache_codec.STRequest) (r *stability_apache_codec.STResponse, err error) {
	// force 50% of the responses to cost over 200ms
	if atomic.AddInt32(&countFlag, 1)%2 == 0 {
		time.Sleep(200 * time.Millisecond)
	}
	resp := &stability_apache_codec.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}

// RunSHMIPCServer creates and runs a SHMIPC server
func RunSHMIPCServer(param *ServerInitParam, handler stability.STService, opts ...server.Option) server.Server {
	shmipcSockPath := "./kitex_tests_shmipc.sock"
	udsSockPath := "./kitex_tests_uds.sock"

	opts = generateServerOptionsFromParam(param, opts...)
	if handler == nil {
		handler = new(STServiceHandler)
	}

	// Add SHMIPC specific options
	opts = append(opts,
		server.WithServiceAddr(&net.UnixAddr{Net: "unix", Name: udsSockPath}),
		server.WithTransServerFactory(trans.NewMultipleTransServerFactory([]remote.TransServerFactory{
			shmipc.NewTransServerFactory(nil, &net.UnixAddr{Net: "unix", Name: shmipcSockPath}),
			netpoll.NewTransServerFactory()})),
		server.WithTransHandlerFactory(shmipc.NewSvrTransHandlerFactory()),
	)

	svr := stservice.NewServer(handler, opts...)

	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	return svr
}
