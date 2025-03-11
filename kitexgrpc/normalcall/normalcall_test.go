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

package normalcall

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	netutil "github.com/shirou/gopsutil/v3/net"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/clientutils"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
)

type ServerAHandler struct {
	grpc_demo.ServiceA
}

func (h *ServerAHandler) CallUnary(ctx context.Context, req *grpc_demo.Request) (*grpc_demo.Reply, error) {
	return &grpc_demo.Reply{Message: req.Name}, nil
}

func (h *ServerAHandler) CallClientStream(stream grpc_demo.ServiceA_CallClientStreamServer) (err error) {
	var msgs []string
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		msgs = append(msgs, req.Name)
	}
	return stream.SendAndClose(&grpc_demo.Reply{Message: "all message: " + strings.Join(msgs, ", ")})
}

func serverAddr(hostPort string) *net.TCPAddr {
	addr, err := net.ResolveTCPAddr("tcp", hostPort)
	if err != nil {
		panic(err)
	}
	return addr
}

func TestDisableRPCInfoReuse(t *testing.T) {
	backupState := rpcinfo.PoolEnabled()
	defer rpcinfo.EnablePool(backupState)

	ln := serverutils.Listen()
	addr := ln.Addr().String()

	var ri rpcinfo.RPCInfo
	svr := servicea.NewServer(
		&ServerAHandler{},
		server.WithListener(ln),
		server.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) error {
				ri = rpcinfo.GetRPCInfo(ctx)
				return next(ctx, req, resp)
			}
		}),
	)
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	defer svr.Stop()

	cli := servicea.MustNewClient("servicea", client.WithHostPorts(addr))
	defer clientutils.CallClose(cli)

	ctx := context.Background()

	t.Run("reuse", func(t *testing.T) {
		_, err := cli.CallUnary(ctx, &grpc_demo.Request{Name: "1"})
		test.Assert(t, err == nil, err)
		test.Assert(t, ri.Invocation().MethodName() == "", ri.Invocation().MethodName()) // zeroed
	})

	t.Run("disable reuse", func(t *testing.T) {
		rpcinfo.EnablePool(false)
		_, err := cli.CallUnary(ctx, &grpc_demo.Request{Name: "2"})
		test.Assert(t, err == nil, err)
		test.Assert(t, ri.Invocation().MethodName() != "", ri.Invocation().MethodName())
	})
}

func TestShortConnection(t *testing.T) {
	ln := serverutils.Listen()
	svrAddrStr := ln.Addr().String()
	svr := servicea.NewServer(
		&ServerAHandler{},
		server.WithListener(ln),
	)
	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()
	defer svr.Stop()

	cli, err := servicea.NewClient("servicea",
		client.WithHostPorts(svrAddrStr),
		client.WithShortConnection(),
	)
	test.Assert(t, err == nil)
	clientStream, err := cli.CallClientStream(context.Background())
	test.Assert(t, err == nil)
	err = clientStream.Send(&grpc_demo.Request{Name: "1"})
	test.Assert(t, err == nil)
	_, err = clientStream.CloseAndRecv()
	test.Assert(t, err == nil)
	time.Sleep(20 * time.Millisecond)

	// the connection should not exist after the call
	_, port, _ := net.SplitHostPort(svrAddrStr)
	exist, err := checkEstablishedConnection(port)
	if err != nil {
		klog.Warnf("check established connection failed, error=%v", err)
	} else {
		test.Assert(t, err == nil)
		test.Assert(t, exist == false)
	}
}

// return true if current pid established connection with the input remote ip and port
func checkEstablishedConnection(port string) (bool, error) {
	pid := os.Getpid()
	conns, err := netutil.ConnectionsPid("all", int32(pid))
	if err != nil {
		return false, err
	}
	for _, c := range conns {
		if c.Status == "ESTABLISHED" {
			if strconv.Itoa(int(c.Raddr.Port)) == port &&
				net.ParseIP(c.Raddr.IP).IsLoopback() { // only need to check IsLoopback for localhost
				fmt.Println(c.Laddr, c.Raddr)
				return true, nil
			}
		}
	}
	return false, nil
}
