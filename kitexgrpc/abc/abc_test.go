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
	"testing"

	"github.com/cloudwego/kitex-tests/common"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/consts"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/servicea"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/serviceb"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/servicec"
	"github.com/cloudwego/kitex-tests/pkg/test"
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

	err = servicea.SendUnary(cli)
	test.Assert(t, err == nil, err)

	err = servicea.SendClientStreaming(cli)
	test.Assert(t, err == nil, err)

	err = servicea.SendServerStreaming(cli)
	test.Assert(t, err == nil, err)

	err = servicea.SendBidiStreaming(cli)
	test.Assert(t, err == nil, err)
}
