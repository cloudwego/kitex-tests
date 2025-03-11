// Copyright 2024 CloudWeGo Authors
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

package utiltest

import (
	"errors"
	"testing"

	"github.com/cloudwego/kitex/pkg/utils"

	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/pkg/test"
)

func TestRPCCodec(t *testing.T) {
	req1 := stability.NewSTRequest()
	req1.Name = "Hello Kitex"
	strMap := make(map[string]string)
	strMap["aa"] = "aa"
	strMap["bb"] = "bb"
	req1.StringMap = strMap

	args1 := stability.NewSTServiceTestSTReqArgs()
	args1.Req = req1

	// encode
	buf, err := thrift.MarshalFastMsg("mockMethod", thrift.CALL, 100, args1)
	test.Assert(t, err == nil, err)

	var argsDecode1 stability.STServiceTestSTReqArgs
	// decode
	method, seqID, err := thrift.UnmarshalFastMsg(buf, &argsDecode1)

	test.Assert(t, err == nil, err)
	test.Assert(t, method == "mockMethod")
	test.Assert(t, seqID == 100)
	test.Assert(t, argsDecode1.Req.Name == req1.Name)
	test.Assert(t, len(argsDecode1.Req.StringMap) == len(req1.StringMap))
	for k := range argsDecode1.Req.StringMap {
		test.Assert(t, argsDecode1.Req.StringMap[k] == req1.StringMap[k])
	}

	// *** reuse ThriftMessageCodec
	req2 := stability.NewSTRequest()
	req2.Name = "Hello Kitex1"
	strMap = make(map[string]string)
	strMap["cc"] = "cc"
	strMap["dd"] = "dd"
	req2.StringMap = strMap
	args2 := stability.NewSTServiceTestSTReqArgs()
	args2.Req = req2
	// encode
	buf, err = thrift.MarshalFastMsg("mockMethod1", thrift.CALL, 101, args2)
	test.Assert(t, err == nil, err)

	// decode
	var argsDecode2 stability.STServiceTestSTReqArgs
	method, seqID, err = thrift.UnmarshalFastMsg(buf, &argsDecode2)

	test.Assert(t, err == nil, err)
	test.Assert(t, method == "mockMethod1")
	test.Assert(t, seqID == 101)
	test.Assert(t, argsDecode2.Req.Name == req2.Name)
	test.Assert(t, len(argsDecode2.Req.StringMap) == len(req2.StringMap))
	for k := range argsDecode2.Req.StringMap {
		test.Assert(t, argsDecode2.Req.StringMap[k] == req2.StringMap[k])
	}
}

func TestSerializer(t *testing.T) {
	req := stability.NewSTRequest()
	req.Name = "Hello Kitex"
	strMap := make(map[string]string)
	strMap["aa"] = "aa"
	strMap["bb"] = "bb"
	req.StringMap = strMap

	args := stability.NewSTServiceTestSTReqArgs()
	args.Req = req

	b := thrift.FastMarshal(args)

	var args2 stability.STServiceTestSTReqArgs
	err := thrift.FastUnmarshal(b, &args2)
	test.Assert(t, err == nil, err)

	test.Assert(t, args2.Req.Name == req.Name)
	test.Assert(t, len(args2.Req.StringMap) == len(req.StringMap))
	for k := range args2.Req.StringMap {
		test.Assert(t, args2.Req.StringMap[k] == req.StringMap[k])
	}
}

func TestException(t *testing.T) {
	errMsg := "my error"
	b := utils.MarshalError("some method", errors.New(errMsg))
	err := utils.UnmarshalError(b)
	test.Assert(t, err.Error() == errMsg, err)
}
