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
	"io"

	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/combine"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/combine/combineservice"
)

var _ combineservice.CombineService = new(CombineServiceImpl)

type CombineServiceImpl struct{}

func (c CombineServiceImpl) Foo(ctx context.Context, req *combine.Req) (rsp *combine.Rsp, err error) {
	rsp = &combine.Rsp{Message: req.Message}
	return
}

func (c CombineServiceImpl) Bar(stream combine.B_BarServer) (err error) {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		err = stream.Send(&combine.Rsp{Message: req.Message})
		if err != nil {
			return err
		}
	}
}
