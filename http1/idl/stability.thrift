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

enum TestEnum {
    FIRST = 1,
    SECOND = 2,
    THIRD = 3,
    FOURTH = 4,
}

struct STRequest {
    1: optional string Name
    2: optional bool on
    3: optional byte b
    4: optional i16 int16 = 42
    5: optional i32 int32
    6: optional i64 int64
    7: optional double d
    8: optional string str
    9: optional binary bin
    10: optional map<string, string> stringMap
    11: optional list<string> stringList
    12: optional set<string> stringSet
    13: optional TestEnum e
    14: optional string flagMsg
    15: optional string mockCost
    16: optional string framework       (go.tag = "query:\"framework\"")
    17: optional string userId    (go.tag = "header:\"X-User-Id\"")
}

struct STResponse {
    1: optional string str
    2: optional map<string, string> mp
    3: optional string name
    4: optional string framework
}

service STService {
    STResponse testSTReq(1: STRequest req)
}
