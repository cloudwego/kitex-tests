// Code generated by Kitex v0.8.0. DO NOT EDIT.

package echoservice

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
	"EchoBidirectional": kitex.NewMethodInfo(
		echoBidirectionalHandler,
		newEchoServiceEchoBidirectionalArgs,
		newEchoServiceEchoBidirectionalResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingBidirectional),
	),
	"EchoClient": kitex.NewMethodInfo(
		echoClientHandler,
		newEchoServiceEchoClientArgs,
		newEchoServiceEchoClientResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingClient),
	),
	"EchoServer": kitex.NewMethodInfo(
		echoServerHandler,
		newEchoServiceEchoServerArgs,
		newEchoServiceEchoServerResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingServer),
	),
	"EchoUnary": kitex.NewMethodInfo(
		echoUnaryHandler,
		newEchoServiceEchoUnaryArgs,
		newEchoServiceEchoUnaryResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingUnary),
	),
	"EchoPingPong": kitex.NewMethodInfo(
		echoPingPongHandler,
		newEchoServiceEchoPingPongArgs,
		newEchoServiceEchoPingPongResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
	"EchoOneway": kitex.NewMethodInfo(
		echoOnewayHandler,
		newEchoServiceEchoOnewayArgs,
		newEchoServiceEchoOnewayResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
	"Ping": kitex.NewMethodInfo(
		pingHandler,
		newEchoServicePingArgs,
		newEchoServicePingResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingNone),
	),
}

var (
	echoServiceServiceInfo                = NewServiceInfo()
	echoServiceServiceInfoForClient       = NewServiceInfoForClient()
	echoServiceServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return echoServiceServiceInfo
}

// for client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return echoServiceServiceInfoForStreamClient
}

// for stream client
func serviceInfoForClient() *kitex.ServiceInfo {
	return echoServiceServiceInfoForClient
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
	serviceName := "EchoService"
	handlerType := (*echo.EchoService)(nil)
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

func echoBidirectionalHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, ok := arg.(*streaming.Args)
	if !ok {
		return errors.New("EchoService.EchoBidirectional is a thrift streaming method, please call with Kitex StreamClient")
	}
	stream := &echoServiceEchoBidirectionalServer{st.Stream}
	return handler.(echo.EchoService).EchoBidirectional(stream)
}

type echoServiceEchoBidirectionalClient struct {
	streaming.Stream
}

func (x *echoServiceEchoBidirectionalClient) Send(m *echo.EchoRequest) error {
	return x.Stream.SendMsg(m)
}
func (x *echoServiceEchoBidirectionalClient) Recv() (*echo.EchoResponse, error) {
	m := new(echo.EchoResponse)
	return m, x.Stream.RecvMsg(m)
}

type echoServiceEchoBidirectionalServer struct {
	streaming.Stream
}

func (x *echoServiceEchoBidirectionalServer) Send(m *echo.EchoResponse) error {
	return x.Stream.SendMsg(m)
}

func (x *echoServiceEchoBidirectionalServer) Recv() (*echo.EchoRequest, error) {
	m := new(echo.EchoRequest)
	return m, x.Stream.RecvMsg(m)
}

func newEchoServiceEchoBidirectionalArgs() interface{} {
	return echo.NewEchoServiceEchoBidirectionalArgs()
}

func newEchoServiceEchoBidirectionalResult() interface{} {
	return echo.NewEchoServiceEchoBidirectionalResult()
}

func echoClientHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, ok := arg.(*streaming.Args)
	if !ok {
		return errors.New("EchoService.EchoClient is a thrift streaming method, please call with Kitex StreamClient")
	}
	stream := &echoServiceEchoClientServer{st.Stream}
	return handler.(echo.EchoService).EchoClient(stream)
}

type echoServiceEchoClientClient struct {
	streaming.Stream
}

func (x *echoServiceEchoClientClient) Send(m *echo.EchoRequest) error {
	return x.Stream.SendMsg(m)
}
func (x *echoServiceEchoClientClient) CloseAndRecv() (*echo.EchoResponse, error) {
	if err := x.Stream.Close(); err != nil {
		return nil, err
	}
	m := new(echo.EchoResponse)
	return m, x.Stream.RecvMsg(m)
}

type echoServiceEchoClientServer struct {
	streaming.Stream
}

func (x *echoServiceEchoClientServer) SendAndClose(m *echo.EchoResponse) error {
	return x.Stream.SendMsg(m)
}

func (x *echoServiceEchoClientServer) Recv() (*echo.EchoRequest, error) {
	m := new(echo.EchoRequest)
	return m, x.Stream.RecvMsg(m)
}

func newEchoServiceEchoClientArgs() interface{} {
	return echo.NewEchoServiceEchoClientArgs()
}

func newEchoServiceEchoClientResult() interface{} {
	return echo.NewEchoServiceEchoClientResult()
}

func echoServerHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, ok := arg.(*streaming.Args)
	if !ok {
		return errors.New("EchoService.EchoServer is a thrift streaming method, please call with Kitex StreamClient")
	}
	stream := &echoServiceEchoServerServer{st.Stream}
	req := new(echo.EchoRequest)
	if err := st.Stream.RecvMsg(req); err != nil {
		return err
	}
	return handler.(echo.EchoService).EchoServer(req, stream)
}

type echoServiceEchoServerClient struct {
	streaming.Stream
}

func (x *echoServiceEchoServerClient) Recv() (*echo.EchoResponse, error) {
	m := new(echo.EchoResponse)
	return m, x.Stream.RecvMsg(m)
}

type echoServiceEchoServerServer struct {
	streaming.Stream
}

func (x *echoServiceEchoServerServer) Send(m *echo.EchoResponse) error {
	return x.Stream.SendMsg(m)
}

func newEchoServiceEchoServerArgs() interface{} {
	return echo.NewEchoServiceEchoServerArgs()
}

func newEchoServiceEchoServerResult() interface{} {
	return echo.NewEchoServiceEchoServerResult()
}

func echoUnaryHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	if streaming.GetStream(ctx) == nil {
		return errors.New("EchoService.EchoUnary is a thrift streaming unary method, please call with Kitex StreamClient or remove the annotation streaming.mode")
	}
	realArg := arg.(*echo.EchoServiceEchoUnaryArgs)
	realResult := result.(*echo.EchoServiceEchoUnaryResult)
	success, err := handler.(echo.EchoService).EchoUnary(ctx, realArg.Req1)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}
func newEchoServiceEchoUnaryArgs() interface{} {
	return echo.NewEchoServiceEchoUnaryArgs()
}

func newEchoServiceEchoUnaryResult() interface{} {
	return echo.NewEchoServiceEchoUnaryResult()
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
	_ = arg.(*echo.EchoServicePingArgs)

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

func (p *kClient) EchoBidirectional(ctx context.Context) (EchoService_EchoBidirectionalClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "EchoBidirectional", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &echoServiceEchoBidirectionalClient{res.Stream}
	return stream, nil
}

func (p *kClient) EchoClient(ctx context.Context) (EchoService_EchoClientClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "EchoClient", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &echoServiceEchoClientClient{res.Stream}
	return stream, nil
}

func (p *kClient) EchoServer(ctx context.Context, req1 *echo.EchoRequest) (EchoService_EchoServerClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "EchoServer", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &echoServiceEchoServerClient{res.Stream}

	if err := stream.Stream.SendMsg(req1); err != nil {
		return nil, err
	}
	if err := stream.Stream.Close(); err != nil {
		return nil, err
	}
	return stream, nil
}

func (p *kClient) EchoUnary(ctx context.Context, req1 *echo.EchoRequest) (r *echo.EchoResponse, err error) {
	var _args echo.EchoServiceEchoUnaryArgs
	_args.Req1 = req1
	var _result echo.EchoServiceEchoUnaryResult
	if err = p.c.Call(ctx, "EchoUnary", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
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
