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
    1: string OUrl
    2: string MyUrl
    3: string MYId
    12:string AHeyIdCard
    13:string BHeyIdcard
    14:string CHeyId
    15:string DHeYId
    16:string EHeyIdCardId
    17:string FHeyIdcardId
    18:string GHeyIdcardIdC
    19:string HHeyIdcardIdc
    10:string IHey2IdCardIdc
    21: string CurrentUuid
    22: string MyUUid
    23: string HisUUId
    24: string BIGUuid
    25: string LARGEUUid
    26: string MY2UUId
    27: string TheUUUid
    28: string MyHtml
    29: string HistHTml
    30: string YourHtMl
    31: string My2Html
    32: string MyUri
}

service SimpleService {

    string method1(1: SimpleReq req),

}