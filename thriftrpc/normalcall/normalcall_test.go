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

package normalcall

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	stservice_slim "github.com/cloudwego/kitex-tests/kitex_gen_slim/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote/codec/thrift"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

var (
	cli     stservice.Client
	cliSlim stservice_slim.Client
)

func TestMain(m *testing.M) {
	svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9001",
	}, nil)
	time.Sleep(time.Second)
	slimSvr := thriftrpc.RunSlimServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9003",
	}, nil)
	time.Sleep(time.Second)
	slimSvrWithFrugalConfigured := thriftrpc.RunSlimServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9004",
	}, nil, server.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalWrite|thrift.FrugalRead)))
	m.Run()
	svr.Stop()
	slimSvr.Stop()
	slimSvrWithFrugalConfigured.Stop()
}

func getKitexClient(p transport.Protocol, opts ...client.Option) stservice.Client {
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{":9001"},
		Protocol:          p,
		ConnMode:          thriftrpc.LongConnection,
	}, opts...)
}

func getSlimKitexClient(p transport.Protocol, hostPorts []string, opts ...client.Option) stservice_slim.Client {
	return thriftrpc.CreateSlimKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa.slim",
		HostPorts:         hostPorts,
		Protocol:          p,
		ConnMode:          thriftrpc.LongConnection,
	}, opts...)
}

func TestStTReq(t *testing.T) {
	cli = getKitexClient(transport.PurePayload)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
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

func TestStTReqWithTTHeader(t *testing.T) {
	cli = getKitexClient(transport.TTHeader)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)
}

func TestStTReqWithFramed(t *testing.T) {
	cli = getKitexClient(transport.Framed)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)
}

func TestStTReqWithTTHeaderFramed(t *testing.T) {
	cli = getKitexClient(transport.TTHeaderFramed)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err == nil, err)
	test.Assert(t, stReq.Str == stResp.Str)
}

func TestObjReq(t *testing.T) {
	cli = getKitexClient(transport.PurePayload)

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = "ObjReq"
	objResp, err := cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	test.Assert(t, objReq.FlagMsg == objResp.FlagMsg)
}

func TestException(t *testing.T) {
	cli = getKitexClient(transport.TTHeader)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	_, err := cli.TestException(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err != nil)
	test.Assert(t, err.Error() == "STException({Message:mock exception})")
	nr, ok := err.(*stability.STException)
	test.Assert(t, ok)
	test.Assert(t, nr.Message == "mock exception")
}

func TestVisitOneway(t *testing.T) {
	cli = getKitexClient(transport.TTHeaderFramed)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	err := cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	err = cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	err = cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)

	time.Sleep(200 * time.Millisecond)
	test.Assert(t, atomic.LoadInt32(&thriftrpc.CheckNum) == int32(3))
}

func TestRPCTimeoutPriority(t *testing.T) {
	cli = getKitexClient(transport.TTHeaderFramed, client.WithRPCTimeout(500*time.Millisecond))
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	durationStr := "300ms"
	stReq.MockCost = &durationStr
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(200*time.Millisecond))
	test.Assert(t, err != nil)
	test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout))
	test.Assert(t, strings.Contains(err.Error(), "timeout=200ms"))
	test.Assert(t, stResp == nil)
}

func TestDisablePoolForRPCInfo(t *testing.T) {
	backupState := rpcinfo.PoolEnabled()
	defer rpcinfo.EnablePool(backupState)

	t.Run("client", func(t *testing.T) {
		var ri rpcinfo.RPCInfo
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		cli = getKitexClient(transport.TTHeaderFramed, client.WithMiddleware(func(endpoint endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				ri = rpcinfo.GetRPCInfo(ctx)
				return nil // no need to send the real request
			}
		}))

		t.Run("enable", func(t *testing.T) {
			rpcinfo.EnablePool(true)
			_, _ = cli.TestSTReq(ctx, stReq)
			test.Assert(t, ri.From() == nil, ri.From()) // zeroed
		})

		t.Run("disable", func(t *testing.T) {
			rpcinfo.EnablePool(false)
			_, _ = cli.TestSTReq(ctx, stReq)
			test.Assert(t, ri.From() != nil, ri.From()) // should not be zeroed
		})
	})

	t.Run("server", func(t *testing.T) {
		var ri1, ri2 rpcinfo.RPCInfo
		svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{Network: "tcp", Address: ":9002"}, nil,
			server.WithMiddleware(func(endpoint endpoint.Endpoint) endpoint.Endpoint {
				return func(ctx context.Context, req, resp interface{}) (err error) {
					err = endpoint(ctx, req, resp)
					request := req.(utils.KitexArgs).GetFirstArgument().(*stability.STRequest)
					if request.Name == "1" {
						ri1 = rpcinfo.GetRPCInfo(ctx)
					} else {
						ri2 = rpcinfo.GetRPCInfo(ctx)
					}
					return err
				}
			}))
		time.Sleep(time.Second)
		defer svr.Stop()

		cli = getKitexClient(transport.TTHeaderFramed)
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())

		t.Run("enable", func(t *testing.T) {
			rpcinfo.EnablePool(true)

			stReq.Name = "1"
			_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(":9002"))
			test.Assert(t, err == nil, err)

			stReq.Name = "2"
			_, err = cli.TestSTReq(ctx, stReq, callopt.WithHostPort(":9002"))
			test.Assert(t, err == nil, err)

			addr1, addr2 := reflect.ValueOf(ri1).Pointer(), reflect.ValueOf(ri2).Pointer()
			test.Assertf(t, addr1 == addr2, "addr1: %v, addr2: %v", addr1, addr2)
		})

		t.Run("disable", func(t *testing.T) {
			rpcinfo.EnablePool(false)
			stReq.Name = "1"
			_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(":9002"))
			test.Assert(t, err == nil, err)

			stReq.Name = "2"
			_, err = cli.TestSTReq(ctx, stReq, callopt.WithHostPort(":9002"))
			test.Assert(t, err == nil, err)

			addr1, addr2 := reflect.ValueOf(ri1).Pointer(), reflect.ValueOf(ri2).Pointer()
			test.Assertf(t, addr1 != addr2, "addr1: %v, addr2: %v", addr1, addr2)
		})

	})
}

// When using slim template and users do not
func TestFrugalFallback(t *testing.T) {
	testCases := []struct {
		desc      string
		hostPorts []string
		opts      []client.Option
		expectErr bool
	}{
		{
			desc:      "use slim template, do not configure thrift codec type",
			hostPorts: []string{":9003"},
			opts:      nil,
		},
		{
			desc:      "use slim template, configure FastWrite | FastRead thrift codec",
			hostPorts: []string{":9003"},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FastWrite | thrift.FastRead)),
			},
		},
		{
			desc:      "use slim template, only configure Basic thrift codec to disable frugal",
			hostPorts: []string{":9003"},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.Basic)),
			},
			expectErr: true,
		},
		{
			desc:      "use slim template, configure FrugalWrite | FrugalRead thrift codec, connect to frugal configured server",
			hostPorts: []string{":9004"},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalWrite | thrift.FrugalRead)),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cliSlim = getSlimKitexClient(transport.TTHeader, tc.hostPorts, tc.opts...)
			ctx, stReq := thriftrpc.CreateSlimSTRequest(context.Background())
			for i := 0; i < 3; i++ {
				stResp, err := cliSlim.TestSTReq(ctx, stReq)
				if tc.expectErr {
					test.Assert(t, err != nil, err)
					continue
				}
				test.Assert(t, err == nil, err)
				test.Assert(t, stReq.Str == stResp.Str)
			}
		})
	}
}

func BenchmarkThriftCall(b *testing.B) {
	cli = getKitexClient(transport.TTHeader)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stReq.FlagMsg = strconv.Itoa(i)
		stResp, err := cli.TestSTReq(ctx, stReq)
		test.Assert(b, err == nil, err)
		test.Assert(b, stReq.FlagMsg == stResp.FlagMsg)

		objReq.FlagMsg = strconv.Itoa(i)
		objResp, err := cli.TestObjReq(ctx, objReq)
		test.Assert(b, err == nil, err)
		test.Assert(b, objReq.FlagMsg == objResp.FlagMsg)
	}
}

func BenchmarkThriftCallParallel(b *testing.B) {
	cli = getKitexClient(transport.PurePayload)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
			_, err := cli.TestException(ctx, stReq)
			test.Assert(b, err != nil)
			test.Assert(b, err.Error() == "STException({Message:mock exception})", err)
			nr, ok := err.(*stability.STException)
			test.Assert(b, ok)
			test.Assert(b, nr.Message == "mock exception")

			err = cli.VisitOneway(ctx, stReq)
			test.Assert(b, err == nil, err)
		}
	})
}

func BenchmarkTTHeaderParallel(b *testing.B) {
	cli = getKitexClient(transport.TTHeader)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
			stReq.FlagMsg = strconv.Itoa(i)
			stResp, err := cli.TestSTReq(ctx, stReq)
			test.Assert(b, err == nil, err)
			test.Assert(b, stReq.FlagMsg == stResp.FlagMsg)

			ctx, objReq := thriftrpc.CreateObjReq(context.Background())
			objReq.FlagMsg = strconv.Itoa(i)
			objResp, err := cli.TestObjReq(ctx, objReq)
			test.Assert(b, err == nil, err)
			test.Assert(b, objReq.FlagMsg == objResp.FlagMsg)
			i++
		}
	})
}
