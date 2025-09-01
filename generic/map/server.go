// Copyright 2021 CloudWeGo Authors
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
	"time"

	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/server/genericserver"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/tenant/echoservice"
	"github.com/cloudwego/kitex-tests/pkg/utils"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
)

func assert(expected, actual interface{}) error {
	if !reflect.DeepEqual(expected, actual) {
		err := fmt.Errorf("expected: %#v, but get: %#v", expected, actual)
		return err
	}
	return nil
}

func runServer(ln net.Listener) server.Server {
	svc := echoservice.NewServer(new(EchoServiceImpl), server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler))
	go func() {
		if err := svc.Run(); err != nil {
			panic(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	return svc
}

var genericAddress string

func runGenericServer() server.Server {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	if err != nil {
		panic(err)
	}
	g, err := generic.MapThriftGeneric(p)
	if err != nil {
		panic(err)
	}
	ln := serverutils.Listen()
	genericAddress = ln.Addr().String()
	svc := genericserver.NewServer(&GenericServiceImpl{}, g, server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler))
	go func() {
		if err := svc.Run(); err != nil {
			panic(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	return svc
}

var genericV2Address string

func runGenericServerV2() server.Server {
	p, err := generic.NewThriftFileProvider("../../idl/tenant.thrift")
	if err != nil {
		panic(err)
	}
	g, err := generic.MapThriftGeneric(p)
	if err != nil {
		panic(err)
	}
	ln := serverutils.Listen()
	genericV2Address = ln.Addr().String()
	svc := genericserver.NewServerV2(utils.ServiceV2Iface2ServiceV2(&GenericServiceImplV2{}), g, server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler),
		server.WithExitWaitTime(500*time.Millisecond),
	)
	go func() {
		if err := svc.Run(); err != nil {
			panic(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	return svc
}
