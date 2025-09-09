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
)

var (
	thriftTestAddr            string
	thriftTestCancelAddr      string
	thriftTestCancelProxyAddr string

	cancelTChan      = make(chan *testing.T, 1)
	cancelSendChan   = make(chan struct{}, 1)
	cancelFinishChan = make(chan struct{}, 1)
)

func TestMain(m *testing.M) {
	var normalSrv server.Server
	var cancelSrv server.Server
	var cancelProxySrv server.Server
	var wg sync.WaitGroup
	cancelSrvReady := make(chan struct{})
	wg.Add(3)
	go func() {
		defer wg.Done()
		thriftTestAddr = serverutils.NextListenAddr()
		normalSrv = runNormalServer(thriftTestAddr)
		serverutils.Wait(thriftTestAddr)
	}()
	go func() {
		defer wg.Done()
		thriftTestCancelAddr = serverutils.NextListenAddr()
		cancelSrv = runTestCancelServer(thriftTestCancelAddr)
		serverutils.Wait(thriftTestCancelAddr)
		close(cancelSrvReady)
	}()
	go func() {
		defer wg.Done()
		<-cancelSrvReady
		thriftTestCancelProxyAddr = serverutils.NextListenAddr()
		cancelProxySrv = runTestCancelServer(thriftTestCancelProxyAddr)
		serverutils.Wait(thriftTestCancelProxyAddr)
	}()
	wg.Wait()
	m.Run()
	normalSrv.Stop()
	cancelSrv.Stop()
	cancelProxySrv.Stop()
	log.Print("streamx thrift test finished")
}
