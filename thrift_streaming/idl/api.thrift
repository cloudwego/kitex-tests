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

namespace go echo

struct EchoRequest {
    1: required string message,
}

struct EchoResponse {
    1: required string message,
}

exception EchoException {
  1: string message
}


service EchoService {
    // streaming api (HTTP2)
    EchoResponse EchoBidirectional (1: EchoRequest req1) (streaming.mode="bidirectional"),
    EchoResponse EchoClient (1: EchoRequest req1) (streaming.mode="client"),
    EchoResponse EchoServer (1: EchoRequest req1) (streaming.mode="server"),
    EchoResponse EchoUnary (1: EchoRequest req1) (streaming.mode="unary"),

    // KitexThrift
    EchoResponse EchoPingPong (1: EchoRequest req1, 2: EchoRequest req2) throws (1: EchoException e),
    void EchoOneway(1: EchoRequest req1),
    void Ping(),
}
