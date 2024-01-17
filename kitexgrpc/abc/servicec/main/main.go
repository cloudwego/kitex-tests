package main

import "github.com/cloudwego/kitex-tests/kitexgrpc/abc/servicec"

func main() {
	svr := servicec.InitServiceCServer()
	defer svr.Stop()
	err := svr.Run()
	panic(err)
}
