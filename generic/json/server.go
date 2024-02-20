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

package tests

import (
	"fmt"
	"net"
	"reflect"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/echoservice"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/server/genericserver"
)

func assert(expected, actual interface{}) error {
	if !reflect.DeepEqual(expected, actual) {
		err := fmt.Errorf("expected: %#v, but get: %#v", expected, actual)
		return err
	}
	return nil
}

const address = ":9009"

func runServer() server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", address)
	svc := echoservice.NewServer(new(EchoServiceImpl), server.WithServiceAddr(addr))
	go func() {
		if err := svc.Run(); err != nil {
			panic(err)
		}
	}()
	return svc
}

const genericAddress = ":9010"

func runGenericServer() server.Server {
	addr, _ := net.ResolveTCPAddr("tcp", genericAddress)
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	if err != nil {
		panic(err)
	}
	g, err := generic.JSONThriftGeneric(p)
	if err != nil {
		panic(err)
	}
	svc := genericserver.NewServer(&GenericServiceImpl{}, g, server.WithServiceAddr(addr), server.WithMetaHandler(transmeta.ServerTTHeaderHandler))
	go func() {
		if err := svc.Run(); err != nil {
			panic(err)
		}
	}()
	return svc
}
