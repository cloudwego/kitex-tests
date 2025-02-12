namespace go echo

struct EchoClientRequest {
    1: string Message,
}

struct EchoClientResponse {
    1: string Message,
}

service TestService {
    EchoClientResponse PingPong(1: EchoClientRequest req)
    EchoClientResponse Unary(1: EchoClientRequest req) (streaming.mode="unary"),
    EchoClientResponse EchoBidi (1: EchoClientRequest req) (streaming.mode="bidirectional"),
    EchoClientResponse EchoClient (1: EchoClientRequest req) (streaming.mode="client"),
    EchoClientResponse EchoServer (1: EchoClientRequest req) (streaming.mode="server"),
}

service EmptyService {

}