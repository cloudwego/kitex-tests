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

package streamx

import (
	"errors"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/codes"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/status"
)

const (
	ReturnErrModeKey = "return_err_mode"

	ReturnGRPCErr  = "return_grpc_err"
	ReturnBizErr   = "return_biz_err"
	ReturnOtherErr = "return_other_err"
)

func GRPCReturnErr(mode string) error {
	switch mode {
	case ReturnGRPCErr:
		return status.Err(codes.Internal, "grpc")
	case ReturnBizErr:
		return kerrors.NewBizStatusError(10000, "biz")
	case ReturnOtherErr:
		return errors.New("other")
	default:
		return nil
	}
}
