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

package pbrpc

import (
	"context"
	"fmt"
	"net"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/stability/stservice"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/server"
)

// ServerInitParam .
type ServerInitParam struct {
	Network  string
	Address  string
	ConnMode ConnectionMode
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
	return resp, nil
}

// TestObjReq .
func (*STServiceHandler) TestObjReq(ctx context.Context, req *instparam.ObjReq) (r *instparam.ObjResp, err error) {
	resp := &instparam.ObjResp{
		Msg:     req.Msg,
		MsgMap:  req.MsgMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}
