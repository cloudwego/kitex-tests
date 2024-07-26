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
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/fallback"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

var cli stservice.Client

func TestMain(m *testing.M) {
	svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: "localhost:9002",
	}, &STServiceHandler{}, server.WithMetaHandler(transmeta.ServerTTHeaderHandler))
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

func TestHandlerReturnBizStatusError(t *testing.T) {
	cli = getKitexClient(transport.TTHeader)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = bizErr.Error()
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err != nil)
	test.Assert(t, stResp == nil)
	bizerror, ok := kerrors.FromBizStatusError(err)
	test.Assert(t, ok)
	test.Assert(t, bizerror.BizStatusCode() == bizErr.BizStatusCode())
	test.Assert(t, bizerror.BizMessage() == bizErr.BizMessage())
	test.Assert(t, reflect.DeepEqual(bizerror.BizExtra(), bizErr.BizExtra()))
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
	test.Assert(t, strings.Contains(err.Error(), "happened in biz handler"), err.Error())
	// internal is *remote.TransError
	te := de.Unwrap().(*remote.TransError)
	test.Assert(t, te.TypeID() == remote.InternalError)
}

func TestFallback4PanicError(t *testing.T) {
	fallbackStr := "mock result"
	fbFunc := func(ctx context.Context, args utils.KitexArgs, result utils.KitexResult, err error) (fbErr error) {
		test.Assert(t, errors.Is(err, kerrors.ErrRemoteOrNetwork))
		test.Assert(t, strings.Contains(err.Error(), "happened in biz handler"))
		test.Assert(t, err.(*kerrors.DetailedError).Unwrap().(*remote.TransError).TypeID() == remote.InternalError)
		result.SetSuccess(&stability.STResponse{Str: fallbackStr})
		return
	}
	cli = getKitexClient(transport.TTHeader, client.WithFallback(fallback.ErrorFallback(fbFunc)))
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = panicStr
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil)
	test.Assert(t, stResp.Str == fallbackStr, stResp.Str)
}

func TestFallback4Timeout(t *testing.T) {
	fallbackStr := "mock result"
	fbFunc := func(ctx context.Context, req, resp interface{}, err error) (fbResp interface{}, fbErr error) {
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		return &stability.STResponse{Str: fallbackStr}, nil
	}
	cli = getKitexClient(transport.TTHeader, client.WithFallback(fallback.ErrorFallback(fallback.UnwrapHelper(fbFunc))))
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = timeout
	waitMS := "100ms"
	stReq.MockCost = &waitMS
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(20*time.Millisecond))
	test.Assert(t, err == nil)
	test.Assert(t, stResp.Str == fallbackStr, stResp.Str)
}

func TestFallback4TimeoutWithCallopt(t *testing.T) {
	cliFallbackStr := "client mock result"
	callFallbackStr := "call mock result"
	cliFallbackExecuted := false
	cliFBFunc := func(ctx context.Context, req, resp interface{}, err error) (fbResp interface{}, fbErr error) {
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		cliFallbackExecuted = true
		return &stability.STResponse{Str: cliFallbackStr}, nil
	}
	callOptFBFunc := func(ctx context.Context, req, resp interface{}, err error) (fbResp interface{}, fbErr error) {
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		return &stability.STResponse{Str: callFallbackStr}, nil
	}
	cli = getKitexClient(transport.TTHeader,
		client.WithRPCTimeout(20*time.Millisecond),
		client.WithFallback(fallback.TimeoutAndCBFallback(fallback.UnwrapHelper(cliFBFunc))))
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = timeout
	waitMS := "100ms"
	stReq.MockCost = &waitMS
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithFallback(fallback.TimeoutAndCBFallback(fallback.UnwrapHelper(callOptFBFunc))))
	test.Assert(t, err == nil)
	test.Assert(t, stResp.Str == callFallbackStr, stResp.Str)
	test.Assert(t, !cliFallbackExecuted)
}

func TestFallbackEnableReportAsFallback(t *testing.T) {
	errForReportIsNil := false
	tracerFinishFunc := func(ctx context.Context) {
		if rpcinfo.GetRPCInfo(ctx).Stats().Error() == nil {
			errForReportIsNil = true
		} else {
			errForReportIsNil = false
		}
	}

	cliFallbackStr := "client mock result"
	cliFBFunc := func(ctx context.Context, req, resp interface{}, err error) (fbResp interface{}, fbErr error) {
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
		return &stability.STResponse{Str: cliFallbackStr}, nil
	}
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = timeout
	waitMS := "100ms"
	stReq.MockCost = &waitMS

	// case 1: no EnableReportAsFallback, errForReportIsNil is false
	cli = getKitexClient(transport.TTHeader,
		client.WithRPCTimeout(20*time.Millisecond),
		client.WithFallback(fallback.TimeoutAndCBFallback(fallback.UnwrapHelper(cliFBFunc))),
		client.WithTracer(&mockTracer{startFunc: nil, finishFunc: tracerFinishFunc}))
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil)
	test.Assert(t, stResp.Str == cliFallbackStr, stResp.Str)
	test.Assert(t, !errForReportIsNil)

	// case 2: EnableReportAsFallback, errForReportIsNil is true
	cli = getKitexClient(transport.TTHeader,
		client.WithRPCTimeout(20*time.Millisecond),
		client.WithFallback(fallback.TimeoutAndCBFallback(fallback.UnwrapHelper(cliFBFunc)).EnableReportAsFallback()),
		client.WithTracer(&mockTracer{startFunc: nil, finishFunc: tracerFinishFunc}))
	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil)
	test.Assert(t, stResp.Str == cliFallbackStr, stResp.Str)
	test.Assert(t, errForReportIsNil)

}

func TestFallback4Resp(t *testing.T) {
	cliFallbackStr := "client mock result"
	fbFunc := func(ctx context.Context, req, resp interface{}, err error) (fbResp interface{}, fbErr error) {
		test.Assert(t, err == nil)
		test.Assert(t, resp.(*stability.STResponse).BaseResp.StatusCode == int32(bizStatusErrCode))
		return &stability.STResponse{Str: cliFallbackStr}, nil
	}
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stReq.Name = baseRespAsBizStatusErr

	// case 1: return mock resp
	cli = getKitexClient(transport.Framed,
		client.WithFallback(fallback.NewFallbackPolicy(fallback.UnwrapHelper(fbFunc))))
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil)
	test.Assert(t, stResp.Str == cliFallbackStr, stResp.Str)

	// case 2: bad case - original error is nil but return error in fallback, won't report err even if EnableReportAsFallback
	errForReportIsNil := false
	mockErr := errors.New("mock")
	tracerFinishFunc := func(ctx context.Context) {
		if rpcinfo.GetRPCInfo(ctx).Stats().Error() == nil {
			errForReportIsNil = true
		} else {
			errForReportIsNil = false
		}
	}
	errFBFunc := func(ctx context.Context, req, resp interface{}, err error) (fbResp interface{}, fbErr error) {
		test.Assert(t, err == nil)
		test.Assert(t, resp.(*stability.STResponse).BaseResp.StatusCode == int32(bizStatusErrCode))
		return nil, mockErr
	}
	cli = getKitexClient(transport.Framed,
		client.WithFallback(fallback.NewFallbackPolicy(fallback.UnwrapHelper(errFBFunc)).EnableReportAsFallback()),
		client.WithTracer(&mockTracer{startFunc: nil, finishFunc: tracerFinishFunc}))
	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == mockErr)
	test.Assert(t, stResp == nil)
	test.Assert(t, errForReportIsNil)
}

func getKitexClient(p transport.Protocol, opts ...client.Option) stservice.Client {
	opts = append(opts, client.WithMetaHandler(transmeta.ClientTTHeaderHandler), client.WithTransportProtocol(transport.TTHeader))
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{":9002"},
		Protocol:          p,
		ConnMode:          thriftrpc.LongConnection,
	}, opts...)
}

type mockTracer struct {
	startFunc  func(ctx context.Context) context.Context
	finishFunc func(ctx context.Context)
}

func (m mockTracer) Start(ctx context.Context) context.Context {
	if m.startFunc != nil {
		return m.startFunc(ctx)
	}
	return ctx
}

func (m mockTracer) Finish(ctx context.Context) {
	if m.finishFunc != nil {
		m.finishFunc(ctx)
	}
}
