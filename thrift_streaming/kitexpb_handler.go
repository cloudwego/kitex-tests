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

	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/kitex_pb"
)

type KitexPBServiceImpl struct{}

func (k KitexPBServiceImpl) EchoPingPong(ctx context.Context, req *kitex_pb.Request) (resp *kitex_pb.Response, err error) {
	resp = &kitex_pb.Response{
		Message: req.Message,
	}
	return
}
