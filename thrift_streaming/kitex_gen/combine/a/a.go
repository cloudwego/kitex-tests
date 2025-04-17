// Code generated by Kitex v0.12.3. DO NOT EDIT.

package a

import (
	"context"
	"errors"
	combine "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/combine"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{
	"Foo": kitex.NewMethodInfo(
		fooHandler,
		newAFooArgs,
		newAFooResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
}

var (
	aServiceInfo                = NewServiceInfo()
	aServiceInfoForClient       = NewServiceInfoForClient()
	aServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return aServiceInfo
}

// for stream client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return aServiceInfoForStreamClient
}

// for client
func serviceInfoForClient() *kitex.ServiceInfo {
	return aServiceInfoForClient
}

// NewServiceInfo creates a new ServiceInfo containing all methods
func NewServiceInfo() *kitex.ServiceInfo {
	return newServiceInfo(false, true, true)
}

// NewServiceInfo creates a new ServiceInfo containing non-streaming methods
func NewServiceInfoForClient() *kitex.ServiceInfo {
	return newServiceInfo(false, false, true)
}
func NewServiceInfoForStreamClient() *kitex.ServiceInfo {
	return newServiceInfo(true, true, false)
}

func newServiceInfo(hasStreaming bool, keepStreamingMethods bool, keepNonStreamingMethods bool) *kitex.ServiceInfo {
	serviceName := "A"
	handlerType := (*combine.A)(nil)
	methods := map[string]kitex.MethodInfo{}
	for name, m := range serviceMethods {
		if m.IsStreaming() && !keepStreamingMethods {
			continue
		}
		if !m.IsStreaming() && !keepNonStreamingMethods {
			continue
		}
		methods[name] = m
	}
	extra := map[string]interface{}{
		"PackageName": "combine",
	}
	if hasStreaming {
		extra["streaming"] = hasStreaming
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         methods,
		PayloadCodec:    kitex.Thrift,
		KiteXGenVersion: "v0.12.3",
		Extra:           extra,
	}
	return svcInfo
}

func fooHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*combine.AFooArgs)
	realResult := result.(*combine.AFooResult)
	success, err := handler.(combine.A).Foo(ctx, realArg.Req)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}
func newAFooArgs() interface{} {
	return combine.NewAFooArgs()
}

func newAFooResult() interface{} {
	return combine.NewAFooResult()
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) Foo(ctx context.Context, req *combine.Req) (r *combine.Rsp, err error) {
	var _args combine.AFooArgs
	_args.Req = req
	var _result combine.AFooResult
	if err = p.c.Call(ctx, "Foo", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}
