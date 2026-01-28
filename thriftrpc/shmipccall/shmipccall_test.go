// Copyright 2025 CloudWeGo Authors
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

package shmipccall

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thriftrpc"
)

const (
	shmipcSockPath = "./kitex_tests_shmipc.sock"
	udsSockPath    = "./kitex_tests_uds.sock"
)

var testaddr string

func TestMain(m *testing.M) {
	// Cleanup socket files before starting
	cleanupSocketFiles()

	testaddr = shmipcSockPath

	svr := thriftrpc.RunSHMIPCServer(&thriftrpc.ServerInitParam{
		Listener: nil, // SHMIPC server will create its own listener
	}, nil)

	m.Run()
	svr.Stop()

	// Cleanup socket files after testing
	cleanupSocketFiles()
}

func cleanupSocketFiles() {
	// Remove socket files if they exist
	if _, err := os.Stat(shmipcSockPath); err == nil {
		os.Remove(shmipcSockPath)
	}
	if _, err := os.Stat(udsSockPath); err == nil {
		os.Remove(udsSockPath)
	}
}

func getSHMIPCKitexClient(p transport.Protocol, opts ...client.Option) stservice.Client {
	return thriftrpc.CreateSHMIPCKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa.shmipc",
		HostPorts:         []string{shmipcSockPath},
		Protocol:          p,
		ConnMode:          thriftrpc.LongConnection,
	}, opts...)
}

func testSTArgs(t *testing.T, req *stability.STRequest, resp *stability.STResponse) {
	test.Assert(t, req.Str == resp.Str)
	test.Assert(t, req.DefaultValue == stability.STRequest_DefaultValue_DEFAULT)
	test.Assert(t, resp.DefaultValue == stability.STResponse_DefaultValue_DEFAULT)
}

func protocolCheckMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			ri := rpcinfo.GetRPCInfo(ctx)
			p := ri.Config().TransportProtocol()
			err = next(ctx, req, resp)
			if ri.Config().TransportProtocol() != p {
				return errors.New("protocol not match")
			}
			return
		}
	}
}

func TestStTReqWithSHMIPC(t *testing.T) {
	cli := getSHMIPCKitexClient(transport.TTHeaderFramed, client.WithMiddleware(protocolCheckMiddleware()))

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err == nil, err)
	testSTArgs(t, stReq, stResp)
}

func TestObjReqWithSHMIPC(t *testing.T) {
	cli := getSHMIPCKitexClient(transport.TTHeaderFramed)

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = "ObjReq"
	objResp, err := cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	testObjArgs(t, objReq, objResp)
}

func testObjArgs(t *testing.T, req *instparam.ObjReq, resp *instparam.ObjResp) {
	test.Assert(t, req.FlagMsg == resp.FlagMsg)
	for _, v := range resp.MsgMap {
		test.Assert(t, v.DefaultValue == instparam.SubMessage_DefaultValue_DEFAULT)
	}
	for _, v := range resp.SubMsgs {
		test.Assert(t, v.DefaultValue == instparam.SubMessage_DefaultValue_DEFAULT)
	}
}

func TestExceptionWithSHMIPC(t *testing.T) {
	cli := getSHMIPCKitexClient(transport.TTHeaderFramed)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	_, err := cli.TestException(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err != nil)
	test.Assert(t, err.Error() == "STException({Message:mock exception})")
	nr, ok := err.(*stability.STException)
	test.Assert(t, ok)
	test.Assert(t, nr.Message == "mock exception")
}
