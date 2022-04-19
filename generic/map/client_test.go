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
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
)

func TestMain(m *testing.M) {
	svc := runServer()
	time.Sleep(time.Second)
	m.Run()
	svc.Stop()
}

func TestClient(t *testing.T) {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	test.Assert(t, err == nil)
	g, err := generic.MapThriftGeneric(p)
	test.Assert(t, err == nil)

	cli, err := genericclient.NewClient("a.b.c", g, client.WithHostPorts(address))
	test.Assert(t, err == nil)

	req := map[string]interface{}{
		"Msg":    "hello",
		"I8":     int8(1),
		"I16":    int16(1),
		"I32":    int32(1),
		"I64":    int64(1),
		"Binary": []byte("hello"),
		"Map": map[interface{}]interface{}{
			"hello": "world",
		},
		"Set":       []interface{}{"hello", "world"},
		"List":      []interface{}{"hello", "world"},
		"ErrorCode": int32(1),
		"Info": map[string]interface{}{
			"Map": map[interface{}]interface{}{
				"hello": "world",
			},
			"ID": int64(232324),
		},
	}
	resp, err := cli.GenericCall(context.Background(), "Echo", req)
	test.Assert(t, err == nil)
	respM, ok := resp.(map[string]interface{})
	test.Assert(t, ok)
	test.Assert(t, "world" == respM["Msg"])
	test.Assert(t, int8(1) == respM["I8"])
	test.Assert(t, int16(1) == respM["I16"])
	test.Assert(t, int32(1) == respM["I32"])
	test.Assert(t, int64(1) == respM["I64"])
	test.Assert(t, int32(1) == respM["ErrorCode"])
	test.Assert(t, "world" == respM["Binary"], respM["Binary"])
	test.DeepEqual(t, map[interface{}]interface{}{"hello": "world"}, respM["Map"])
	test.DeepEqual(t, []interface{}{"hello", "world"}, respM["Set"])
	test.DeepEqual(t, []interface{}{"hello", "world"}, respM["List"])
	test.DeepEqual(t, map[string]interface{}{
		"Map": map[interface{}]interface{}{
			"hello": "world",
		},
		"ID": int64(233333),
	}, respM["Info"])
}

func TestGeneric(t *testing.T) {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	test.Assert(t, err == nil)
	g, err := generic.MapThriftGeneric(p)
	test.Assert(t, err == nil)

	cli, err := genericclient.NewClient("a.b.c", g, client.WithHostPorts(address))
	test.Assert(t, err == nil)

	req := map[string]interface{}{
		"Msg":    "hello",
		"I8":     int8(1),
		"I16":    int16(1),
		"I32":    int32(1),
		"I64":    int64(1),
		"Binary": []byte("hello"),
		"Map": map[interface{}]interface{}{
			"hello": "world",
		},
		"Set":       []interface{}{"hello", "world"},
		"List":      []interface{}{"hello", "world"},
		"ErrorCode": int32(1),
		"Info": map[string]interface{}{
			"Map": map[interface{}]interface{}{
				"hello": "world",
			},
			"ID": int64(232324),
		},
	}
	num := 10
	for i := 0; i < num; i++ {
		_, err = cli.GenericCall(context.Background(), "EchoOneway", req)
		test.Assert(t, err == nil)
	}
	// wait for request received
	time.Sleep(200 * time.Millisecond)
	test.Assert(t, atomic.LoadInt32(&checkNum) == int32(num))
}
