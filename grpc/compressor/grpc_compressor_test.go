// Copyright 2023 CloudWeGo Authors
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

package compressor

import (
	"context"
	"errors"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex-tests/pkg/test"
	client_opt "github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"net"
	"testing"
	"time"
)

func TestKitexWithoutCompressor(t *testing.T) {
	hostport := "localhost:9020"
	addr, _ := net.ResolveTCPAddr("tcp", hostport)
	svr := servicea.NewServer(new(ServiceAImpl), server.WithServiceAddr(addr))
	go func() {
		err := svr.Run()
		test.Assert(t, err == nil, err)
	}()
	defer svr.Stop()
	time.Sleep(time.Second)
	client, err := GetClient(hostport)
	test.Assert(t, err == nil, err)
	resp, err := client.RunUnary()
	test.Assert(t, err == nil, err)
	test.Assert(t, resp != nil && resp.Message == "Kitex Hello!")
	resp, err = client.RunClientStream()
	test.Assert(t, err == nil, err)
	test.Assert(t, resp != nil && resp.Message == "all message: kitex-0, kitex-1, kitex-2")
	respArr, err := client.RunServerStream()
	test.Assert(t, err == nil, err)
	test.Assert(t, len(respArr) == 3 && respArr[0].Message == "kitex-0" && respArr[1].Message == "kitex-1" && respArr[2].Message == "kitex-2")
	respArr, err = client.RunBidiStream()
	test.Assert(t, err == nil, err)
	test.Assert(t, len(respArr) == 3 && respArr[0].Message == "kitex-0" && respArr[1].Message == "kitex-1" && respArr[2].Message == "kitex-2")
}

func TestKitexCompressor(t *testing.T) {
	hostport := "localhost:9021"
	addr, _ := net.ResolveTCPAddr("tcp", hostport)
	svr := servicea.NewServer(new(ServiceAImpl), server.WithServiceAddr(addr))
	go func() {
		err := svr.Run()
		test.Assert(t, err == nil, err)
	}()
	defer svr.Stop()
	time.Sleep(time.Second)
	client, err := GetClient(hostport)
	test.Assert(t, err == nil, err)
	resp, err := client.RunUnary(callopt.WithGRPCCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, resp != nil && resp.Message == "Kitex Hello!")
	resp, err = client.RunClientStream(callopt.WithGRPCCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, resp != nil && resp.Message == "all message: kitex-0, kitex-1, kitex-2")
	respArr, err := client.RunServerStream(callopt.WithGRPCCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, len(respArr) == 3 && respArr[0].Message == "kitex-0" && respArr[1].Message == "kitex-1" && respArr[2].Message == "kitex-2")
	respArr, err = client.RunBidiStream(callopt.WithGRPCCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, len(respArr) == 3 && respArr[0].Message == "kitex-0" && respArr[1].Message == "kitex-1" && respArr[2].Message == "kitex-2")
}

func TestKitexCompressorWithGRPCClient(t *testing.T) {
	hostport := "localhost:9022"
	addr, _ := net.ResolveTCPAddr("tcp", hostport)
	svr := servicea.NewServer(new(ServiceAImpl), server.WithServiceAddr(addr))
	go func() {
		err := svr.Run()
		test.Assert(t, err == nil, err)
	}()
	defer svr.Stop()
	time.Sleep(time.Second)

	conn, err := grpc.Dial(hostport, grpc.WithInsecure(), grpc.WithBlock())
	test.Assert(t, err == nil, err)
	defer conn.Close()
	client, err := GetGRPCClient(hostport)
	test.Assert(t, err == nil, err)
	resp, err := client.RunUnary(grpc.UseCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, resp != nil && resp.Message == "Grpc Hello!")
	resp, err = client.RunClientStream(grpc.UseCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, resp != nil && resp.Message == "all message: grpc-0, grpc-1, grpc-2")
	respArr, err := client.RunServerStream(grpc.UseCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, len(respArr) == 3 && respArr[0].Message == "grpc-0" && respArr[1].Message == "grpc-1" && respArr[2].Message == "grpc-2")
	respArr, err = client.RunBidiStream(grpc.UseCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, len(respArr) == 3 && respArr[0].Message == "grpc-0" && respArr[1].Message == "grpc-1" && respArr[2].Message == "grpc-2")
}

func ServiceNameMW(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request, response interface{}) error {

		ri := rpcinfo.GetRPCInfo(ctx)
		ink := ri.Invocation()
		if ink, ok := ink.(rpcinfo.InvocationSetter); ok {
			ink.SetPackageName("grpc_demo_2")
		} else {
			return errors.New("the interface Invocation doesn't implement InvocationSetter")
		}
		return next(ctx, request, response)
	}
}

func TestKitexCompressorWithGRPCServer(t *testing.T) {
	hostport := "localhost:9023"
	go func() {
		err := RunGRPCServer(hostport)
		test.Assert(t, err == nil, err)
	}()
	time.Sleep(time.Second)

	client, err := GetClient(hostport, client_opt.WithMiddleware(ServiceNameMW))
	test.Assert(t, err == nil, err)
	resp, err := client.RunUnary(callopt.WithGRPCCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, resp != nil && resp.Message == "Kitex Hello!")
	resp, err = client.RunClientStream(callopt.WithGRPCCompressor(gzip.Name))
	test.Assert(t, err == nil, t)
	test.Assert(t, resp != nil && resp.Message == "all message: kitex-0, kitex-1, kitex-2")
	respArr, err := client.RunServerStream(callopt.WithGRPCCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, len(respArr) == 3 && respArr[0].Message == "kitex-0" && respArr[1].Message == "kitex-1" && respArr[2].Message == "kitex-2")
	respArr, err = client.RunBidiStream(callopt.WithGRPCCompressor(gzip.Name))
	test.Assert(t, err == nil, err)
	test.Assert(t, len(respArr) == 3 && respArr[0].Message == "kitex-0" && respArr[1].Message == "kitex-1" && respArr[2].Message == "kitex-2")
}
