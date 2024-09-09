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

package abc

import (
	"context"
	"github.com/cloudwego/kitex-tests/common"
	common2 "github.com/cloudwego/kitex-tests/kitexgrpc/failure/common"
	"github.com/cloudwego/kitex-tests/kitexgrpc/failure/consts"
	"github.com/cloudwego/kitex-tests/kitexgrpc/failure/servicea"
	"github.com/cloudwego/kitex-tests/kitexgrpc/failure/serviceb"
	"github.com/cloudwego/kitex-tests/kitexgrpc/failure/servicec"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	svrC := servicec.InitServiceCServer()
	go func() {
		defer svrC.Stop()
		err := svrC.Run()
		if err != nil {
			panic(err)
		}
	}()
	svrB := serviceb.InitServiceBServer()
	go func() {
		defer svrB.Stop()
		err := svrB.Run()
		if err != nil {
			panic(err)
		}
	}()

	common.WaitServer(consts.ServiceCAddr)
	common.WaitServer(consts.ServiceBAddr)

	m.Run()
}

func TestABC(t *testing.T) {
	cli, err := servicea.InitServiceAClient()
	test.Assert(t, err == nil, err)

	inj := common2.Injector{
		Settings: []common2.Setting{common2.NewCloseConnSetting()},
	}

	err = servicea.SendUnary(cli)
	test.Assert(t, err == nil, err)

	err = servicea.SendClientStreaming(cli, context.Background(), inj)
	test.Assert(t, err == nil, err)

	err = servicea.SendServerStreaming(cli, context.Background(), inj)
	test.Assert(t, err == nil, err)

	err = servicea.SendBidiStreaming(cli, context.Background(), inj)
	test.Assert(t, err == nil, err)
}

func TestError(t *testing.T) {
	clientStreaming(t)
	serverStreaming(t)
	bidiStreaming(t)
}

func TestClientStreaming(t *testing.T) {
	clientStreaming(t)
}

func clientStreaming(t *testing.T) {
	cli, err := servicea.InitServiceAClient(common2.NewWrapDialer())
	test.Assert(t, err == nil, err)

	// 1. close conn
	closeConnTest := func() {
		t.Logf("ClientStreaming[self close connertion] start")
		ctx := context.Background()
		inj := common2.Injector{
			Settings: []common2.Setting{common2.NewCloseConnSetting()},
		}
		err = servicea.SendClientStreaming(cli, ctx, inj)
		time.Sleep(time.Second)
		t.Logf("ClientStreaming[self close connertion] error=%v\n", err)
	}
	closeConnTest()

	// 2. cancel
	cancelTest := func() {
		t.Logf("ClientStreaming[self context cancel] start")
		ctx, cancel := context.WithCancel(context.Background())
		inj := common2.Injector{
			Settings: []common2.Setting{common2.NewSelfCancelSetting(cancel)},
		}
		err = servicea.SendClientStreaming(cli, ctx, inj)
		time.Sleep(time.Second)
		t.Logf("ClientStreaming[self context cancel] error=%v\n", err)
	}
	cancelTest()

	// 3. timeout
	timeoutTest := func() {
		t.Logf("ClientStreaming[context timeout] start")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)
		err = servicea.SendClientStreaming(cli, ctx, common2.NewDummyInjector())
		time.Sleep(time.Second)
		t.Logf("ClientStreaming[context timeout] error=%v\n", err)
	}
	timeoutTest()
}

func serverStreaming(t *testing.T) {
	cli, err := servicea.InitServiceAClient(common2.NewWrapDialer())
	test.Assert(t, err == nil, err)

	// 1. close conn
	closeConnTest := func() {
		t.Logf("ServerStreaming[self close connertion] start")
		ctx := context.Background()
		inj := common2.Injector{
			Settings: []common2.Setting{common2.NewCloseConnSetting()},
		}
		err = servicea.SendServerStreaming(cli, ctx, inj)
		t.Logf("ServerStreaming[self close connertion] error=%v\n", err)
	}
	closeConnTest()

	// 2. cancel
	cancelTest := func() {
		t.Logf("ServerStreaming[self context cancel] start")
		ctx, cancel := context.WithCancel(context.Background())
		inj := common2.Injector{
			Settings: []common2.Setting{common2.NewSelfCancelSetting(cancel)},
		}
		err = servicea.SendServerStreaming(cli, ctx, inj)
		time.Sleep(time.Second)
		t.Logf("ServerStreaming[self context cancel] error=%v\n", err)
	}
	cancelTest()

	// 3. timeout
	timeoutTest := func() {
		t.Logf("ServerStreaming[context timeout] start")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)
		err = servicea.SendServerStreaming(cli, ctx, common2.NewDummyInjector())
		time.Sleep(time.Second)
		t.Logf("ServerStreaming[context timeout] error=%v\n", err)
	}
	timeoutTest()
}

func bidiStreaming(t *testing.T) {
	cli, err := servicea.InitServiceAClient(common2.NewWrapDialer())
	test.Assert(t, err == nil, err)

	// 1. close conn
	closeConnTest := func() {
		t.Log("Bidistreaming[self close connertion] start")
		ctx := context.Background()
		inj := common2.Injector{
			Settings: []common2.Setting{common2.NewCloseConnSetting()},
		}
		err = servicea.SendBidiStreaming(cli, ctx, inj)
		t.Logf("Bidistreaming[self close connertion] error=%v\n", err)
	}
	closeConnTest()

	// 2. cancel
	cancelTest := func() {
		t.Log("Bidistreaming[self context cancel] start")
		ctx, cancel := context.WithCancel(context.Background())
		inj := common2.Injector{
			Settings: []common2.Setting{common2.NewSelfCancelSetting(cancel)},
		}
		err = servicea.SendBidiStreaming(cli, ctx, inj)
		time.Sleep(time.Second)
		t.Logf("Bidistreaming[self context cancel] error=%v\n", err)
	}
	cancelTest()

	// 3. timeout
	timeoutTest := func() {
		t.Log("Bidistreaming[context timeout] start")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)
		err = servicea.SendBidiStreaming(cli, ctx, common2.NewDummyInjector())
		t.Logf("Bidistreaming[context timeout] error=%v\n", err)
	}
	timeoutTest()
}
