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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/klog"
)

func TestMain(m *testing.M) {
	klog.SetLevel(klog.LevelFatal)
	svc := runServer()
	time.Sleep(time.Second)
	m.Run()
	svc.Stop()
}

func TestClient(t *testing.T) {
	p, err := generic.NewThriftFileProvider("../../idl/http.thrift")
	test.Assert(t, err == nil, err)
	g, err := generic.HTTPThriftGeneric(p)
	test.Assert(t, err == nil, err)
	cli := newGenericClient("a.b.c", g, address)

	body := map[string]interface{}{
		"text": "text",
		"some": map[string]interface{}{
			"MyID": "1",
			"text": "text",
		},
		"req_items_map": map[string]interface{}{
			"1": map[string]interface{}{
				"MyID": "1",
				"text": "text",
			},
		},
	}

	data, err := json.Marshal(body)
	test.Assert(t, err == nil, err)
	url := "http://example.com/life/client/1/1?v_int64=1&req_items=item1,item2,item3&cids=1,2,3&vids=1,2,3&api_version=1"
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(data))
	test.Assert(t, err == nil, err)
	req.Header.Set("token", "1")
	customReq, err := generic.FromHTTPRequest(req)
	test.Assert(t, err == nil, err)
	respI, err := cli.GenericCall(context.Background(), "", customReq, callopt.WithRPCTimeout(time.Second))
	test.Assert(t, err == nil, err)
	resp, ok := respI.(*generic.HTTPResponse)
	test.Assert(t, ok)

	test.DeepEqual(t, "1", resp.Header.Get("T"))
	test.DeepEqual(t, map[string]interface{}{"1": map[string]interface{}{"item_id": int64(1), "text": "1"}}, resp.Body["rsp_items"])
	test.DeepEqual(t, nil, resp.Body["v_enum"])
	test.DeepEqual(t, []interface{}{map[string]interface{}{"item_id": int64(1), "text": "1"}}, resp.Body["rsp_item_list"])
	test.DeepEqual(t, int32(1), resp.StatusCode)
	test.DeepEqual(t, "1,2,3", resp.Header.Get("item_count"))
}
