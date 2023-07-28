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

package error_handler

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/transport"
)

var cli stservice.Client

func TestMain(m *testing.M) {
	svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network:  "tcp",
		Address:  ":9001",
		ConnMode: thriftrpc.ConnectionMultiplexed,
	}, &STServiceHandler{})
	time.Sleep(time.Second)
	m.Run()
	svr.Stop()
}

func TestHandlerReturnNormalError(t *testing.T) {
	cli = getKitexClient(transport.PurePayload)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = normalErr.Error()
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	de := err.(*kerrors.DetailedError)
	// wrap error is kerrors.ErrRemoteOrNetwork
	test.Assert(t, de.ErrorType() == kerrors.ErrRemoteOrNetwork)
	test.Assert(t, strings.Contains(err.Error(), "remote or network error[remote]: biz error: mock handler normal err"), err.Error())
	// internal is *remote.TransError
	te := de.Unwrap().(*remote.TransError)
	test.Assert(t, te.TypeID() == remote.InternalError, te.TypeID())
}

func TestHandlerReturnTransError(t *testing.T) {
	cli = getKitexClient(transport.Framed)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = kitexTransErr.Error()
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	de := err.(*kerrors.DetailedError)
	// wrap error is kerrors.ErrRemoteOrNetwork
	test.Assert(t, de.ErrorType() == kerrors.ErrRemoteOrNetwork)
	test.Assert(t, strings.Contains(err.Error(), "remote or network error[remote]: mock handler TransError [biz error]"), err.Error())
	// internal is *remote.TransError
	te := de.Unwrap().(*remote.TransError)
	test.Assert(t, te.TypeID() == kitexTransErr.TypeID(), te.TypeID())
}

func TestHandlerReturnStatusError(t *testing.T) {
	cli = getKitexClient(transport.TTHeader)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = grpcStatus.Error()
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	de := err.(*kerrors.DetailedError)
	// wrap error is kerrors.ErrRemoteOrNetwork
	test.Assert(t, de.ErrorType() == kerrors.ErrRemoteOrNetwork)
	test.Assert(t, strings.Contains(err.Error(), "remote or network error[remote]: biz error: rpc error: code = 1900 desc = mock handler StatusError"), err.Error())
	// internal is *remote.TransError
	te := de.Unwrap().(*remote.TransError)
	test.Assert(t, te.TypeID() == remote.InternalError)
}

func TestHandlerPanic(t *testing.T) {
	cli = getKitexClient(transport.TTHeader)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = panicStr
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	de := err.(*kerrors.DetailedError)
	// wrap error is kerrors.ErrRemoteOrNetwork
	test.Assert(t, de.ErrorType() == kerrors.ErrRemoteOrNetwork)
	test.Assert(t, strings.Contains(err.Error(), "remote or network error[remote]: panic: [happened in biz handler, method=testSTReq] mock handler panic"), err.Error())
	// internal is *remote.TransError
	te := de.Unwrap().(*remote.TransError)
	test.Assert(t, te.TypeID() == remote.InternalError)
}

func getKitexClient(p transport.Protocol) stservice.Client {
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{":9001"},
		Protocol:          p,
		ConnMode:          thriftrpc.ConnectionMultiplexed,
	})
}
