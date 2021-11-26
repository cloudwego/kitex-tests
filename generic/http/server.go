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

package http

import (
	"fmt"
	"net"
	"reflect"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/http/bizservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/server"
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
	svc := bizservice.NewServer(new(BizServiceImpl), server.WithServiceAddr(addr))
	go func() {
		if err := svc.Run(); err != nil {
			println(err)
		}
	}()
	return svc
}

func newGenericClient(serviceName string, g generic.Generic, targetIPPort string) genericclient.Client {
	var opts []client.Option
	opts = append(opts, client.WithHostPorts(targetIPPort))
	genericCli, err := genericclient.NewClient(serviceName, g, opts...)
	if err != nil {
		panic(err)
	}
	return genericCli
}
