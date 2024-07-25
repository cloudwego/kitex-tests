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
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/transport"
)

func TestMain(m *testing.M) {
	svc := runServer()
	gsvc := runGenericServer()
	time.Sleep(100 * time.Millisecond)
	m.Run()
	svc.Stop()
	gsvc.Stop()
}

func TestClient(t *testing.T) {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	test.Assert(t, err == nil)
	g, err := generic.JSONThriftGeneric(p)
	test.Assert(t, err == nil)

	cli, err := genericclient.NewClient("a.b.c", g, client.WithHostPorts(address))
	test.Assert(t, err == nil)

	req := map[string]interface{}{
		"Msg": "hello",
		"I8":  int8(1),
		"I16": int16(1),
		"I32": int32(1),
		"I64": int64(1),
		"Map": map[string]interface{}{
			"hello": "world",
		},
		"Set":       []interface{}{"hello", "world"},
		"List":      []interface{}{"hello", "world"},
		"ErrorCode": int32(1),
		"Info": map[string]interface{}{
			"Map": map[string]interface{}{
				"hello": "world",
			},
			"ID": int64(232324),
		},
	}
	reqStr, _ := json.Marshal(req)
	_, err = cli.GenericCall(context.Background(), "Echo", string(reqStr))
	test.Assert(t, err == nil)
}

func TestGeneric(t *testing.T) {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	test.Assert(t, err == nil)
	g, err := generic.JSONThriftGeneric(p)
	test.Assert(t, err == nil)

	cli, err := genericclient.NewClient("a.b.c", g, client.WithHostPorts(address))
	test.Assert(t, err == nil)

	req := map[string]interface{}{
		"Msg": "hello",
		"I8":  int8(1),
		"I16": int16(1),
		"I32": int32(1),
		"I64": int64(1),
		"Map": map[string]interface{}{
			"hello": "world",
		},
		"Set":       []interface{}{"hello", "world"},
		"List":      []interface{}{"hello", "world"},
		"ErrorCode": int32(1),
		"Info": map[string]interface{}{
			"Map": map[string]interface{}{
				"hello": "world",
			},
			"ID": int64(232324),
		},
	}
	reqStr, _ := json.Marshal(req)
	num := 10
	for i := 0; i < num; i++ {
		_, err = cli.GenericCall(context.Background(), "EchoOneway", string(reqStr))
		test.Assert(t, err == nil)
	}
	// wait for request received
	time.Sleep(200 * time.Millisecond)
	test.Assert(t, atomic.LoadInt32(&checkNum) == int32(num))
}

func TestBizErr(t *testing.T) {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	test.Assert(t, err == nil)
	g, err := generic.JSONThriftGeneric(p)
	test.Assert(t, err == nil)

	cli, err := genericclient.NewClient("a.b.c", g, client.WithHostPorts(genericAddress), client.WithMetaHandler(transmeta.ClientTTHeaderHandler), client.WithTransportProtocol(transport.TTHeader))
	test.Assert(t, err == nil)
	_, err = cli.GenericCall(context.Background(), "Echo", nil)
	bizerr, ok := kerrors.FromBizStatusError(err)
	test.Assert(t, ok)
	test.Assert(t, bizerr.BizStatusCode() == 404)
	test.Assert(t, bizerr.BizMessage() == "not found")
}
