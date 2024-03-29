// Code generated by Kitex v0.8.0. DO NOT EDIT.

package pingpongonlyservicechild

import (
	"context"
	"errors"
	"fmt"
	echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen_slim/echo"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{
	"EchoPingPongNew": kitex.NewMethodInfo(
		echoPingPongNewHandler,
		newPingPongOnlyServiceEchoPingPongNewArgs,
		newPingPongOnlyServiceEchoPingPongNewResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
	"EchoBidirectionalExtended": kitex.NewMethodInfo(
		echoBidirectionalExtendedHandler,
		newPingPongOnlyServiceChildEchoBidirectionalExtendedArgs,
		newPingPongOnlyServiceChildEchoBidirectionalExtendedResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingBidirectional),
	),
}

var (
	pingPongOnlyServiceChildServiceInfo                = NewServiceInfo()
	pingPongOnlyServiceChildServiceInfoForClient       = NewServiceInfoForClient()
	pingPongOnlyServiceChildServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return pingPongOnlyServiceChildServiceInfo
}

// for client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return pingPongOnlyServiceChildServiceInfoForStreamClient
}

// for stream client
func serviceInfoForClient() *kitex.ServiceInfo {
	return pingPongOnlyServiceChildServiceInfoForClient
}

// NewServiceInfo creates a new ServiceInfo containing all methods
func NewServiceInfo() *kitex.ServiceInfo {
	return newServiceInfo(true, true, true)
}

// NewServiceInfo creates a new ServiceInfo containing non-streaming methods
func NewServiceInfoForClient() *kitex.ServiceInfo {
	return newServiceInfo(false, false, true)
}
func NewServiceInfoForStreamClient() *kitex.ServiceInfo {
	return newServiceInfo(true, true, false)
}

func newServiceInfo(hasStreaming bool, keepStreamingMethods bool, keepNonStreamingMethods bool) *kitex.ServiceInfo {
	serviceName := "PingPongOnlyServiceChild"
	handlerType := (*echo.PingPongOnlyServiceChild)(nil)
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
		"PackageName": "echo",
	}
	if hasStreaming {
		extra["streaming"] = hasStreaming
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

func echoPingPongNewHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*echo.PingPongOnlyServiceEchoPingPongNewArgs)
	realResult := result.(*echo.PingPongOnlyServiceEchoPingPongNewResult)
	success, err := handler.(echo.PingPongOnlyService).EchoPingPongNew(ctx, realArg.Req1)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}
func newPingPongOnlyServiceEchoPingPongNewArgs() interface{} {
	return echo.NewPingPongOnlyServiceEchoPingPongNewArgs()
}

func newPingPongOnlyServiceEchoPingPongNewResult() interface{} {
	return echo.NewPingPongOnlyServiceEchoPingPongNewResult()
}

func echoBidirectionalExtendedHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, ok := arg.(*streaming.Args)
	if !ok {
		return errors.New("PingPongOnlyServiceChild.EchoBidirectionalExtended is a thrift streaming method, please call with Kitex StreamClient")
	}
	stream := &pingPongOnlyServiceChildEchoBidirectionalExtendedServer{st.Stream}
	return handler.(echo.PingPongOnlyServiceChild).EchoBidirectionalExtended(stream)
}

type pingPongOnlyServiceChildEchoBidirectionalExtendedClient struct {
	streaming.Stream
}

func (x *pingPongOnlyServiceChildEchoBidirectionalExtendedClient) Send(m *echo.EchoRequest) error {
	return x.Stream.SendMsg(m)
}
func (x *pingPongOnlyServiceChildEchoBidirectionalExtendedClient) Recv() (*echo.EchoResponse, error) {
	m := new(echo.EchoResponse)
	return m, x.Stream.RecvMsg(m)
}

type pingPongOnlyServiceChildEchoBidirectionalExtendedServer struct {
	streaming.Stream
}

func (x *pingPongOnlyServiceChildEchoBidirectionalExtendedServer) Send(m *echo.EchoResponse) error {
	return x.Stream.SendMsg(m)
}

func (x *pingPongOnlyServiceChildEchoBidirectionalExtendedServer) Recv() (*echo.EchoRequest, error) {
	m := new(echo.EchoRequest)
	return m, x.Stream.RecvMsg(m)
}

func newPingPongOnlyServiceChildEchoBidirectionalExtendedArgs() interface{} {
	return echo.NewPingPongOnlyServiceChildEchoBidirectionalExtendedArgs()
}

func newPingPongOnlyServiceChildEchoBidirectionalExtendedResult() interface{} {
	return echo.NewPingPongOnlyServiceChildEchoBidirectionalExtendedResult()
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) EchoPingPongNew(ctx context.Context, req1 *echo.EchoRequest) (r *echo.EchoResponse, err error) {
	var _args echo.PingPongOnlyServiceEchoPingPongNewArgs
	_args.Req1 = req1
	var _result echo.PingPongOnlyServiceEchoPingPongNewResult
	if err = p.c.Call(ctx, "EchoPingPongNew", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) EchoBidirectionalExtended(ctx context.Context) (PingPongOnlyServiceChild_EchoBidirectionalExtendedClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "EchoBidirectionalExtended", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &pingPongOnlyServiceChildEchoBidirectionalExtendedClient{res.Stream}
	return stream, nil
}
