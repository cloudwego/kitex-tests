// Copyright 2024 CloudWeGo Authors
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

package tls

import (
	"context"
	"errors"
	"log"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	client_opt "github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"

	"github.com/cloudwego/kitex-tests/pkg/test"
)

func TestMain(m *testing.M) {
	err := generateCert()
	if err != nil {
		log.Printf("TLS test failed, error: %v", err)
		return
	}
	m.Run()
}

// Only Kitex gPRC client support TLS now.
func TestKitexClientGRPCTLSWithGRPCServer(t *testing.T) {
	hostport := "localhost:9020"
	svr, err := RunGRPCTLSServer(hostport)
	test.Assert(t, err == nil, err)
	defer svr.Stop()

	time.Sleep(time.Second)
	cfg, err := clientLoadTLSCredentials()
	test.Assert(t, err == nil, err)
	client, err := GetClient(hostport, client_opt.WithMiddleware(ServiceNameMW), client_opt.WithGRPCTLSConfig(cfg))
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

// For modifying package name to avoid collision
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

func generateCert() error {
	cmd := exec.Command("/bin/bash", "gen.sh")
	path, err := filepath.Abs("cert")
	if err != nil {
		return err
	}
	cmd.Dir = path
	return cmd.Run()
}
