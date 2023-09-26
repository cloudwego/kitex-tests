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

namespace go codegen.test.simple

struct SimpleReq {
    1: required i64 id,
    2: required string type,
    3: string newA
    4: string NewB
    5: string myResult
    6: string hisresult
    7: string your_result
    8: string myArgs
    9: string hisargs
    10: string your_args
    11: string NewResult
    12: string newAResult
    13: string newb_result
    14: string newresult
    15: string new_result
    16: string NewRaesult
    17: string new_raesult
}

struct SimpleResp {
    1: required i64 id,
    2: required string ret,
}

service SimpleService {

    SimpleResp method1(1: SimpleReq req),

    SimpleResp method2(1: SimpleReq req,2: string extra),

    SimpleResp method3(1: SimpleReq req),

    SimpleResp method4(1: string extra),

    SimpleResp method5(),

    oneway void method6(1: SimpleReq req),

    oneway void method7(),

}