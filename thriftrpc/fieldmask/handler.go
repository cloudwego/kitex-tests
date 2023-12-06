package main

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
	// check if reques has been masked
	if req.A != "" { // req.A not in mask
		return nil, errors.New("request must mask BizRequest.A!")
	}
	if req.B == "" { // req.B in mask
		return nil, errors.New("request must not mask BizRequest.B!")
	}

	resp = fieldmask0.NewBizResponse()

	// check if request carries a fieldmask
	if req.RespMask != nil {
		println("got fm", string(req.RespMask))
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
