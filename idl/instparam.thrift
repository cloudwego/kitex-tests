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

namespace go thrift.instparam

struct SubMessage {
    1: optional i64 id;
    2: optional string value;
}

struct Message {
    1: optional i64 id;
    2: optional string value;
    3: optional list<SubMessage> subMessages;
}

struct ObjReq {
    1: required Message msg
    2: required map<Message, SubMessage> msgMap
    3: required list<SubMessage> subMsgs
    4: optional set<Message> msgSet
    5: required string flagMsg
    6: optional string mockCost
}

struct ObjResp {
    1: required Message msg
    2: required map<Message, SubMessage> msgMap,
    3: required list<SubMessage> subMsgs
    4: optional set<Message> msgSet
    5: required string flagMsg
}
