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
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

const (
	transientKV  = "aservice_transient"
	transientKV2  = "aservice_transient2"
	persistentKV = "aservice_persist"
)

func TestMain(m *testing.M) {
	cli := getKitexClient(transport.TTHeader, client.WithHostPorts(":9002"))

	// b
	svrb := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9001",
	}, &stServiceHandler{cli: cli}, server.WithMetaHandler(testMetaHandler{}), server.WithContextBackup(true, true, func(ctx context.Context) bool {
		return ctx.Value(OriginalKey) != OriginalVal
	}))

	 // c
	svrc := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9002",
	}, &stServiceHandler{})

	time.Sleep(time.Second)
	m.Run()
	svrb.Stop()
	svrc.Stop()
}

func TestTransientKV(t *testing.T) {
	// a
	cli := getKitexClient(transport.TTHeader, client.WithHostPorts(":9001"))

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	ctx = metainfo.WithPersistentValue(ctx, persistentKV, persistentKV)
	ctx = metainfo.WithValue(ctx, transientKV, transientKV)

	// a -> b
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
	if h.cli != nil { // a -> b
		// it is service b, both has transient and persist key
		val, ok := metainfo.GetPersistentValue(ctx, persistentKV)
		if !ok || val != persistentKV {
			return nil, fmt.Errorf("persist info[%s] is not right, expect=%s, actual=%s", persistentKV, persistentKV, val)
		}
		val, ok = metainfo.GetValue(ctx, transientKV)
		if !ok || val != transientKV {
			return nil, fmt.Errorf("transit info[%s] is not right, expect=%s, actual=%s", transientKV, transientKV, val)
		}
		// Case 3: transmit async
		sig := make(chan struct{})
		go func() {
			// Case 1: intentionally miss context here, without OriginalKey
			ctx = context.Background()
			// Case 2: set transient kv meanwhile
			ctx = metainfo.WithValue(ctx, transientKV2, transientKV2)
			
			r, err = h.cli.TestSTReq(ctx, req)
			sig <- struct{}{}
		}()
		<- sig
		return
	} else { // b -> c
		// it is service c, both has  persist key from a and transient key from b
		val, ok := metainfo.GetPersistentValue(ctx, persistentKV)
		if !ok || val != persistentKV {
			return nil, fmt.Errorf("persist info[%s] is not right, expect=%s, actual=%s", persistentKV, persistentKV, val)
		}
		val, ok = metainfo.GetValue(ctx, transientKV)
		if ok {
			return nil, fmt.Errorf("transit info[%s] is not right, expect=%s, actual=%s", transientKV, "", val)
		}
		val, ok = metainfo.GetValue(ctx, transientKV2)
		if !ok || val != transientKV2 {
			return nil, fmt.Errorf("transit info[%s] is not right, expect=%s, actual=%s", transientKV, transientKV, val)
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

var (
	OriginalKey originalKeyType
	OriginalVal originalKeyType
)

type originalKeyType struct{}

type testMetaHandler struct {
	SetMeta bool
}

func (t testMetaHandler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}

func (t testMetaHandler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	tk1, tv1 := "tk1", "tv1"
	// to check if kitex filter the transient key
	ctx = metainfo.SetMetaInfoFromMap(ctx, map[string]string{metainfo.PrefixTransient + tk1: tv1})

	ctx = context.WithValue(ctx, OriginalKey, OriginalVal)
	return ctx, nil
}
