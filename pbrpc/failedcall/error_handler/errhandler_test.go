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
	"reflect"
	"strings"
	"testing"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/stability/stservice"
	"github.com/cloudwego/kitex-tests/pbrpc"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/codes"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/status"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
	"google.golang.org/protobuf/proto"
)

var testaddr string

func TestMain(m *testing.M) {
	ln := serverutils.Listen()
	testaddr = ln.Addr().String()
	svr := pbrpc.RunServer(&pbrpc.ServerInitParam{
		Listener: ln,
	}, &STServiceHandler{}, server.WithMetaHandler(transmeta.ServerTTHeaderHandler))
	m.Run()
	svr.Stop()
}

func TestHandlerReturnNormalError(t *testing.T) {
	// Kitex Protobuf
	cli := getKitexClient(transport.Framed)
	ctx, stReq := pbrpc.CreateSTRequest(context.Background())
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

	// Kitex gRPC unary
	cli = getKitexClient(transport.GRPC)
	ctx, stReq = pbrpc.CreateSTRequest(context.Background())
	stReq.Name = normalErr.Error()
	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	de = err.(*kerrors.DetailedError)
	// wrap error is kerrors.ErrRemoteOrNetwork
	test.Assert(t, de.ErrorType() == kerrors.ErrRemoteOrNetwork)
	test.Assert(t, strings.Contains(err.Error(), "remote or network error: rpc error: code = 13 desc = mock handler normal err [biz error]"), err.Error())
	// internal is *status.Error
	se, ok := status.FromError(err)
	test.Assert(t, ok)
	test.Assert(t, se.Code() == codes.Internal, te.TypeID())
}

func TestHandlerReturnTransError(t *testing.T) {
	cli := getKitexClient(transport.TTHeader)
	ctx, stReq := pbrpc.CreateSTRequest(context.Background())
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

	// Kitex gRPC unary
	cli = getKitexClient(transport.GRPC)
	ctx, stReq = pbrpc.CreateSTRequest(context.Background())
	stReq.Name = kitexTransErr.Error()
	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	de = err.(*kerrors.DetailedError)
	// wrap error is kerrors.ErrRemoteOrNetwork
	test.Assert(t, de.ErrorType() == kerrors.ErrRemoteOrNetwork)
	test.Assert(t, strings.Contains(err.Error(), "remote or network error: rpc error: code = 1900 desc = mock handler TransError [biz error]"), err.Error())
	// internal is *status.Error
	se, ok := status.FromError(err)
	test.Assert(t, ok)
	test.Assert(t, uint32(se.Code()) == uint32(kitexTransErr.TypeID()))
}

func TestHandlerReturnStatusError(t *testing.T) {
	cli := getKitexClient(transport.TTHeader)
	ctx, stReq := pbrpc.CreateSTRequest(context.Background())
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

	// Kitex gRPC unary
	cli = getKitexClient(transport.GRPC)
	ctx, stReq = pbrpc.CreateSTRequest(context.Background())
	stReq.Name = grpcStatus.Error()
	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	de = err.(*kerrors.DetailedError)
	// wrap error is kerrors.ErrRemoteOrNetwork
	test.Assert(t, de.ErrorType() == kerrors.ErrRemoteOrNetwork)
	test.Assert(t, strings.Contains(err.Error(), "remote or network error: rpc error: code = 1900 desc = mock handler StatusError [biz error]"), err.Error())
	// internal is *status.Error
	se, ok := status.FromError(err)
	test.Assert(t, ok)
	test.Assert(t, se.Code() == grpcStatus.(*status.Error).GRPCStatus().Code())
}

func TestHandlerReturnBizError(t *testing.T) {
	bizErr := kerrors.NewBizStatusErrorWithExtra(502, "bad gateway", map[string]string{"version": "v1.0.0"})
	cli := getKitexClient(transport.TTHeader)
	ctx, stReq := pbrpc.CreateSTRequest(context.Background())
	stReq.Name = "bizErr"
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	bizerror, ok := kerrors.FromBizStatusError(err)
	test.Assert(t, ok)
	test.Assert(t, bizerror.BizStatusCode() == bizErr.BizStatusCode())
	test.Assert(t, bizerror.BizMessage() == bizErr.BizMessage())
	test.Assert(t, reflect.DeepEqual(bizerror.BizExtra(), bizErr.BizExtra()))

	// Kitex gRPC unary
	cli = getKitexClient(transport.GRPC)
	ctx, stReq = pbrpc.CreateSTRequest(context.Background())
	stReq.Name = "bizErr"
	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	bizerror, ok = kerrors.FromBizStatusError(err)
	test.Assert(t, ok)
	test.Assert(t, bizerror.BizStatusCode() == bizErr.BizStatusCode())
	test.Assert(t, bizerror.BizMessage() == bizErr.BizMessage())
	test.Assert(t, reflect.DeepEqual(bizerror.BizExtra(), bizErr.BizExtra()))

	// Kitex gRPC unary with details
	cli = getKitexClient(transport.GRPC)
	ctx, stReq = pbrpc.CreateSTRequest(context.Background())
	stReq.Name = "bizErrWithDetail"
	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	bizerror, ok = kerrors.FromBizStatusError(err)
	test.Assert(t, ok)
	test.Assert(t, bizerror.BizStatusCode() == 404)
	test.Assert(t, bizerror.BizMessage() == "not found")
	test.Assert(t, reflect.DeepEqual(bizerror.BizExtra(), map[string]string{"version": "v1.0.0"}))
	// only works for proto.Message
	if _, ok := any((*stability.STRequest)(nil)).(proto.Message); ok {
		str := bizerror.(status.Iface).GRPCStatus().Details()[0].(*stability.STRequest).Str
		test.Assert(t, reflect.DeepEqual(str, "hello world"))
	}
}

func TestHandlerPanic(t *testing.T) {
	cli := getKitexClient(transport.TTHeader)
	ctx, stReq := pbrpc.CreateSTRequest(context.Background())
	stReq.Name = panicStr
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	de := err.(*kerrors.DetailedError)
	// wrap error is kerrors.ErrRemoteOrNetwork
	test.Assert(t, de.ErrorType() == kerrors.ErrRemoteOrNetwork)
	test.Assert(t, strings.Contains(err.Error(), "happened in biz handler"), err.Error())
	// internal is *remote.TransError
	te := de.Unwrap().(*remote.TransError)
	test.Assert(t, te.TypeID() == remote.InternalError)

	// Kitex gRPC unary
	cli = getKitexClient(transport.GRPC)
	ctx, stReq = pbrpc.CreateSTRequest(context.Background())
	stReq.Name = panicStr
	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	de = err.(*kerrors.DetailedError)
	// wrap error is kerrors.ErrRemoteOrNetwork
	test.Assert(t, de.ErrorType() == kerrors.ErrRemoteOrNetwork)
	test.Assert(t, strings.Contains(err.Error(), "happened in biz handler"), err.Error())
	// internal is *status.Error
	se, ok := status.FromError(err)
	test.Assert(t, ok)
	test.Assert(t, se.Code() == codes.Internal)
}

func getKitexClient(p transport.Protocol) stservice.Client {
	return pbrpc.CreateKitexClient(&pbrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{testaddr},
		Protocol:          p,
		ConnMode:          pbrpc.LongConnection,
	}, client.WithMetaHandler(transmeta.ClientTTHeaderHandler))
}
