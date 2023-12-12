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
	"net"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo/echoservice"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb"
	grpcpbservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb/pbservice"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb"
	kitexpbservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb/pbservice"
	cross_echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_cross/echo"
	cross_echoservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_cross/echo/echoservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/server"
)

func WithServerAddr(hostPort string) server.Option {
	addr, err := net.ResolveTCPAddr("tcp", hostPort)
	if err != nil {
		panic(err)
	}
	return server.WithServiceAddr(addr)
}

func WaitServer(hostPort string) {
	for begin := time.Now(); time.Since(begin) < time.Second; {
		if _, err := net.Dial("tcp", hostPort); err == nil {
			klog.Infof("server %s is up", hostPort)
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
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
	WaitServer(addr)
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
	WaitServer(addr)
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
	WaitServer(addr)
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
	WaitServer(addr)
	return svr
}

var (
	initialPort = int32(9000)
	thriftAddr  = addrAllocator()
	crossAddr   = addrAllocator()
	slimAddr    = addrAllocator()
	grpcAddr    = addrAllocator()
	pbAddr      = addrAllocator()
)

func addrAllocator() string {
	addr := "127.0.0.1:" + strconv.Itoa(int(atomic.LoadInt32(&initialPort)))
	atomic.AddInt32(&initialPort, 1)
	return addr
}

func TestMain(m *testing.M) {
	var thriftSvr, thriftCrossSvr, slimServer, grpcServer, pbServer server.Server
	go func() { thriftSvr = RunThriftServer(&EchoServiceImpl{}, thriftAddr) }()
	go func() { thriftCrossSvr = RunThriftCrossServer(&CrossEchoServiceImpl{}, crossAddr) }()
	go func() { grpcServer = RunGRPCPBServer(&GRPCPBServiceImpl{}, grpcAddr) }()
	go func() { pbServer = RunKitexPBServer(&KitexPBServiceImpl{}, pbAddr) }()
	go func() { slimServer = RunSlimThriftServer(&SlimEchoServiceImpl{}, slimAddr) }()
	defer func() {
		go thriftSvr.Stop()
		go grpcServer.Stop()
		go pbServer.Stop()
		go slimServer.Stop()
		go thriftCrossSvr.Stop()
	}()
	WaitServer(thriftAddr)
	WaitServer(crossAddr)
	WaitServer(grpcAddr)
	WaitServer(pbAddr)
	WaitServer(slimAddr)
	m.Run()
}
