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
    3: string write,
    4: optional string alias,
    5: required string myIdCard,
    6: required string ApiType,
    7: required string By2ApiType,
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