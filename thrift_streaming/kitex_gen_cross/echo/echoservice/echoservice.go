// Code generated by Kitex v0.8.0. DO NOT EDIT.

package echoservice

import (
	"context"
	echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_cross/echo"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
)

func serviceInfo() *kitex.ServiceInfo {
	return echoServiceServiceInfo
}

var echoServiceServiceInfo = NewServiceInfo()

func NewServiceInfo() *kitex.ServiceInfo {
	serviceName := "EchoService"
	handlerType := (*echo.EchoService)(nil)
	methods := map[string]kitex.MethodInfo{
		"EchoPingPong": kitex.NewMethodInfo(echoPingPongHandler, newEchoServiceEchoPingPongArgs, newEchoServiceEchoPingPongResult, false),
		"EchoOneway":   kitex.NewMethodInfo(echoOnewayHandler, newEchoServiceEchoOnewayArgs, newEchoServiceEchoOnewayResult, false),
		"Ping":         kitex.NewMethodInfo(pingHandler, newEchoServicePingArgs, newEchoServicePingResult, false),
	}
	extra := map[string]interface{}{
		"PackageName":     "echo",
		"ServiceFilePath": `idl/api.thrift`,
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         methods,
		PayloadCodec:    kitex.Thrift,
		KiteXGenVersion: "v0.8.0",
		Extra:           extra,
	}
	return svcInfo
}

func echoPingPongHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*echo.EchoServiceEchoPingPongArgs)
	realResult := result.(*echo.EchoServiceEchoPingPongResult)
	success, err := handler.(echo.EchoService).EchoPingPong(ctx, realArg.Req1, realArg.Req2)
	if err != nil {
		switch v := err.(type) {
		case *echo.EchoException:
			realResult.E = v
		default:
			return err
		}
	} else {
		realResult.Success = success
	}
	return nil
}
func newEchoServiceEchoPingPongArgs() interface{} {
	return echo.NewEchoServiceEchoPingPongArgs()
}

func newEchoServiceEchoPingPongResult() interface{} {
	return echo.NewEchoServiceEchoPingPongResult()
}

func echoOnewayHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*echo.EchoServiceEchoOnewayArgs)

	err := handler.(echo.EchoService).EchoOneway(ctx, realArg.Req1)
	if err != nil {
		return err
	}

	return nil
}
func newEchoServiceEchoOnewayArgs() interface{} {
	return echo.NewEchoServiceEchoOnewayArgs()
}

func newEchoServiceEchoOnewayResult() interface{} {
	return echo.NewEchoServiceEchoOnewayResult()
}

func pingHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {

	err := handler.(echo.EchoService).Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}
func newEchoServicePingArgs() interface{} {
	return echo.NewEchoServicePingArgs()
}

func newEchoServicePingResult() interface{} {
	return echo.NewEchoServicePingResult()
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) EchoPingPong(ctx context.Context, req1 *echo.EchoRequest, req2 *echo.EchoRequest) (r *echo.EchoResponse, err error) {
	var _args echo.EchoServiceEchoPingPongArgs
	_args.Req1 = req1
	_args.Req2 = req2
	var _result echo.EchoServiceEchoPingPongResult
	if err = p.c.Call(ctx, "EchoPingPong", &_args, &_result); err != nil {
		return
	}
	switch {
	case _result.E != nil:
		return r, _result.E
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) EchoOneway(ctx context.Context, req1 *echo.EchoRequest) (err error) {
	var _args echo.EchoServiceEchoOnewayArgs
	_args.Req1 = req1
	var _result echo.EchoServiceEchoOnewayResult
	if err = p.c.Call(ctx, "EchoOneway", &_args, &_result); err != nil {
		return
	}
	return nil
}

func (p *kClient) Ping(ctx context.Context) (err error) {
	var _args echo.EchoServicePingArgs
	var _result echo.EchoServicePingResult
	if err = p.c.Call(ctx, "Ping", &_args, &_result); err != nil {
		return
	}
	return nil
}
