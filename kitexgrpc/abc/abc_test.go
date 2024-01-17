package abc

import (
	"testing"

	"github.com/cloudwego/kitex-tests/common"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/consts"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/servicea"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/serviceb"
	"github.com/cloudwego/kitex-tests/kitexgrpc/abc/servicec"
	"github.com/cloudwego/kitex-tests/pkg/test"
)

func TestMain(m *testing.M) {
	svrC := servicec.InitServiceCServer()
	go func() {
		defer svrC.Stop()
		err := svrC.Run()
		if err != nil {
			panic(err)
		}
	}()
	svrB := serviceb.InitServiceBServer()
	go func() {
		defer svrB.Stop()
		err := svrB.Run()
		if err != nil {
			panic(err)
		}
	}()

	common.WaitServer(consts.ServiceCAddr)
	common.WaitServer(consts.ServiceBAddr)

	m.Run()
}

func TestABC(t *testing.T) {
	cli, err := servicea.InitServiceAClient()
	test.Assert(t, err == nil, err)

	err = servicea.SendUnary(cli)
	test.Assert(t, err == nil, err)

	err = servicea.SendClientStreaming(cli)
	test.Assert(t, err == nil, err)

	err = servicea.SendServerStreaming(cli)
	test.Assert(t, err == nil, err)

	err = servicea.SendBidiStreaming(cli)
	test.Assert(t, err == nil, err)
}
