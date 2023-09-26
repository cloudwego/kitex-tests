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

namespace py base
namespace go base
namespace rs base
namespace java com.thrift.base

struct Conf {
    1: bool Open = false,
    2: string Env = "",
}

struct Base {
    1: string A = "",
    2: string B = "",
    5: optional Conf C,
    6: optional map<string, string> Extra,
}

struct BaseResp {
    1: string D = "",
    2: i32 E = 0,
    3: optional map<string, string> Extra,
}

struct CommonResp{
    1: string hey
    2: BaseResp baseResp
}