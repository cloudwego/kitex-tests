// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crrst

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

var testaddr string

type stServiceHandler struct {
	stability.STService
}

func TestMain(m *testing.M) {
	testaddr = serverutils.NextListenAddr()
	svr := thriftrpc.RunEmptyServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: testaddr,
	}, &stServiceHandler{}, server.WithMetaHandler(transmeta.ServerTTHeaderHandler))

	serverutils.Wait(testaddr)
	m.Run()
	svr.Stop()
}

func TestCrrstFlag(t *testing.T) {
	cli := thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		Protocol:          transport.TTHeader,
		ConnMode:          thriftrpc.LongConnection,
	}, client.WithHostPorts(testaddr), client.WithMetaHandler(transmeta.ClientTTHeaderHandler), client.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			err = next(ctx, req, resp)
			ri := rpcinfo.GetRPCInfo(ctx)
			val, ok := ri.To().Tag(rpcinfo.ConnResetTag)
			if !ok || val != "1" {
				panic("conn reset flag is not set")
			}
			return err
		}
	}))

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	_, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, strings.Contains(err.Error(), "unknown method"))
}
