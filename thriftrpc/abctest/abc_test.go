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

	"github.com/bytedance/gopkg/cloud/metainfo"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

const (
	transientKV  = "aservice_transient"
	persistentKV = "aservice_persist"
)

var testaddr1, testaddr2 string

func TestMain(m *testing.M) {
	testaddr1 = serverutils.NextListenAddr()
	testaddr2 = serverutils.NextListenAddr()
	cli := getKitexClient(transport.TTHeader, client.WithHostPorts(testaddr2))
	svrb := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: testaddr1,
	}, &stServiceHandler{cli: cli}, server.WithMetaHandler(testMetaHandler{}))

	svrc := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: testaddr2,
	}, &stServiceHandler{})

	serverutils.Wait(testaddr1)
	serverutils.Wait(testaddr2)
	m.Run()
	svrb.Stop()
	svrc.Stop()
}

func TestTransientKV(t *testing.T) {
	cli := getKitexClient(transport.TTHeader, client.WithHostPorts(testaddr1))

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

// stServiceHandler .
type stServiceHandler struct {
	cli stservice.Client
}

// TestSTReq .
func (h *stServiceHandler) TestSTReq(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	if req.DefaultValue != "default" {
		return nil, fmt.Errorf("req DefaultValue is not right, expect=default, actual=%s", req.DefaultValue)
	}
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
func (h *stServiceHandler) TestObjReq(ctx context.Context, req *instparam.ObjReq) (r *instparam.ObjResp, err error) {
	resp := &instparam.ObjResp{
		Msg:     req.Msg,
		MsgSet:  req.MsgSet,
		MsgMap:  req.MsgMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}

// TestException .
func (h *stServiceHandler) TestException(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	err = &stability.STException{Message: "mock exception"}
	return nil, err
}

func (h *stServiceHandler) VisitOneway(ctx context.Context, req *stability.STRequest) (err error) {
	return nil
}

// CircuitBreakTest .
func (h *stServiceHandler) CircuitBreakTest(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	resp := &stability.STResponse{
		Str:     req.Str,
		Mp:      req.StringMap,
		FlagMsg: req.FlagMsg,
	}
	return resp, nil
}

type testMetaHandler struct{}

func (t testMetaHandler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}

func (t testMetaHandler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	tk1, tv1 := "tk1", "tv1"
	// to check if kitex filter the transient key
	ctx = metainfo.SetMetaInfoFromMap(ctx, map[string]string{metainfo.PrefixTransient + tk1: tv1})
	return ctx, nil
}

func (t testMetaHandler) OnConnectStream(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (t testMetaHandler) OnReadStream(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
