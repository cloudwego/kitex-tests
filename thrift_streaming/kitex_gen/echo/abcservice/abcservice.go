// Code generated by Kitex v0.9.1. DO NOT EDIT.

package abcservice

import (
	"context"
	"errors"
	"fmt"
	c "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/a/b/c"
	echo "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{
	"Echo": kitex.NewMethodInfo(
		echoHandler,
		newABCServiceEchoArgs,
		newABCServiceEchoResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
	"EchoBidirectional": kitex.NewMethodInfo(
		echoBidirectionalHandler,
		newABCServiceEchoBidirectionalArgs,
		newABCServiceEchoBidirectionalResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingBidirectional),
	),
	"EchoServer": kitex.NewMethodInfo(
		echoServerHandler,
		newABCServiceEchoServerArgs,
		newABCServiceEchoServerResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingServer),
	),
	"EchoClient": kitex.NewMethodInfo(
		echoClientHandler,
		newABCServiceEchoClientArgs,
		newABCServiceEchoClientResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingClient),
	),
	"EchoUnary": kitex.NewMethodInfo(
		echoUnaryHandler,
		newABCServiceEchoUnaryArgs,
		newABCServiceEchoUnaryResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingUnary),
	),
}

var (
	aBCServiceServiceInfo                = NewServiceInfo()
	aBCServiceServiceInfoForClient       = NewServiceInfoForClient()
	aBCServiceServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return aBCServiceServiceInfo
}

// for client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return aBCServiceServiceInfoForStreamClient
}

// for stream client
func serviceInfoForClient() *kitex.ServiceInfo {
	return aBCServiceServiceInfoForClient
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
	serviceName := "ABCService"
	handlerType := (*echo.ABCService)(nil)
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
		KiteXGenVersion: "v0.9.1",
		Extra:           extra,
	}
	return svcInfo
}

func echoHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	realArg := arg.(*echo.ABCServiceEchoArgs)
	realResult := result.(*echo.ABCServiceEchoResult)
	success, err := handler.(echo.ABCService).Echo(ctx, realArg.Req1, realArg.Req2)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}
func newABCServiceEchoArgs() interface{} {
	return echo.NewABCServiceEchoArgs()
}

func newABCServiceEchoResult() interface{} {
	return echo.NewABCServiceEchoResult()
}

func echoBidirectionalHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, ok := arg.(*streaming.Args)
	if !ok {
		return errors.New("ABCService.EchoBidirectional is a thrift streaming method, please call with Kitex StreamClient")
	}
	stream := &aBCServiceEchoBidirectionalServer{st.Stream}
	return handler.(echo.ABCService).EchoBidirectional(stream)
}

type aBCServiceEchoBidirectionalClient struct {
	streaming.Stream
}

func (x *aBCServiceEchoBidirectionalClient) DoFinish(err error) {
	if finisher, ok := x.Stream.(streaming.WithDoFinish); ok {
		finisher.DoFinish(err)
	} else {
		panic(fmt.Sprintf("streaming.WithDoFinish is not implemented by %T", x.Stream))
	}
}
func (x *aBCServiceEchoBidirectionalClient) Send(m *c.Request) error {
	return x.Stream.SendMsg(m)
}
func (x *aBCServiceEchoBidirectionalClient) Recv() (*c.Response, error) {
	m := new(c.Response)
	return m, x.Stream.RecvMsg(m)
}

type aBCServiceEchoBidirectionalServer struct {
	streaming.Stream
}

func (x *aBCServiceEchoBidirectionalServer) Send(m *c.Response) error {
	return x.Stream.SendMsg(m)
}

func (x *aBCServiceEchoBidirectionalServer) Recv() (*c.Request, error) {
	m := new(c.Request)
	return m, x.Stream.RecvMsg(m)
}

func newABCServiceEchoBidirectionalArgs() interface{} {
	return echo.NewABCServiceEchoBidirectionalArgs()
}

func newABCServiceEchoBidirectionalResult() interface{} {
	return echo.NewABCServiceEchoBidirectionalResult()
}

func echoServerHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, ok := arg.(*streaming.Args)
	if !ok {
		return errors.New("ABCService.EchoServer is a thrift streaming method, please call with Kitex StreamClient")
	}
	stream := &aBCServiceEchoServerServer{st.Stream}
	req := new(c.Request)
	if err := st.Stream.RecvMsg(req); err != nil {
		return err
	}
	return handler.(echo.ABCService).EchoServer(req, stream)
}

type aBCServiceEchoServerClient struct {
	streaming.Stream
}

func (x *aBCServiceEchoServerClient) DoFinish(err error) {
	if finisher, ok := x.Stream.(streaming.WithDoFinish); ok {
		finisher.DoFinish(err)
	} else {
		panic(fmt.Sprintf("streaming.WithDoFinish is not implemented by %T", x.Stream))
	}
}
func (x *aBCServiceEchoServerClient) Recv() (*c.Response, error) {
	m := new(c.Response)
	return m, x.Stream.RecvMsg(m)
}

type aBCServiceEchoServerServer struct {
	streaming.Stream
}

func (x *aBCServiceEchoServerServer) Send(m *c.Response) error {
	return x.Stream.SendMsg(m)
}

func newABCServiceEchoServerArgs() interface{} {
	return echo.NewABCServiceEchoServerArgs()
}

func newABCServiceEchoServerResult() interface{} {
	return echo.NewABCServiceEchoServerResult()
}

func echoClientHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, ok := arg.(*streaming.Args)
	if !ok {
		return errors.New("ABCService.EchoClient is a thrift streaming method, please call with Kitex StreamClient")
	}
	stream := &aBCServiceEchoClientServer{st.Stream}
	return handler.(echo.ABCService).EchoClient(stream)
}

type aBCServiceEchoClientClient struct {
	streaming.Stream
}

func (x *aBCServiceEchoClientClient) DoFinish(err error) {
	if finisher, ok := x.Stream.(streaming.WithDoFinish); ok {
		finisher.DoFinish(err)
	} else {
		panic(fmt.Sprintf("streaming.WithDoFinish is not implemented by %T", x.Stream))
	}
}
func (x *aBCServiceEchoClientClient) Send(m *c.Request) error {
	return x.Stream.SendMsg(m)
}
func (x *aBCServiceEchoClientClient) CloseAndRecv() (*c.Response, error) {
	if err := x.Stream.Close(); err != nil {
		return nil, err
	}
	m := new(c.Response)
	return m, x.Stream.RecvMsg(m)
}

type aBCServiceEchoClientServer struct {
	streaming.Stream
}

func (x *aBCServiceEchoClientServer) SendAndClose(m *c.Response) error {
	return x.Stream.SendMsg(m)
}

func (x *aBCServiceEchoClientServer) Recv() (*c.Request, error) {
	m := new(c.Request)
	return m, x.Stream.RecvMsg(m)
}

func newABCServiceEchoClientArgs() interface{} {
	return echo.NewABCServiceEchoClientArgs()
}

func newABCServiceEchoClientResult() interface{} {
	return echo.NewABCServiceEchoClientResult()
}

func echoUnaryHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	if streaming.GetStream(ctx) == nil {
		return errors.New("ABCService.EchoUnary is a thrift streaming unary method, please call with Kitex StreamClient or remove the annotation streaming.mode")
	}
	realArg := arg.(*echo.ABCServiceEchoUnaryArgs)
	realResult := result.(*echo.ABCServiceEchoUnaryResult)
	success, err := handler.(echo.ABCService).EchoUnary(ctx, realArg.Req1)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}
func newABCServiceEchoUnaryArgs() interface{} {
	return echo.NewABCServiceEchoUnaryArgs()
}

func newABCServiceEchoUnaryResult() interface{} {
	return echo.NewABCServiceEchoUnaryResult()
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) Echo(ctx context.Context, req1 *c.Request, req2 *c.Request) (r *c.Response, err error) {
	var _args echo.ABCServiceEchoArgs
	_args.Req1 = req1
	_args.Req2 = req2
	var _result echo.ABCServiceEchoResult
	if err = p.c.Call(ctx, "Echo", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) EchoBidirectional(ctx context.Context) (ABCService_EchoBidirectionalClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "EchoBidirectional", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &aBCServiceEchoBidirectionalClient{res.Stream}
	return stream, nil
}

func (p *kClient) EchoServer(ctx context.Context, req1 *c.Request) (ABCService_EchoServerClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "EchoServer", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &aBCServiceEchoServerClient{res.Stream}

	if err := stream.Stream.SendMsg(req1); err != nil {
		return nil, err
	}
	if err := stream.Stream.Close(); err != nil {
		return nil, err
	}
	return stream, nil
}

func (p *kClient) EchoClient(ctx context.Context) (ABCService_EchoClientClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "EchoClient", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &aBCServiceEchoClientClient{res.Stream}
	return stream, nil
}

func (p *kClient) EchoUnary(ctx context.Context, req1 *c.Request) (r *c.Response, err error) {
	var _args echo.ABCServiceEchoUnaryArgs
	_args.Req1 = req1
	var _result echo.ABCServiceEchoUnaryResult
	if err = p.c.Call(ctx, "EchoUnary", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}
