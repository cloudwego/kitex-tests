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
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/server"
)

var _ stability.STService = &STServiceHandler{}

// ServerInitParam .
type ServerInitParam struct {
	Network  string
	Address  string
	ConnMode ConnectionMode
	Limiter  *limit.Option
}

// RunServer .
func RunServer(param *ServerInitParam, handler stability.STService, opts ...server.Option) server.Server {
	var addr net.Addr
	var err error
	switch v := param.Network; v {
	case "unix":
		addr, err = net.ResolveUnixAddr(v, param.Address)
	case "tcp":
		addr, err = net.ResolveTCPAddr(v, param.Address)
	default:
		panic(fmt.Errorf("unsupported network: %s", v))
	}
	if err != nil {
		panic(err)
	}

	opts = append(opts, server.WithServiceAddr(addr))
	opts = append(opts, server.WithLimit(&limit.Option{
		MaxConnections: 30000, MaxQPS: 300000, UpdateControl: func(u limit.Updater) {},
	}))
	if param.ConnMode == ConnectionMultiplexed {
		opts = append(opts, server.WithMuxTransport())
	}
	if param.Limiter != nil {
		opts = append(opts, server.WithLimit(param.Limiter))
	}
	if handler == nil {
		handler = new(STServiceHandler)
	}
	svr := stservice.NewServer(handler, opts...)

	go func() {
		if err = svr.Run(); err != nil {
			panic(err)
		}
	}()
	return svr
}

// STServiceHandler .
type STServiceHandler struct{}

// TestSTReq .
func (*STServiceHandler) TestSTReq(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	resp := &stability.STResponse{
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

// TestObjReq .
func (*STServiceHandler) TestObjReq(ctx context.Context, req *instparam.ObjReq) (r *instparam.ObjResp, err error) {
	resp := &instparam.ObjResp{
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
