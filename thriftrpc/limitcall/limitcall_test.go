// Copyright 2022 CloudWeGo Authors
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

package failedcall

import (
	"context"
	"fmt"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/server"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/thriftrpc"
	"github.com/cloudwego/kitex/transport"
)

func getKitexClient(p transport.Protocol) stservice.Client {
	return thriftrpc.CreateKitexClient(&thriftrpc.ClientInitParam{
		TargetServiceName: "cloudwego.kitex.testa",
		HostPorts:         []string{":9001"},
		Protocol:          p,
		ConnMode:          thriftrpc.LongConnection,
	})
}

// TestConnectionLimit method tries to trigger server limitation of concurrency
func TestConnectionLimit(t *testing.T) {

	svr := RunServer(&limit.Option{MaxConnections: 10, MaxQPS: 200, UpdateControl: func(u limit.Updater) {}})
	defer svr.Stop()
	time.Sleep(time.Second)

	// init multi clients to concurrency send requests to server
	var wg sync.WaitGroup
	num := 15
	wg.Add(num)
	var success int32
	var failed int32
	for i := 0; i < num; i++ {
		go func() {
			cli := getKitexClient(transport.TTHeader)
			ctx, objReq := thriftrpc.CreateObjReq(context.Background())
			cost := "100ms"
			objReq.MockCost = &cost
			_, err := cli.TestObjReq(ctx, objReq)
			if err == nil {
				atomic.AddInt32(&success, 1)
			} else {
				atomic.AddInt32(&failed, 1)
			}
			fmt.Println(err)
			wg.Done()
		}()
	}
	wg.Wait()
	test.Assert(t, failed == 5)
	test.Assert(t, success == 10)
}

// TestQPSLimit method tries to trigger server limitation of qps
func TestQPSLimit(t *testing.T) {
	svr := RunServer(&limit.Option{MaxConnections: 100, MaxQPS: 10, UpdateControl: func(u limit.Updater) {}})
	defer svr.Stop()
	time.Sleep(time.Second)

	cli := getKitexClient(transport.TTHeader)
	var wg sync.WaitGroup
	num := 20
	wg.Add(num)
	var success int32
	var failed int32
	for i := 0; i < num; i++ {
		go func() {
			ctx, objReq := thriftrpc.CreateObjReq(context.Background())
			_, err := cli.TestObjReq(ctx, objReq)
			if err == nil {
				atomic.AddInt32(&success, 1)
			} else {
				atomic.AddInt32(&failed, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	test.Assert(t, failed == 10)
	test.Assert(t, success == 10)
}

func RunServer(limitOpt *limit.Option) server.Server {
	return thriftrpc.RunServer(&thriftrpc.ServerInitParam{
		Network: "tcp",
		Address: ":9001",
		Limiter: limitOpt,
	}, nil)
}
