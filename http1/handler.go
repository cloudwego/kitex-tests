package http1

import (
	"context"
	"math"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/pkg/kerrors"

	"github.com/cloudwego/kitex-tests/http1/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/pkg/utils"
)

type STServiceHandler struct{}

func (*STServiceHandler) TestSTReq(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	if req.Str != nil && *req.Str == "biz_err" {
		return nil, kerrors.NewBizStatusError(10001, "biz error")
	}
	resp := stability.NewSTResponse()
	resp.Str = req.Str
	resp.Mp = req.StringMap
	resp.Name = req.UserId
	resp.Framework = req.Framework
	if req.MockCost != nil {
		if mockSleep, err := time.ParseDuration(*req.MockCost); err != nil {
			return nil, err
		} else {
			time.Sleep(mockSleep)
		}
	}
	return resp, nil
}

func CreateSTRequest(ctx context.Context) (context.Context, *stability.STRequest) {
	req := stability.NewSTRequest()
	name := "byted"
	req.Name = &name
	req.On = utils.BoolPtr(true)
	b := int8(10)
	req.B = &b
	req.Int16 = int16(10)
	i32 := int32(math.MaxInt32)
	req.Int32 = &i32
	i64 := int64(math.MaxInt64)
	req.Int64 = &i64
	d := 0.0
	req.D = &d
	str := utils.RandomString(100)
	req.Str = &str
	req.Bin = []byte{1, 'a', '*'}
	req.StringMap = map[string]string{
		"key1": utils.RandomString(100),
		"key2": utils.RandomString(10),
	}
	req.StringList = []string{
		utils.RandomString(10),
		utils.RandomString(20),
		utils.RandomString(30),
	}
	req.StringSet = []string{
		utils.RandomString(10),
		utils.RandomString(100),
	}
	e := stability.TestEnum_FIRST
	req.E = &e

	ctx = metainfo.WithValue(ctx, "TK", "TV")
	ctx = metainfo.WithPersistentValue(ctx, "PK", "PV")
	return ctx, req
}
