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
	"net"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/grpc_demo/servicea"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

type ServerAHandler struct {
	grpc_demo.ServiceA
}

func (h *ServerAHandler) CallUnary(ctx context.Context, req *grpc_demo.Request) (*grpc_demo.Reply, error) {
	return &grpc_demo.Reply{Message: req.Name}, nil
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

	var ri rpcinfo.RPCInfo
	svr := servicea.NewServer(
		&ServerAHandler{},
		server.WithServiceAddr(serverAddr("127.0.0.1:9005")),
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
	time.Sleep(time.Second)
	defer svr.Stop()

	cli := servicea.MustNewClient("servicea", client.WithHostPorts("127.0.0.1:9005"))

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
