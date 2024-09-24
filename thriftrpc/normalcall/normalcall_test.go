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
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/bytedance/gopkg/lang/fastrand"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/circuitbreak"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote/codec"
	"github.com/cloudwego/kitex/pkg/remote/codec/thrift"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	stservice_noDefSerdes "github.com/cloudwego/kitex-tests/kitex_gen_noDefSerdes/thrift/stability/stservice"
	stservice_slim "github.com/cloudwego/kitex-tests/kitex_gen_slim/thrift/stability/stservice"
	tutils "github.com/cloudwego/kitex-tests/pkg/utils"
	"github.com/cloudwego/kitex-tests/thriftrpc"
)

var testaddr string

func TestMain(m *testing.M) {
	testaddr = serverutils.NextListenAddr()
	svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: testaddr,
	}, nil)
	serverutils.Wait(testaddr)

	m.Run()
	svr.Stop()
}

func getKitexClient(p transport.Protocol, opts ...client.Option) stservice.Client {
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{testaddr},
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

func getNoDefSerdesKitexClient(p transport.Protocol, hostPorts []string, opts ...client.Option) stservice_noDefSerdes.Client {
	return thriftrpc.CreateNoDefSerdesKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa.noDefSerdes",
		HostPorts:         hostPorts,
		Protocol:          p,
		ConnMode:          thriftrpc.LongConnection,
	}, opts...)
}

func testSTArgs(t *testing.T, req *stability.STRequest, resp *stability.STResponse) {
	test.Assert(t, req.Str == resp.Str)
	test.Assert(t, req.DefaultValue == stability.STRequest_DefaultValue_DEFAULT)
	test.Assert(t, resp.DefaultValue == stability.STResponse_DefaultValue_DEFAULT)
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

func TestStTReq(t *testing.T) {
	cli := getKitexClient(transport.PurePayload)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	testSTArgs(t, stReq, stResp)

	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	testSTArgs(t, stReq, stResp)

	stResp, err = cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	testSTArgs(t, stReq, stResp)
}

func TestStTReqWithTTHeader(t *testing.T) {
	cli := getKitexClient(transport.TTHeader)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq)
	test.Assert(t, err == nil, err)
	testSTArgs(t, stReq, stResp)
}

func TestStTReqWithFramed(t *testing.T) {
	cli := getKitexClient(transport.Framed)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err == nil, err)
	testSTArgs(t, stReq, stResp)
}

func TestStTReqWithTTHeaderFramed(t *testing.T) {
	cli := getKitexClient(transport.TTHeaderFramed)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	stResp, err := cli.TestSTReq(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err == nil, err)
	testSTArgs(t, stReq, stResp)
}

func TestObjReq(t *testing.T) {
	cli := getKitexClient(transport.PurePayload)

	ctx, objReq := thriftrpc.CreateObjReq(context.Background())
	objReq.FlagMsg = "ObjReq"
	objResp, err := cli.TestObjReq(ctx, objReq)
	test.Assert(t, err == nil, err)
	testObjArgs(t, objReq, objResp)
}

func TestException(t *testing.T) {
	cli := getKitexClient(transport.TTHeader)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	_, err := cli.TestException(ctx, stReq, callopt.WithRPCTimeout(1*time.Second))
	test.Assert(t, err != nil)
	test.Assert(t, err.Error() == "STException({Message:mock exception})")
	nr, ok := err.(*stability.STException)
	test.Assert(t, ok)
	test.Assert(t, nr.Message == "mock exception")
}

func TestVisitOneway(t *testing.T) {
	cli := getKitexClient(transport.TTHeaderFramed)
	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	err := cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	err = cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)

	ctx, stReq = thriftrpc.CreateSTRequest(context.Background())
	err = cli.VisitOneway(ctx, stReq)
	test.Assert(t, err == nil, err)

	time.Sleep(20 * time.Millisecond)
	test.Assert(t, atomic.LoadInt32(&thriftrpc.CheckNum) == int32(3))
}

func TestRPCTimeoutPriority(t *testing.T) {
	cli := getKitexClient(transport.TTHeaderFramed, client.WithRPCTimeout(500*time.Millisecond))
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
		cli := getKitexClient(transport.TTHeaderFramed, client.WithMiddleware(func(endpoint endpoint.Endpoint) endpoint.Endpoint {
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
		disablePoolAddr := serverutils.NextListenAddr()
		var ri1, ri2 rpcinfo.RPCInfo
		svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{Network: "tcp", Address: disablePoolAddr}, nil,
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
		defer svr.Stop()
		serverutils.Wait(disablePoolAddr)

		cli := getKitexClient(transport.TTHeaderFramed)
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())

		t.Run("enable", func(t *testing.T) {
			rpcinfo.EnablePool(true)

			stReq.Name = "1"
			_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(disablePoolAddr))
			test.Assert(t, err == nil, err)

			stReq.Name = "2"
			_, err = cli.TestSTReq(ctx, stReq, callopt.WithHostPort(disablePoolAddr))
			test.Assert(t, err == nil, err)

			addr1, addr2 := reflect.ValueOf(ri1).Pointer(), reflect.ValueOf(ri2).Pointer()
			test.Assertf(t, addr1 == addr2, "addr1: %v, addr2: %v", addr1, addr2)
		})

		t.Run("disable", func(t *testing.T) {
			rpcinfo.EnablePool(false)
			stReq.Name = "1"
			_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(disablePoolAddr))
			test.Assert(t, err == nil, err)

			stReq.Name = "2"
			_, err = cli.TestSTReq(ctx, stReq, callopt.WithHostPort(disablePoolAddr))
			test.Assert(t, err == nil, err)

			addr1, addr2 := reflect.ValueOf(ri1).Pointer(), reflect.ValueOf(ri2).Pointer()
			test.Assertf(t, addr1 != addr2, "addr1: %v, addr2: %v", addr1, addr2)
		})
	})
}

// When using slim template and users do not
func TestFrugalFallback(t *testing.T) {
	slimAddr := serverutils.NextListenAddr()
	s0 := thriftrpc.RunSlimServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: slimAddr,
	}, nil)
	defer s0.Stop()
	serverutils.Wait(slimAddr)

	slimFrugalAddr := serverutils.NextListenAddr()
	s1 := thriftrpc.RunSlimServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: slimFrugalAddr,
	}, nil, server.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalWrite|thrift.FrugalRead)))
	serverutils.Wait(slimFrugalAddr)
	defer s1.Stop()

	testCases := []struct {
		desc      string
		hostPorts []string
		opts      []client.Option
		expectErr bool
	}{
		{
			desc:      "use slim template, do not configure thrift codec type",
			hostPorts: []string{slimAddr},
			opts:      nil,
		},
		{
			desc:      "use slim template, configure FastWrite | FastRead thrift codec",
			hostPorts: []string{slimAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FastWrite | thrift.FastRead)),
			},
		},
		{
			desc:      "use slim template, only configure Basic thrift codec to disable frugal",
			hostPorts: []string{slimAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.Basic)),
			},
			expectErr: true,
		},
		{
			desc:      "use slim template, configure FrugalWrite | FrugalRead thrift codec, connect to frugal configured server",
			hostPorts: []string{slimFrugalAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalWrite | thrift.FrugalRead)),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cli := getSlimKitexClient(transport.TTHeader, tc.hostPorts, tc.opts...)
			ctx, stReq := thriftrpc.CreateSlimSTRequest(context.Background())
			for i := 0; i < 3; i++ {
				stResp, err := cli.TestSTReq(ctx, stReq)
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

func TestCircuitBreakerCustomErrorTypeFunc(t *testing.T) {
	cli := getKitexClient(transport.TTHeader,
		client.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				return nil
			}
		}),
		client.WithCircuitBreaker(circuitbreak.NewCBSuite(
			func(ri rpcinfo.RPCInfo) string {
				return fmt.Sprintf("%s:%s", ri.To().ServiceName(), ri.To().Method())
			},
			circuitbreak.WithServiceGetErrorType(
				func(ctx context.Context, request, response interface{}, err error) circuitbreak.ErrorType {
					return circuitbreak.TypeFailure
				},
			),
		)),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	fuseCount := 0
	for i := 0; i < 300; i++ { // minSample = 200
		if _, err := cli.TestSTReq(ctx, stReq); kerrors.IsKitexError(err) { // circuit breaker err
			fuseCount += 1
		}
	}
	test.Assert(t, fuseCount >= 100, fuseCount)
}

func TestCircuitBreakerCustomInstanceErrorTypeFunc(t *testing.T) {
	cli := getKitexClient(transport.TTHeader,
		client.WithInstanceMW(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				return nil
			}
		}),
		client.WithCircuitBreaker(circuitbreak.NewCBSuite(
			func(ri rpcinfo.RPCInfo) string {
				return fmt.Sprintf("%s:%s", ri.To().ServiceName(), ri.To().Method())
			},
			circuitbreak.WithInstanceGetErrorType(
				func(ctx context.Context, request, response interface{}, err error) circuitbreak.ErrorType {
					return circuitbreak.TypeFailure
				},
			),
		)),
	)

	ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
	fuseCount := 0
	for i := 0; i < 300; i++ { // minSample = 200
		if _, err := cli.TestSTReq(ctx, stReq); kerrors.IsKitexError(err) { // circuit breaker err
			fuseCount += 1
		}
	}
	test.Assert(t, fuseCount >= 100, fuseCount)
}

func TestNoDefaultSerdes(t *testing.T) {
	frugalOnlyAddr := serverutils.NextListenAddr()
	fastcodecOnlyAddr := serverutils.NextListenAddr()
	s0 := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: frugalOnlyAddr,
	}, nil, server.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalWrite|thrift.FrugalRead|thrift.EnableSkipDecoder)))
	defer s0.Stop()

	s1 := thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: fastcodecOnlyAddr,
	}, nil, server.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FastWrite|thrift.FastRead|thrift.EnableSkipDecoder)))
	defer s1.Stop()

	serverutils.Wait(frugalOnlyAddr)
	serverutils.Wait(fastcodecOnlyAddr)

	testCases := []struct {
		desc      string
		hostPorts []string
		opts      []client.Option
		expectErr bool
	}{
		{
			desc:      "use FastCodec and SkipDecoder, connect to Frugal and SkipDecoder enabled server",
			hostPorts: []string{frugalOnlyAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FastWrite | thrift.FastRead | thrift.EnableSkipDecoder)),
			},
		},
		{
			desc:      "use Frugal and SkipDecoder, connect to Frugal and SkipDecoder enabled server",
			hostPorts: []string{frugalOnlyAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalWrite | thrift.FrugalRead | thrift.EnableSkipDecoder)),
			},
		},
		{
			desc:      "use FastCodec and SkipDecoder, connect to FastCodec and SkipDecoder enabled server",
			hostPorts: []string{fastcodecOnlyAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FastWrite | thrift.FastRead | thrift.EnableSkipDecoder)),
			},
		},
		{
			desc:      "use Frugal and SkipDecoder, connect to FastCodec and SkipDecoder enabled server",
			hostPorts: []string{fastcodecOnlyAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalWrite | thrift.FrugalRead | thrift.EnableSkipDecoder)),
			},
		},
		{
			desc:      "use FastCodec, connect to Frugal and SkipDecoder enabled server",
			hostPorts: []string{frugalOnlyAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FastWrite | thrift.FastRead)),
			},
			expectErr: true,
		},
		{
			desc:      "use Frugal, connect to Frugal and SkipDecoder enabled server",
			hostPorts: []string{frugalOnlyAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalWrite | thrift.FrugalRead)),
			},
			expectErr: true,
		},
		{
			desc:      "use FastCodec, connect to FastCodec and SkipDecoder enabled server",
			hostPorts: []string{fastcodecOnlyAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FastWrite | thrift.FastRead)),
			},
			expectErr: true,
		},
		{
			desc:      "use Frugal, connect to FastCodec and SkipDecoder enabled server",
			hostPorts: []string{fastcodecOnlyAddr},
			opts: []client.Option{
				client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalWrite | thrift.FrugalRead)),
			},
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cli := getNoDefSerdesKitexClient(transport.PurePayload, tc.hostPorts, tc.opts...)
			ctx, stReq := thriftrpc.CreateNoDefSerdesSTRequest(context.Background())
			for i := 0; i < 3; i++ {
				stResp, err := cli.TestSTReq(ctx, stReq)
				if !tc.expectErr {
					test.Assert(t, err == nil, err)
					test.Assert(t, stReq.Str == stResp.Str)
				} else {
					test.Assert(t, err != nil)
				}
			}
		})
	}
}

func TestCRC32PayloadValidator(t *testing.T) {
	codecOpt := client.WithCodec(codec.NewDefaultCodecWithConfig(codec.CodecConfig{CRC32Check: true}))
	crcClient := getKitexClient(transport.TTHeaderFramed, codecOpt)

	t.Run("serverWithoutCRC", func(t *testing.T) {
		// request server without crc32 check config
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		for i := 0; i < 10; i++ {
			_, err := crcClient.TestSTReq(ctx, stReq, callopt.WithHostPort(testaddr))
			test.Assert(t, err == nil, err)
		}
	})

	t.Run("serverWithCRC", func(t *testing.T) {
		// request server with crc config
		addr := serverutils.NextListenAddr()
		svrCodecOpt := server.WithCodec(codec.NewDefaultCodecWithConfig(codec.CodecConfig{CRC32Check: true}))
		svrWithCRC := thriftrpc.RunServer(&thriftrpc.ServerInitParam{Network: "tcp", Address: addr}, nil, svrCodecOpt)
		defer svrWithCRC.Stop()
		serverutils.Wait(addr)

		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())
		for i := 0; i < 10; i++ {
			_, err := crcClient.TestSTReq(ctx, stReq, callopt.WithHostPort(addr))
			test.Assert(t, err == nil, err)
		}
	})
}

func BenchmarkThriftCall(b *testing.B) {
	cli := getKitexClient(transport.TTHeader)
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
	cli := getKitexClient(transport.PurePayload)
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
	cli := getKitexClient(transport.TTHeader)
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

func BenchmarkThriftWithComplexData(b *testing.B) {
	cli := getKitexClient(transport.Framed)
	ctx, objReq := createComplexObjReq(context.Background())

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cli.TestObjReq(ctx, objReq)
		}
	})
}

func createComplexObjReq(ctx context.Context) (context.Context, *instparam.ObjReq) {
	id := int64(fastrand.Int31n(100))
	smallSubMsg := &instparam.SubMessage{
		Id:    &id,
		Value: stringPtr(tutils.RandomString(10)),
	}
	subMsg1K := &instparam.SubMessage{
		Id:    &id,
		Value: stringPtr(tutils.RandomString(1024)),
	}

	subMsgList2Items := []*instparam.SubMessage{smallSubMsg, smallSubMsg}

	msg := instparam.NewMessage()
	msg.Id = &id
	msg.Value = stringPtr(tutils.RandomString(1024))
	msg.SubMessages = subMsgList2Items

	msgMap := make(map[*instparam.Message]*instparam.SubMessage)
	for i := 0; i < 5; i++ {
		msgMap[instparam.NewMessage()] = subMsg1K
	}

	subMsgList100Items := make([]*instparam.SubMessage, 100)
	for i := 0; i < len(subMsgList100Items); i++ {
		subMsgList100Items[i] = smallSubMsg
	}

	req := instparam.NewObjReq()
	req.Msg = msg
	req.MsgMap = msgMap
	req.SubMsgs = subMsgList100Items
	req.MsgSet = []*instparam.Message{msg}

	ctx = metainfo.WithValue(ctx, "TK", "TV")
	ctx = metainfo.WithPersistentValue(ctx, "PK", "PV")
	return ctx, req
}

func stringPtr(v string) *string { return &v }
