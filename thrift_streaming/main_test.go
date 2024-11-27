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

package thrift_streaming

import (
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/kitex"
	"github.com/cloudwego/kitex/server"

	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/combine/combineservice"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo/echoservice"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb"
	grpcpbservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb/pbservice"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb"
	kitexpbservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb/pbservice"
	cross_echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_cross/echo"
	cross_echoservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_cross/echo/echoservice"
)

func WithServerAddr(hostPort string) server.Option {
	addr, err := net.ResolveTCPAddr("tcp", hostPort)
	if err != nil {
		panic(err)
	}
	return server.WithServiceAddr(addr)
}

func RunThriftServer(handler echo.EchoService, addr string, opts ...server.Option) server.Server {
	opts = append(opts, WithServerAddr(addr))
	opts = append(opts, server.WithExitWaitTime(time.Millisecond*10))
	svr := echoservice.NewServer(handler, opts...)
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	serverutils.Wait(addr)
	return svr
}

func RunCombineThriftServer(handler combineservice.CombineService, addr string, opts ...server.Option) server.Server {
	opts = append(opts, WithServerAddr(addr))
	opts = append(opts, server.WithExitWaitTime(time.Millisecond*10))
	svr := combineservice.NewServer(handler, opts...)
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	serverutils.Wait(addr)
	return svr
}

func RunThriftCrossServer(handler cross_echo.EchoService, addr string, opts ...server.Option) server.Server {
	opts = append(opts, WithServerAddr(addr))
	opts = append(opts, server.WithExitWaitTime(time.Millisecond*10))
	svr := cross_echoservice.NewServer(handler, opts...)
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	serverutils.Wait(addr)
	return svr
}

func RunGRPCPBServer(handler grpc_pb.PBService, addr string, opts ...server.Option) server.Server {
	opts = append(opts, WithServerAddr(addr))
	opts = append(opts, server.WithExitWaitTime(time.Millisecond*10))
	svr := grpcpbservice.NewServer(handler, opts...)
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	serverutils.Wait(addr)
	return svr
}

func RunKitexPBServer(handler kitex_pb.PBService, addr string, opts ...server.Option) server.Server {
	opts = append(opts, WithServerAddr(addr))
	opts = append(opts, server.WithExitWaitTime(time.Millisecond*10))
	svr := kitexpbservice.NewServer(handler, opts...)
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	serverutils.Wait(addr)
	return svr
}

var (
	thriftAddr  string
	crossAddr   string
	slimAddr    string
	grpcAddr    string
	pbAddr      string
	combineAddr string
)

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	wg.Add(6)
	thriftAddr = serverutils.NextListenAddr()
	crossAddr = serverutils.NextListenAddr()
	slimAddr = serverutils.NextListenAddr()
	grpcAddr = serverutils.NextListenAddr()
	pbAddr = serverutils.NextListenAddr()
	combineAddr = serverutils.NextListenAddr()
	var thriftSvr, thriftCrossSvr, slimServer, grpcServer, pbServer, combineServer server.Server
	go func() {
		thriftSvr = RunThriftServer(&EchoServiceImpl{}, thriftAddr)
		wg.Done()
	}()
	go func() {
		thriftCrossSvr = RunThriftCrossServer(&CrossEchoServiceImpl{}, crossAddr)
		wg.Done()
	}()
	go func() {
		grpcServer = RunGRPCPBServer(&GRPCPBServiceImpl{}, grpcAddr)
		wg.Done()
	}()
	go func() {
		pbServer = RunKitexPBServer(&KitexPBServiceImpl{}, pbAddr)
		wg.Done()
	}()
	go func() {
		slimServer = RunSlimThriftServer(&SlimEchoServiceImpl{}, slimAddr)
		wg.Done()
	}()
	go func() {
		combineServer = RunCombineThriftServer(&CombineServiceImpl{}, combineAddr)
		wg.Done()
	}()
	// wait for all servers to start
	wg.Wait()
	defer func() {
		if thriftSvr != nil {
			thriftSvr.Stop()
		}
		if thriftCrossSvr != nil {
			thriftCrossSvr.Stop()
		}
		if slimServer != nil {
			slimServer.Stop()
		}
		if grpcServer != nil {
			grpcServer.Stop()
		}
		if pbServer != nil {
			pbServer.Stop()
		}
		if combineServer != nil {
			combineServer.Stop()
		}
	}()
	log.Printf("testing Kitex %s", kitex.Version)
	m.Run()
}
