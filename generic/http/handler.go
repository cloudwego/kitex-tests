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
	"context"

	"github.com/cloudwego/kitex-tests/pkg/utils"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/http"
)

// BizServiceImpl implements the last service interface defined in the IDL.
type BizServiceImpl struct{}

var resp = &http.BizResponse{
	T: utils.StringPtr("1"),
	RspItems: map[int64]*http.RspItem{
		1: {
			ItemId: utils.Int64Ptr(1),
			Text:   utils.StringPtr("1"),
		},
	},
	VEnum: utils.Int32Ptr(1),
	RspItemList: []*http.RspItem{
		{
			ItemId: utils.Int64Ptr(1),
			Text:   utils.StringPtr("1"),
		},
	},
	HttpCode:  utils.Int32Ptr(1),
	ItemCount: []int64{1, 2, 3},
}

// BizMethod1 implements the BizServiceImpl interface.
func (s *BizServiceImpl) BizMethod1(ctx context.Context, req *http.BizRequest) (r *http.BizResponse, err error) {
	if err := assert(req.GetVInt64(), int64(1)); err != nil {
		return nil, err
	}
	if err := assert(req.GetText(), "text"); err != nil {
		return nil, err
	}
	if err := assert(req.GetToken(), int32(1)); err != nil {
		return nil, err
	}
	if err := assert(req.GetReqItemsMap(), map[int64]*http.ReqItem{1: {
		Id:   utils.Int64Ptr(1),
		Text: utils.StringPtr("text"),
	}}); err != nil {
		return nil, err
	}
	if err := assert(req.GetSome(), &http.ReqItem{
		Id:   utils.Int64Ptr(1),
		Text: utils.StringPtr("text"),
	}); err != nil {
		return nil, err
	}
	if err := assert(req.GetReqItems(), []string{
		"item1", "item2", "item3",
	}); err != nil {
		return nil, err
	}
	if err := assert(req.GetApiVersion(), int32(1)); err != nil {
		return nil, err
	}
	if err := assert(req.GetUid(), int64(1)); err != nil {
		return nil, err
	}
	if err := assert(req.GetCids(), []int64{
		1, 2, 3,
	}); err != nil {
		return nil, err
	}
	if err := assert(req.GetVids(), []string{
		"1", "2", "3",
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

// BizMethod2 implements the BizServiceImpl interface.
func (s *BizServiceImpl) BizMethod2(ctx context.Context, req *http.BizRequest) (r *http.BizResponse, err error) {
	return resp, nil
}

// BizMethod3 implements the BizServiceImpl interface.
func (s *BizServiceImpl) BizMethod3(ctx context.Context, req *http.BizRequest) (r *http.BizResponse, err error) {
	return resp, nil
}
