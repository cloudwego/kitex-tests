// Code generated by Kitex v0.12.3. DO NOT EDIT.

package pbservice

import (
	"context"
	"errors"
	"fmt"
	grpc_pb "github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
	proto "google.golang.org/protobuf/proto"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{
	"Echo": kitex.NewMethodInfo(
		echoHandler,
		newEchoArgs,
		newEchoResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingBidirectional),
	),
	"EchoClient": kitex.NewMethodInfo(
		echoClientHandler,
		newEchoClientArgs,
		newEchoClientResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingClient),
	),
	"EchoServer": kitex.NewMethodInfo(
		echoServerHandler,
		newEchoServerArgs,
		newEchoServerResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingServer),
	),
	"EchoPingPong": kitex.NewMethodInfo(
		echoPingPongHandler,
		newEchoPingPongArgs,
		newEchoPingPongResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingUnary),
	),
}

var (
	pBServiceServiceInfo                = NewServiceInfo()
	pBServiceServiceInfoForClient       = NewServiceInfoForClient()
	pBServiceServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return pBServiceServiceInfo
}

// for stream client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return pBServiceServiceInfoForStreamClient
}

// for client
func serviceInfoForClient() *kitex.ServiceInfo {
	return pBServiceServiceInfoForClient
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
	serviceName := "PBService"
	handlerType := (*grpc_pb.PBService)(nil)
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
		"PackageName": "grpc_pb",
	}
	if hasStreaming {
		extra["streaming"] = hasStreaming
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         methods,
		PayloadCodec:    kitex.Protobuf,
		KiteXGenVersion: "v0.12.3",
		Extra:           extra,
	}
	return svcInfo
}

func echoHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	streamingArgs, ok := arg.(*streaming.Args)
	if !ok {
		return errInvalidMessageType
	}
	st := streamingArgs.Stream
	stream := &pBServiceEchoServer{st}
	return handler.(grpc_pb.PBService).Echo(stream)
}

type pBServiceEchoClient struct {
	streaming.Stream
}

func (x *pBServiceEchoClient) DoFinish(err error) {
	if finisher, ok := x.Stream.(streaming.WithDoFinish); ok {
		finisher.DoFinish(err)
	} else {
		panic(fmt.Sprintf("streaming.WithDoFinish is not implemented by %T", x.Stream))
	}
}
func (x *pBServiceEchoClient) Send(m *grpc_pb.Request) error {
	return x.Stream.SendMsg(m)
}
func (x *pBServiceEchoClient) Recv() (*grpc_pb.Response, error) {
	m := new(grpc_pb.Response)
	return m, x.Stream.RecvMsg(m)
}

type pBServiceEchoServer struct {
	streaming.Stream
}

func (x *pBServiceEchoServer) Send(m *grpc_pb.Response) error {
	return x.Stream.SendMsg(m)
}

func (x *pBServiceEchoServer) Recv() (*grpc_pb.Request, error) {
	m := new(grpc_pb.Request)
	return m, x.Stream.RecvMsg(m)
}

func newEchoArgs() interface{} {
	return &EchoArgs{}
}

func newEchoResult() interface{} {
	return &EchoResult{}
}

type EchoArgs struct {
	Req *grpc_pb.Request
}

func (p *EchoArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(grpc_pb.Request)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *EchoArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *EchoArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *EchoArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *EchoArgs) Unmarshal(in []byte) error {
	msg := new(grpc_pb.Request)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var EchoArgs_Req_DEFAULT *grpc_pb.Request

func (p *EchoArgs) GetReq() *grpc_pb.Request {
	if !p.IsSetReq() {
		return EchoArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *EchoArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *EchoArgs) GetFirstArgument() interface{} {
	return p.Req
}

type EchoResult struct {
	Success *grpc_pb.Response
}

var EchoResult_Success_DEFAULT *grpc_pb.Response

func (p *EchoResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(grpc_pb.Response)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *EchoResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *EchoResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *EchoResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *EchoResult) Unmarshal(in []byte) error {
	msg := new(grpc_pb.Response)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *EchoResult) GetSuccess() *grpc_pb.Response {
	if !p.IsSetSuccess() {
		return EchoResult_Success_DEFAULT
	}
	return p.Success
}

func (p *EchoResult) SetSuccess(x interface{}) {
	p.Success = x.(*grpc_pb.Response)
}

func (p *EchoResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *EchoResult) GetResult() interface{} {
	return p.Success
}

func echoClientHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	streamingArgs, ok := arg.(*streaming.Args)
	if !ok {
		return errInvalidMessageType
	}
	st := streamingArgs.Stream
	stream := &pBServiceEchoClientServer{st}
	return handler.(grpc_pb.PBService).EchoClient(stream)
}

type pBServiceEchoClientClient struct {
	streaming.Stream
}

func (x *pBServiceEchoClientClient) DoFinish(err error) {
	if finisher, ok := x.Stream.(streaming.WithDoFinish); ok {
		finisher.DoFinish(err)
	} else {
		panic(fmt.Sprintf("streaming.WithDoFinish is not implemented by %T", x.Stream))
	}
}
func (x *pBServiceEchoClientClient) Send(m *grpc_pb.Request) error {
	return x.Stream.SendMsg(m)
}
func (x *pBServiceEchoClientClient) CloseAndRecv() (*grpc_pb.Response, error) {
	if err := x.Stream.Close(); err != nil {
		return nil, err
	}
	m := new(grpc_pb.Response)
	return m, x.Stream.RecvMsg(m)
}

type pBServiceEchoClientServer struct {
	streaming.Stream
}

func (x *pBServiceEchoClientServer) SendAndClose(m *grpc_pb.Response) error {
	return x.Stream.SendMsg(m)
}

func (x *pBServiceEchoClientServer) Recv() (*grpc_pb.Request, error) {
	m := new(grpc_pb.Request)
	return m, x.Stream.RecvMsg(m)
}

func newEchoClientArgs() interface{} {
	return &EchoClientArgs{}
}

func newEchoClientResult() interface{} {
	return &EchoClientResult{}
}

type EchoClientArgs struct {
	Req *grpc_pb.Request
}

func (p *EchoClientArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(grpc_pb.Request)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *EchoClientArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *EchoClientArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *EchoClientArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *EchoClientArgs) Unmarshal(in []byte) error {
	msg := new(grpc_pb.Request)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var EchoClientArgs_Req_DEFAULT *grpc_pb.Request

func (p *EchoClientArgs) GetReq() *grpc_pb.Request {
	if !p.IsSetReq() {
		return EchoClientArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *EchoClientArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *EchoClientArgs) GetFirstArgument() interface{} {
	return p.Req
}

type EchoClientResult struct {
	Success *grpc_pb.Response
}

var EchoClientResult_Success_DEFAULT *grpc_pb.Response

func (p *EchoClientResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(grpc_pb.Response)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *EchoClientResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *EchoClientResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *EchoClientResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *EchoClientResult) Unmarshal(in []byte) error {
	msg := new(grpc_pb.Response)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *EchoClientResult) GetSuccess() *grpc_pb.Response {
	if !p.IsSetSuccess() {
		return EchoClientResult_Success_DEFAULT
	}
	return p.Success
}

func (p *EchoClientResult) SetSuccess(x interface{}) {
	p.Success = x.(*grpc_pb.Response)
}

func (p *EchoClientResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *EchoClientResult) GetResult() interface{} {
	return p.Success
}

func echoServerHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	streamingArgs, ok := arg.(*streaming.Args)
	if !ok {
		return errInvalidMessageType
	}
	st := streamingArgs.Stream
	stream := &pBServiceEchoServerServer{st}
	req := new(grpc_pb.Request)
	if err := st.RecvMsg(req); err != nil {
		return err
	}
	return handler.(grpc_pb.PBService).EchoServer(req, stream)
}

type pBServiceEchoServerClient struct {
	streaming.Stream
}

func (x *pBServiceEchoServerClient) DoFinish(err error) {
	if finisher, ok := x.Stream.(streaming.WithDoFinish); ok {
		finisher.DoFinish(err)
	} else {
		panic(fmt.Sprintf("streaming.WithDoFinish is not implemented by %T", x.Stream))
	}
}
func (x *pBServiceEchoServerClient) Recv() (*grpc_pb.Response, error) {
	m := new(grpc_pb.Response)
	return m, x.Stream.RecvMsg(m)
}

type pBServiceEchoServerServer struct {
	streaming.Stream
}

func (x *pBServiceEchoServerServer) Send(m *grpc_pb.Response) error {
	return x.Stream.SendMsg(m)
}

func newEchoServerArgs() interface{} {
	return &EchoServerArgs{}
}

func newEchoServerResult() interface{} {
	return &EchoServerResult{}
}

type EchoServerArgs struct {
	Req *grpc_pb.Request
}

func (p *EchoServerArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(grpc_pb.Request)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *EchoServerArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *EchoServerArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *EchoServerArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *EchoServerArgs) Unmarshal(in []byte) error {
	msg := new(grpc_pb.Request)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var EchoServerArgs_Req_DEFAULT *grpc_pb.Request

func (p *EchoServerArgs) GetReq() *grpc_pb.Request {
	if !p.IsSetReq() {
		return EchoServerArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *EchoServerArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *EchoServerArgs) GetFirstArgument() interface{} {
	return p.Req
}

type EchoServerResult struct {
	Success *grpc_pb.Response
}

var EchoServerResult_Success_DEFAULT *grpc_pb.Response

func (p *EchoServerResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(grpc_pb.Response)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *EchoServerResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *EchoServerResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *EchoServerResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *EchoServerResult) Unmarshal(in []byte) error {
	msg := new(grpc_pb.Response)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *EchoServerResult) GetSuccess() *grpc_pb.Response {
	if !p.IsSetSuccess() {
		return EchoServerResult_Success_DEFAULT
	}
	return p.Success
}

func (p *EchoServerResult) SetSuccess(x interface{}) {
	p.Success = x.(*grpc_pb.Response)
}

func (p *EchoServerResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *EchoServerResult) GetResult() interface{} {
	return p.Success
}

func echoPingPongHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new(grpc_pb.Request)
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.(grpc_pb.PBService).EchoPingPong(ctx, req)
		if err != nil {
			return err
		}
		return st.SendMsg(resp)
	case *EchoPingPongArgs:
		success, err := handler.(grpc_pb.PBService).EchoPingPong(ctx, s.Req)
		if err != nil {
			return err
		}
		realResult := result.(*EchoPingPongResult)
		realResult.Success = success
		return nil
	default:
		return errInvalidMessageType
	}
}
func newEchoPingPongArgs() interface{} {
	return &EchoPingPongArgs{}
}

func newEchoPingPongResult() interface{} {
	return &EchoPingPongResult{}
}

type EchoPingPongArgs struct {
	Req *grpc_pb.Request
}

func (p *EchoPingPongArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(grpc_pb.Request)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *EchoPingPongArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *EchoPingPongArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *EchoPingPongArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *EchoPingPongArgs) Unmarshal(in []byte) error {
	msg := new(grpc_pb.Request)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var EchoPingPongArgs_Req_DEFAULT *grpc_pb.Request

func (p *EchoPingPongArgs) GetReq() *grpc_pb.Request {
	if !p.IsSetReq() {
		return EchoPingPongArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *EchoPingPongArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *EchoPingPongArgs) GetFirstArgument() interface{} {
	return p.Req
}

type EchoPingPongResult struct {
	Success *grpc_pb.Response
}

var EchoPingPongResult_Success_DEFAULT *grpc_pb.Response

func (p *EchoPingPongResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(grpc_pb.Response)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *EchoPingPongResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *EchoPingPongResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *EchoPingPongResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *EchoPingPongResult) Unmarshal(in []byte) error {
	msg := new(grpc_pb.Response)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *EchoPingPongResult) GetSuccess() *grpc_pb.Response {
	if !p.IsSetSuccess() {
		return EchoPingPongResult_Success_DEFAULT
	}
	return p.Success
}

func (p *EchoPingPongResult) SetSuccess(x interface{}) {
	p.Success = x.(*grpc_pb.Response)
}

func (p *EchoPingPongResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *EchoPingPongResult) GetResult() interface{} {
	return p.Success
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) Echo(ctx context.Context) (PBService_EchoClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "Echo", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &pBServiceEchoClient{res.Stream}
	return stream, nil
}

func (p *kClient) EchoClient(ctx context.Context) (PBService_EchoClientClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "EchoClient", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &pBServiceEchoClientClient{res.Stream}
	return stream, nil
}

func (p *kClient) EchoServer(ctx context.Context, req *grpc_pb.Request) (PBService_EchoServerClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "EchoServer", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &pBServiceEchoServerClient{res.Stream}

	if err := stream.Stream.SendMsg(req); err != nil {
		return nil, err
	}
	if err := stream.Stream.Close(); err != nil {
		return nil, err
	}
	return stream, nil
}

func (p *kClient) EchoPingPong(ctx context.Context, Req *grpc_pb.Request) (r *grpc_pb.Response, err error) {
	var _args EchoPingPongArgs
	_args.Req = Req
	var _result EchoPingPongResult
	if err = p.c.Call(ctx, "EchoPingPong", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}
