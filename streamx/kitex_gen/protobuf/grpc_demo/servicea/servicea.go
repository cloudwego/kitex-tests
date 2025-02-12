// Code generated by Kitex v0.12.1. DO NOT EDIT.

package servicea

import (
	"context"
	"errors"
	grpc_demo "github.com/cloudwego/kitex-tests/streamx/kitex_gen/protobuf/grpc_demo"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
	proto "google.golang.org/protobuf/proto"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{
	"CallUnary": kitex.NewMethodInfo(
		callUnaryHandler,
		newCallUnaryArgs,
		newCallUnaryResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingUnary),
	),
	"CallClientStream": kitex.NewMethodInfo(
		callClientStreamHandler,
		newCallClientStreamArgs,
		newCallClientStreamResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingClient),
	),
	"CallServerStream": kitex.NewMethodInfo(
		callServerStreamHandler,
		newCallServerStreamArgs,
		newCallServerStreamResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingServer),
	),
	"CallBidiStream": kitex.NewMethodInfo(
		callBidiStreamHandler,
		newCallBidiStreamArgs,
		newCallBidiStreamResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingBidirectional),
	),
}

var (
	serviceAServiceInfo = NewServiceInfo()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return serviceAServiceInfo
}

// NewServiceInfo creates a new ServiceInfo
func NewServiceInfo() *kitex.ServiceInfo {
	return newServiceInfo()
}

func newServiceInfo() *kitex.ServiceInfo {
	serviceName := "ServiceA"
	handlerType := (*grpc_demo.ServiceA)(nil)
	extra := map[string]interface{}{
		"PackageName": "grpc_demo",
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         serviceMethods,
		PayloadCodec:    kitex.Protobuf,
		KiteXGenVersion: "v0.12.1",
		Extra:           extra,
	}
	return svcInfo
}

func callUnaryHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new(grpc_demo.Request)
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.(grpc_demo.ServiceA).CallUnary(ctx, req)
		if err != nil {
			return err
		}
		return st.SendMsg(resp)
	case *CallUnaryArgs:
		success, err := handler.(grpc_demo.ServiceA).CallUnary(ctx, s.Req)
		if err != nil {
			return err
		}
		realResult := result.(*CallUnaryResult)
		realResult.Success = success
		return nil
	default:
		return errInvalidMessageType
	}
}

func newCallUnaryArgs() interface{} {
	return &CallUnaryArgs{}
}

func newCallUnaryResult() interface{} {
	return &CallUnaryResult{}
}

type CallUnaryArgs struct {
	Req *grpc_demo.Request
}

func (p *CallUnaryArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(grpc_demo.Request)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *CallUnaryArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *CallUnaryArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *CallUnaryArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *CallUnaryArgs) Unmarshal(in []byte) error {
	msg := new(grpc_demo.Request)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var CallUnaryArgs_Req_DEFAULT *grpc_demo.Request

func (p *CallUnaryArgs) GetReq() *grpc_demo.Request {
	if !p.IsSetReq() {
		return CallUnaryArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *CallUnaryArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *CallUnaryArgs) GetFirstArgument() interface{} {
	return p.Req
}

type CallUnaryResult struct {
	Success *grpc_demo.Reply
}

var CallUnaryResult_Success_DEFAULT *grpc_demo.Reply

func (p *CallUnaryResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(grpc_demo.Reply)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *CallUnaryResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *CallUnaryResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *CallUnaryResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *CallUnaryResult) Unmarshal(in []byte) error {
	msg := new(grpc_demo.Reply)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *CallUnaryResult) GetSuccess() *grpc_demo.Reply {
	if !p.IsSetSuccess() {
		return CallUnaryResult_Success_DEFAULT
	}
	return p.Success
}

func (p *CallUnaryResult) SetSuccess(x interface{}) {
	p.Success = x.(*grpc_demo.Reply)
}

func (p *CallUnaryResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *CallUnaryResult) GetResult() interface{} {
	return p.Success
}

func callClientStreamHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, err := streaming.GetServerStreamFromArg(arg)
	if err != nil {
		return err
	}
	stream := streaming.NewClientStreamingServer[grpc_demo.Request, grpc_demo.Reply](st)
	return handler.(grpc_demo.ServiceA).CallClientStream(ctx, stream)
}

func newCallClientStreamArgs() interface{} {
	return &CallClientStreamArgs{}
}

func newCallClientStreamResult() interface{} {
	return &CallClientStreamResult{}
}

type CallClientStreamArgs struct {
	Req *grpc_demo.Request
}

func (p *CallClientStreamArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(grpc_demo.Request)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *CallClientStreamArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *CallClientStreamArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *CallClientStreamArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *CallClientStreamArgs) Unmarshal(in []byte) error {
	msg := new(grpc_demo.Request)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var CallClientStreamArgs_Req_DEFAULT *grpc_demo.Request

func (p *CallClientStreamArgs) GetReq() *grpc_demo.Request {
	if !p.IsSetReq() {
		return CallClientStreamArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *CallClientStreamArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *CallClientStreamArgs) GetFirstArgument() interface{} {
	return p.Req
}

type CallClientStreamResult struct {
	Success *grpc_demo.Reply
}

var CallClientStreamResult_Success_DEFAULT *grpc_demo.Reply

func (p *CallClientStreamResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(grpc_demo.Reply)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *CallClientStreamResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *CallClientStreamResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *CallClientStreamResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *CallClientStreamResult) Unmarshal(in []byte) error {
	msg := new(grpc_demo.Reply)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *CallClientStreamResult) GetSuccess() *grpc_demo.Reply {
	if !p.IsSetSuccess() {
		return CallClientStreamResult_Success_DEFAULT
	}
	return p.Success
}

func (p *CallClientStreamResult) SetSuccess(x interface{}) {
	p.Success = x.(*grpc_demo.Reply)
}

func (p *CallClientStreamResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *CallClientStreamResult) GetResult() interface{} {
	return p.Success
}

func callServerStreamHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, err := streaming.GetServerStreamFromArg(arg)
	if err != nil {
		return err
	}
	stream := streaming.NewServerStreamingServer[grpc_demo.Reply](st)
	req := new(grpc_demo.Request)
	if err := stream.RecvMsg(ctx, req); err != nil {
		return err
	}
	return handler.(grpc_demo.ServiceA).CallServerStream(ctx, req, stream)
}

func newCallServerStreamArgs() interface{} {
	return &CallServerStreamArgs{}
}

func newCallServerStreamResult() interface{} {
	return &CallServerStreamResult{}
}

type CallServerStreamArgs struct {
	Req *grpc_demo.Request
}

func (p *CallServerStreamArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(grpc_demo.Request)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *CallServerStreamArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *CallServerStreamArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *CallServerStreamArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *CallServerStreamArgs) Unmarshal(in []byte) error {
	msg := new(grpc_demo.Request)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var CallServerStreamArgs_Req_DEFAULT *grpc_demo.Request

func (p *CallServerStreamArgs) GetReq() *grpc_demo.Request {
	if !p.IsSetReq() {
		return CallServerStreamArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *CallServerStreamArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *CallServerStreamArgs) GetFirstArgument() interface{} {
	return p.Req
}

type CallServerStreamResult struct {
	Success *grpc_demo.Reply
}

var CallServerStreamResult_Success_DEFAULT *grpc_demo.Reply

func (p *CallServerStreamResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(grpc_demo.Reply)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *CallServerStreamResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *CallServerStreamResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *CallServerStreamResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *CallServerStreamResult) Unmarshal(in []byte) error {
	msg := new(grpc_demo.Reply)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *CallServerStreamResult) GetSuccess() *grpc_demo.Reply {
	if !p.IsSetSuccess() {
		return CallServerStreamResult_Success_DEFAULT
	}
	return p.Success
}

func (p *CallServerStreamResult) SetSuccess(x interface{}) {
	p.Success = x.(*grpc_demo.Reply)
}

func (p *CallServerStreamResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *CallServerStreamResult) GetResult() interface{} {
	return p.Success
}

func callBidiStreamHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, err := streaming.GetServerStreamFromArg(arg)
	if err != nil {
		return err
	}
	stream := streaming.NewBidiStreamingServer[grpc_demo.Request, grpc_demo.Reply](st)
	return handler.(grpc_demo.ServiceA).CallBidiStream(ctx, stream)
}

func newCallBidiStreamArgs() interface{} {
	return &CallBidiStreamArgs{}
}

func newCallBidiStreamResult() interface{} {
	return &CallBidiStreamResult{}
}

type CallBidiStreamArgs struct {
	Req *grpc_demo.Request
}

func (p *CallBidiStreamArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(grpc_demo.Request)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *CallBidiStreamArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *CallBidiStreamArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *CallBidiStreamArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *CallBidiStreamArgs) Unmarshal(in []byte) error {
	msg := new(grpc_demo.Request)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var CallBidiStreamArgs_Req_DEFAULT *grpc_demo.Request

func (p *CallBidiStreamArgs) GetReq() *grpc_demo.Request {
	if !p.IsSetReq() {
		return CallBidiStreamArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *CallBidiStreamArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *CallBidiStreamArgs) GetFirstArgument() interface{} {
	return p.Req
}

type CallBidiStreamResult struct {
	Success *grpc_demo.Reply
}

var CallBidiStreamResult_Success_DEFAULT *grpc_demo.Reply

func (p *CallBidiStreamResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(grpc_demo.Reply)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *CallBidiStreamResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *CallBidiStreamResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *CallBidiStreamResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *CallBidiStreamResult) Unmarshal(in []byte) error {
	msg := new(grpc_demo.Reply)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *CallBidiStreamResult) GetSuccess() *grpc_demo.Reply {
	if !p.IsSetSuccess() {
		return CallBidiStreamResult_Success_DEFAULT
	}
	return p.Success
}

func (p *CallBidiStreamResult) SetSuccess(x interface{}) {
	p.Success = x.(*grpc_demo.Reply)
}

func (p *CallBidiStreamResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *CallBidiStreamResult) GetResult() interface{} {
	return p.Success
}

type kClient struct {
	c  client.Client
	sc client.Streaming
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c:  c,
		sc: c.(client.Streaming),
	}
}

func (p *kClient) CallUnary(ctx context.Context, Req *grpc_demo.Request) (r *grpc_demo.Reply, err error) {
	var _args CallUnaryArgs
	_args.Req = Req
	var _result CallUnaryResult
	if err = p.c.Call(ctx, "CallUnary", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) CallClientStream(ctx context.Context) (ServiceA_CallClientStreamClient, error) {
	st, err := p.sc.StreamX(ctx, "CallClientStream")
	if err != nil {
		return nil, err
	}
	stream := streaming.NewClientStreamingClient[grpc_demo.Request, grpc_demo.Reply](st)
	return stream, nil
}

func (p *kClient) CallServerStream(ctx context.Context, req *grpc_demo.Request) (ServiceA_CallServerStreamClient, error) {
	st, err := p.sc.StreamX(ctx, "CallServerStream")
	if err != nil {
		return nil, err
	}
	stream := streaming.NewServerStreamingClient[grpc_demo.Reply](st)
	if err := stream.SendMsg(ctx, req); err != nil {
		return nil, err
	}
	if err := stream.CloseSend(ctx); err != nil {
		return nil, err
	}
	return stream, nil
}

func (p *kClient) CallBidiStream(ctx context.Context) (ServiceA_CallBidiStreamClient, error) {
	st, err := p.sc.StreamX(ctx, "CallBidiStream")
	if err != nil {
		return nil, err
	}
	stream := streaming.NewBidiStreamingClient[grpc_demo.Request, grpc_demo.Reply](st)
	return stream, nil
}
