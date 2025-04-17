// Code generated by Kitex v0.12.3. DO NOT EDIT.

package bizservice

import (
	"context"
	"errors"
	http "github.com/cloudwego/kitex-tests/kitex_gen/thrift/http"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{
	"BizMethod1": kitex.NewMethodInfo(
		bizMethod1Handler,
		newBizServiceBizMethod1Args,
		newBizServiceBizMethod1Result,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
	"BizMethod2": kitex.NewMethodInfo(
		bizMethod2Handler,
		newBizServiceBizMethod2Args,
		newBizServiceBizMethod2Result,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
	"BizMethod3": kitex.NewMethodInfo(
		bizMethod3Handler,
		newBizServiceBizMethod3Args,
		newBizServiceBizMethod3Result,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
}

var (
	bizServiceServiceInfo                = NewServiceInfo()
	bizServiceServiceInfoForClient       = NewServiceInfoForClient()
	bizServiceServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return bizServiceServiceInfo
}

// for stream client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return bizServiceServiceInfoForStreamClient
}

// for client
func serviceInfoForClient() *kitex.ServiceInfo {
	return bizServiceServiceInfoForClient
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
	serviceName := "BizService"
	handlerType := (*http.BizService)(nil)
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
		"PackageName": "http",
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

func bizMethod1Handler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*http.BizServiceBizMethod1Args)
	realResult := result.(*http.BizServiceBizMethod1Result)
	success, err := handler.(http.BizService).BizMethod1(ctx, realArg.Req)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}
func newBizServiceBizMethod1Args() interface{} {
	return http.NewBizServiceBizMethod1Args()
}

func newBizServiceBizMethod1Result() interface{} {
	return http.NewBizServiceBizMethod1Result()
}

func bizMethod2Handler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*http.BizServiceBizMethod2Args)
	realResult := result.(*http.BizServiceBizMethod2Result)
	success, err := handler.(http.BizService).BizMethod2(ctx, realArg.Req)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}
func newBizServiceBizMethod2Args() interface{} {
	return http.NewBizServiceBizMethod2Args()
}

func newBizServiceBizMethod2Result() interface{} {
	return http.NewBizServiceBizMethod2Result()
}

func bizMethod3Handler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*http.BizServiceBizMethod3Args)
	realResult := result.(*http.BizServiceBizMethod3Result)
	success, err := handler.(http.BizService).BizMethod3(ctx, realArg.Req)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}
func newBizServiceBizMethod3Args() interface{} {
	return http.NewBizServiceBizMethod3Args()
}

func newBizServiceBizMethod3Result() interface{} {
	return http.NewBizServiceBizMethod3Result()
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) BizMethod1(ctx context.Context, req *http.BizRequest) (r *http.BizResponse, err error) {
	var _args http.BizServiceBizMethod1Args
	_args.Req = req
	var _result http.BizServiceBizMethod1Result
	if err = p.c.Call(ctx, "BizMethod1", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) BizMethod2(ctx context.Context, req *http.BizRequest) (r *http.BizResponse, err error) {
	var _args http.BizServiceBizMethod2Args
	_args.Req = req
	var _result http.BizServiceBizMethod2Result
	if err = p.c.Call(ctx, "BizMethod2", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) BizMethod3(ctx context.Context, req *http.BizRequest) (r *http.BizResponse, err error) {
	var _args http.BizServiceBizMethod3Args
	_args.Req = req
	var _result http.BizServiceBizMethod3Result
	if err = p.c.Call(ctx, "BizMethod3", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}
