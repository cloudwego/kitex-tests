// Copyright 2023 CloudWeGo Authors
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

package thrift_streaming

import (
	"context"
	"testing"

	"github.com/cloudwego/kitex/client"

	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb"
	kitexpbservice "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb/pbservice"
)

func TestKitexPBClient(t *testing.T) {
	req := &kitex_pb.Request{
		Message: "hello",
	}

	t.Run("kitex_pb client -> kitex_pb client", func(t *testing.T) {
		cli := kitexpbservice.MustNewClient("test", client.WithHostPorts(pbAddr))
		resp, err := cli.EchoPingPong(context.Background(), req)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "hello")
	})

	t.Run("kitex_pb client -> kitex_pb server", func(t *testing.T) {
		cli := kitexpbservice.MustNewClient("test", client.WithHostPorts(grpcAddr))
		resp, err := cli.EchoPingPong(context.Background(), req)
		test.Assert(t, err == nil, err)
		test.Assert(t, resp.Message == "hello")
	})
}
