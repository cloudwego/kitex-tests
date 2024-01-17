package main

import "github.com/cloudwego/kitex-tests/kitexgrpc/abc/servicea"

func main() {
	cli, err := servicea.InitServiceAClient()
	if err != nil {
		panic(err)
	}

	servicea.SendUnary(cli)
	servicea.SendClientStreaming(cli)
	servicea.SendServerStreaming(cli)
	servicea.SendBidiStreaming(cli)
}
