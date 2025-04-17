// Code generated by Kitex v0.12.3. DO NOT EDIT.

package serviceb

import (
	"context"
	"errors"
	multi_service "github.com/cloudwego/kitex-tests/kitex_gen/thrift/multi_service"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{
	"Echo2": kitex.NewMethodInfo(
		echo2Handler,
		newServiceBEcho2Args,
		newServiceBEcho2Result,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
}

var (
	serviceBServiceInfo                = NewServiceInfo()
	serviceBServiceInfoForClient       = NewServiceInfoForClient()
	serviceBServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return serviceBServiceInfo
}

// for stream client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return serviceBServiceInfoForStreamClient
}

// for client
func serviceInfoForClient() *kitex.ServiceInfo {
	return serviceBServiceInfoForClient
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
	serviceName := "ServiceB"
	handlerType := (*multi_service.ServiceB)(nil)
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
		"PackageName": "multi_service",
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

func echo2Handler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*multi_service.ServiceBEcho2Args)
	realResult := result.(*multi_service.ServiceBEcho2Result)
	success, err := handler.(multi_service.ServiceB).Echo2(ctx, realArg.Req)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}
func newServiceBEcho2Args() interface{} {
	return multi_service.NewServiceBEcho2Args()
}

func newServiceBEcho2Result() interface{} {
	return multi_service.NewServiceBEcho2Result()
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) Echo2(ctx context.Context, req *multi_service.Request) (r *multi_service.Response, err error) {
	var _args multi_service.ServiceBEcho2Args
	_args.Req = req
	var _result multi_service.ServiceBEcho2Result
	if err = p.c.Call(ctx, "Echo2", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}
