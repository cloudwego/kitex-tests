// Copyright 2024 CloudWeGo Authors
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

package fieldmask

import (
	"context"
	"errors"

	fieldmask0 "github.com/cloudwego/kitex-tests/kitex_gen/fieldmask"
	"github.com/cloudwego/thriftgo/fieldmask"
)

// BizServiceImpl implements the last service interface defined in the IDL.
type BizServiceImpl struct{}

// BizMethod1 implements the BizServiceImpl interface.
func (s *BizServiceImpl) BizMethod1(ctx context.Context, req *fieldmask0.BizRequest) (resp *fieldmask0.BizResponse, err error) {
	// check if request has been masked
	if req.A != "" { // req.A must be filtered
		return nil, errors.New("request must filter BizRequest.A!")
	}
	if req.B == "" { // req.B must not be filtered
		return nil, errors.New("request must not filter BizRequest.B!")
	}

	resp = fieldmask0.NewBizResponse()

	// check if request carries a fieldmask
	if req.RespMask != nil {
		println("got fieldmask", string(req.RespMask))
		fm, err := fieldmask.Unmarshal(req.RespMask)
		if err != nil {
			return nil, err
		}
		// set fieldmask for response
		resp.Set_FieldMask(fm)
	}

	resp.A = "A"
	resp.B = "B"
	resp.C = "C"
	return
}
