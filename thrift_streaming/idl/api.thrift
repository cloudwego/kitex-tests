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

include "a.b.c.thrift"

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

// for checking whether the generated code is ok
service PingPongOnlyService {
    EchoResponse EchoPingPongNew (1: EchoRequest req1),
}

// Also streaming
service PingPongOnlyServiceChild extends PingPongOnlyService {
    EchoResponse EchoBidirectionalExtended (1: EchoRequest req1) (streaming.mode="bidirectional"),
}

// Also streaming
service PingPongOnlyServiceChildChild extends PingPongOnlyServiceChild {}

// for checking whether the generated code is ok
service StreamOnlyService {
    EchoResponse EchoBidirectionalNew (1: EchoRequest req1) (streaming.mode="bidirectional"),
}

// for checking services extending a service extending a service
service StreamOnlyServiceChild extends StreamOnlyService {}

// should also be a streaming service
service StreamOnlyServiceChildChild extends StreamOnlyServiceChild {}

service ABCService {
    a.b.c.Response Echo(1: a.b.c.Request req1, 2: a.b.c.Request req2),
    a.b.c.Response EchoBidirectional(1: a.b.c.Request req1) (streaming.mode="bidirectional"),
    a.b.c.Response EchoServer(1: a.b.c.Request req1) (streaming.mode="server"),
    a.b.c.Response EchoClient(1: a.b.c.Request req1) (streaming.mode="client"),
    a.b.c.Response EchoUnary(1: a.b.c.Request req1) (streaming.mode="unary"),
}