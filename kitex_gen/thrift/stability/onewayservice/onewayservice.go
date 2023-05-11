// Code generated by Kitex v0.5.2. DO NOT EDIT.

package onewayservice

import (
	"context"
	stability "github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
)

func serviceInfo() *kitex.ServiceInfo {
	return onewayServiceServiceInfo
}

var onewayServiceServiceInfo = NewServiceInfo()

func NewServiceInfo() *kitex.ServiceInfo {
	serviceName := "OnewayService"
	handlerType := (*stability.OnewayService)(nil)
	methods := map[string]kitex.MethodInfo{
		"VisitOneway": kitex.NewMethodInfo(visitOnewayHandler, newOnewayServiceVisitOnewayArgs, nil, true),
	}
	extra := map[string]interface{}{
		"PackageName": "stability",
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         methods,
		PayloadCodec:    kitex.Thrift,
		KiteXGenVersion: "v0.5.2",
		Extra:           extra,
	}
	return svcInfo
}

func visitOnewayHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*stability.OnewayServiceVisitOnewayArgs)

	err := handler.(stability.OnewayService).VisitOneway(ctx, realArg.Req)
	if err != nil {
		return err
	}

	return nil
}
func newOnewayServiceVisitOnewayArgs() interface{} {
	return stability.NewOnewayServiceVisitOnewayArgs()
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) VisitOneway(ctx context.Context, req *stability.STRequest) (err error) {
	var _args stability.OnewayServiceVisitOnewayArgs
	_args.Req = req
	if err = p.c.Call(ctx, "VisitOneway", &_args, nil); err != nil {
		return
	}
	return nil
}
