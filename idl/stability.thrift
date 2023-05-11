/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

namespace go thrift.stability

include "instparam.thrift"
include "base.thrift"

enum TestEnum {
    FIRST = 1,
    SECOND = 2,
    THIRD = 3,
    FOURTH = 4,
}

exception STException {
    1:string message;
}

struct STRequest {
    1: required string Name,
    2: optional bool on,
    3: byte b,
    4: optional i16 int16 = 42,
    5: i32 int32,
    6: i64 int64,
    7: double d,
    8: string str,
    9: binary bin,
    10: map<string, string> stringMap,
    11: list<string> stringList,
    12: set<string> stringSet,
    13: TestEnum e,
    14: required string flagMsg
    15: required string framework = "kitex",
    16: optional string mockCost,
    17: map<i32, instparam.SubMessage> subMsgs,

    255: optional base.Base Base,
}

struct STResponse {
    1: string str,
    2: map<string, string> mp,
    3: required string flagMsg
    255: optional base.BaseResp BaseResp,
}

service OnewayService {
    oneway void VisitOneway(1: STRequest req)
}

service STService extends OnewayService {
    STResponse testSTReq(1: STRequest req)
    instparam.ObjResp testObjReq(1: instparam.ObjReq req)
    STResponse testException(1: STRequest req) throws (1:STException stException);
    STResponse circuitBreakTest(1: STRequest req)

}
