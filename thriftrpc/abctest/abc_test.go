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

package abctest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/transport"
)

const (
	transientKV  = "aservice_transient"
	persistentKV = "aservice_persist"
)

func TestMain(m *testing.M) {
	cli := getKitexClient(transport.TTHeader, client.WithHostPorts(":9002"))
	svrb := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9001",
	}, &STServiceHandler{cli: cli})

	svrc := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9002",
	}, &STServiceHandler{})
	time.Sleep(time.Second)
	m.Run()
	svrb.Stop()
	svrc.Stop()
}

func TestTransientKV(t *testing.T) {
	cli := getKitexClient(transport.TTHeader, client.WithHostPorts(":9001"))

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	ctx = metainfo.WithValue(ctx, transientKV, transientKV)
	ctx = metainfo.WithPersistentValue(ctx, persistentKV, persistentKV)

	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)

	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)

	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)
}

func getKitexClient(p transport.Protocol, opts ...client.Option) stservice.Client {
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		Protocol:          p,
		ConnMode:          thriftrpc.LongConnection,
	}, opts...)
}

// STServiceHandler .
type STServiceHandler struct {
	cli stservice.Client
}

// TestSTReq .
func (h *STServiceHandler) TestSTReq(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	if h.cli != nil {
		// it is service b, both has transient and persist key
		val, ok := metainfo.GetValue(ctx, transientKV)
		if !ok || val != transientKV {
			return nil, fmt.Errorf("transit info[%s] is not right, expect=%s, actual=%s", transientKV, transientKV, val)
		}
		val, ok = metainfo.GetPersistentValue(ctx, persistentKV)
		if !ok || val != persistentKV {
			return nil, fmt.Errorf("transit info[%s] is not right, expect=%s, actual=%s", persistentKV, persistentKV, val)
		}
		return h.cli.TestSTReq(ctx, req)
	} else {
		// it is service c, only has persist key
		val, ok := metainfo.GetPersistentValue(ctx, persistentKV)
		if !ok || val != persistentKV {
			return nil, fmt.Errorf("transit info[%s] is not right, expect=%s, actual=%s", persistentKV, persistentKV, val)
		}
		// it is service b, both has transient and persist key
		val, ok = metainfo.GetValue(ctx, transientKV)
		if ok {
			return nil, fmt.Errorf("transit info[%s] is not right, expect=%s, actual=%s", transientKV, "", val)
		}
	}
	resp := &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
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
	return resp, nil
}

// TestException .
func (h *STServiceHandler) TestException(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	err = &stability.STException{Message: "mock exception"}
	return nil, err
}

func (h *STServiceHandler) VisitOneway(ctx context.Context, req *stability.STRequest) (err error) {
	return nil
}

// CircuitBreakTest .
func (h *STServiceHandler) CircuitBreakTest(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	resp := &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}
