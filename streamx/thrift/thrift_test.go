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

package streamx_thrift

import (
	"log"
	"sync"
	"testing"

	"github.com/cloudwego/kitex/server"

	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/streamx"
)

var (
	thriftTestAddr        string
	thriftTestCancelAddr  string
	thriftTestTimeoutAddr string
)

func TestMain(m *testing.M) {
	var normalSrv server.Server
	var cancelSrv server.Server
	var timeoutSrv server.Server
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		thriftTestAddr = serverutils.NextListenAddr()
		normalSrv = streamx.RunNormalServer(thriftTestAddr)
		serverutils.Wait(thriftTestAddr)
	}()
	go func() {
		defer wg.Done()
		thriftTestCancelAddr = serverutils.NextListenAddr()
		cancelSrv = streamx.RunTestCancelServer(thriftTestCancelAddr)
		serverutils.Wait(thriftTestCancelAddr)
	}()
	go func() {
		defer wg.Done()
		thriftTestTimeoutAddr = serverutils.NextListenAddr()
		timeoutSrv = streamx.RunTestTimeoutServer(thriftTestTimeoutAddr)
		serverutils.Wait(thriftTestTimeoutAddr)
	}()
	wg.Wait()
	m.Run()
	normalSrv.Stop()
	cancelSrv.Stop()
	timeoutSrv.Stop()
	log.Print("streamx thrift test finished")
}

func TestNormal(t *testing.T) {
	streamx.TestThriftNormal(t, thriftTestAddr)
}

func TestCancel(t *testing.T) {
	streamx.TestThriftCancel(t, thriftTestCancelAddr)
}

func TestTimeout(t *testing.T) {
	streamx.TestThriftTimeout(t, thriftTestTimeoutAddr)
}
