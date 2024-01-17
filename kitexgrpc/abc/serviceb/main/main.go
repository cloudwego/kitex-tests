package main

import "github.com/cloudwego/kitex-tests/kitexgrpc/abc/serviceb"

func main() {
	svr := serviceb.InitServiceBServer()
	defer svr.Stop()
	err := svr.Run()
	panic(err)
}
