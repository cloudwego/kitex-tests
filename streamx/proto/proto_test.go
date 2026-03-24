// Copyright 2025 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package streamx_proto_test

import (
	"log"
	"sync"
	"testing"

	"github.com/cloudwego/kitex/server"

	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/streamx"
)

var (
	pbTestAddr       string
	pbCancelAddr     string
	pbTimeoutAddr    string
	pbGRPCCompatAddr string
)

func TestMain(m *testing.M) {
	var normalSrv server.Server
	var cancelSrv server.Server
	var timeoutSrv server.Server
	var grpcCompatSrv server.Server
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		pbTestAddr = serverutils.NextListenAddr()
		normalSrv = streamx.RunNormalServer(pbTestAddr)
		serverutils.Wait(pbTestAddr)
	}()
	go func() {
		defer wg.Done()
		pbCancelAddr = serverutils.NextListenAddr()
		cancelSrv = streamx.RunTestCancelServer(pbCancelAddr)
		serverutils.Wait(pbCancelAddr)
	}()
	go func() {
		defer wg.Done()
		pbTimeoutAddr = serverutils.NextListenAddr()
		timeoutSrv = streamx.RunTestTimeoutServer(pbTimeoutAddr)
		serverutils.Wait(pbTimeoutAddr)
	}()
	go func() {
		defer wg.Done()
		pbGRPCCompatAddr = serverutils.NextListenAddr()
		grpcCompatSrv = streamx.RunGRPCCompatServer(pbGRPCCompatAddr)
		serverutils.Wait(pbGRPCCompatAddr)
	}()
	wg.Wait()
	m.Run()
	normalSrv.Stop()
	cancelSrv.Stop()
	timeoutSrv.Stop()
	grpcCompatSrv.Stop()
	log.Print("streamx pb test finished")
}

func TestNormal(t *testing.T) {
	streamx.TestPbNormal(t, pbTestAddr)
}

func TestCancel(t *testing.T) {
	streamx.TestPbCancel(t, pbCancelAddr)
}

func TestTimeout(t *testing.T) {
	streamx.TestPbTimeout(t, pbTimeoutAddr)
}

func TestGRPCCompat(t *testing.T) {
	streamx.TestPbGRPCCompat(t, pbGRPCCompatAddr)
}
