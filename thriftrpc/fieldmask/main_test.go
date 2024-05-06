// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fieldmask

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	fieldmask0 "github.com/cloudwego/kitex-tests/kitex_gen/fieldmask"
	"github.com/cloudwego/kitex-tests/kitex_gen/fieldmask/bizservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/thriftgo/fieldmask"
)

func TestMain(m *testing.M) {
	addr, err := net.ResolveTCPAddr("tcp", ":8999")
	if err != nil {
		panic(err.Error())
	}
	svr := bizservice.NewServer(new(BizServiceImpl), server.WithServiceAddr(addr))

	go func() {
		err = svr.Run()
		if err != nil {
			panic(err.Error())
		}
	}()

	// initialize request and response fieldmasks and cache them
	respMask, err := fieldmask.NewFieldMask((*fieldmask0.BizResponse)(nil).GetTypeDescriptor(), "$.A")
	if err != nil {
		panic(err)
	}
	fmCache.Store("BizResponse", respMask)
	reqMask, err := fieldmask.NewFieldMask((*fieldmask0.BizRequest)(nil).GetTypeDescriptor(), "$.B", "$.RespMask")
	if err != nil {
		panic(err)
	}
	fmCache.Store("BizRequest", reqMask)

	// black list mod
	respMaskBlack, err := fieldmask.Options{BlackListMode: true}.NewFieldMask((*fieldmask0.BizResponse)(nil).GetTypeDescriptor(), "$.A", "$.B")
	if err != nil {
		panic(err)
	}
	fmCache.Store("BizResponse-Black", respMaskBlack)
	reqMaskBlack, err := fieldmask.Options{BlackListMode: true}.NewFieldMask((*fieldmask0.BizRequest)(nil).GetTypeDescriptor(), "$.A")
	if err != nil {
		panic(err)
	}
	fmCache.Store("BizRequest-Black", reqMaskBlack)

	time.Sleep(time.Second)
	m.Run()
	svr.Stop()
}

var fmCache sync.Map

func TestFieldMask(t *testing.T) {
	cli, err := bizservice.NewClient("BizService", client.WithHostPorts(":8999"))
	if err != nil {
		t.Fatal(err)
	}

	req := fieldmask0.NewBizRequest()
	req.A = "A"
	req.B = "B"
	// try load request's fieldmask
	reqMask, ok := fmCache.Load("BizRequest")
	if ok {
		req.Set_FieldMask(reqMask.(*fieldmask.FieldMask))
	}

	// try get response's fieldmask
	respMask, ok := fmCache.Load("BizResponse")
	if ok {
		// serialize the respMask
		fm, err := fieldmask.Marshal(respMask.(*fieldmask.FieldMask))
		if err != nil {
			t.Fatal(err)
		}
		// let request carry fm
		req.RespMask = fm
	}

	resp, err := cli.BizMethod1(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", resp)

	if resp.A == "" { // resp.A in mask
		t.Fail()
	}
	if resp.B == "" { // resp.B not in mask, but it's required, so still written
		t.Fail()
	}
	if resp.C != "" { // resp.C not in mask
		t.Fail()
	}
}

func TestFieldMask_BlackList(t *testing.T) {
	cli, err := bizservice.NewClient("BizService", client.WithHostPorts(":8999"))
	if err != nil {
		t.Fatal(err)
	}

	req := fieldmask0.NewBizRequest()
	req.A = "A"
	req.B = "B"
	// try load request's fieldmask
	reqMask, ok := fmCache.Load("BizRequest-Black")
	if ok {
		req.Set_FieldMask(reqMask.(*fieldmask.FieldMask))
	}

	// try get reponse's fieldmask
	respMask, ok := fmCache.Load("BizResponse-Black")
	if ok {
		// serialize the respMask
		fm, err := fieldmask.Marshal(respMask.(*fieldmask.FieldMask))
		if err != nil {
			t.Fatal(err)
		}
		// let request carry fm
		req.RespMask = fm
	}

	resp, err := cli.BizMethod1(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", resp)

	if resp.A != "" { // resp.A in mask
		t.Fail()
	}
	if resp.B == "" { // resp.B not in mask, but it's required, so still written
		t.Fail()
	}
	if resp.C == "" { // resp.C not in mask
		t.Fail()
	}
}
