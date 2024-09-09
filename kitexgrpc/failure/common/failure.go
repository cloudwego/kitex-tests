package common

import (
	"context"
	"fmt"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/trans/netpoll"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
	"github.com/cloudwego/kitex/pkg/streaming"
	"net"
	"time"
	_ "unsafe"
)

type ClientStream interface {
	streaming.Stream
}

type ServerStream interface {
	streaming.Stream
}

type ServerStreamingClient[Res any] interface {
	Recv() (*Res, error)
	ClientStream
}

type ServerStreamingServer[Res any] interface {
	Send(res *Res) error
	ServerStream
}

type ClientStreamingClient[Req any, Res any] interface {
	Send(req *Req) error
	CloseAndRecv() (*Res, error)
	ClientStream
}

type ClientStreamingServer[Req any, Res any] interface {
	Recv() (*Req, error)
	SendAndClose(res *Res) error
	ServerStream
}

type BidiStreamingClient[Req any, Res any] interface {
	Send(req *Req) error
	Recv() (*Res, error)
	ClientStream
}

type BidiStreamingServer[Req any, Res any] interface {
	Recv() (*Req, error)
	Send(res *Res) error
	ServerStream
}

const (
	StreamingUnary         = serviceinfo.StreamingUnary
	StreamingClient        = serviceinfo.StreamingClient
	StreamingServer        = serviceinfo.StreamingServer
	StreamingBidirectional = serviceinfo.StreamingBidirectional
)

type Side int

const (
	clientSide Side = iota
	serverSide
)

func NewWrapClientStream[Req any, Resp any](mode serviceinfo.StreamingMode) *WrapClientStream[Req, Resp] {
	return &WrapClientStream[Req, Resp]{
		Mode: mode,
	}
}

type WrapClientStream[Req any, Resp any] struct {
	Mode serviceinfo.StreamingMode
	Side Side

	// Injection (connection close, timeout, context cancel)
	Injector Injector

	// server streaming
	SSC ServerStreamingClient[Resp]
	// client streaming
	CSC ClientStreamingClient[Req, Resp]
	// bidi streaming
	BSC BidiStreamingClient[Req, Resp]
}

func (ws *WrapClientStream[Req, Resp]) Send(ctx context.Context, req *Req) error {
	var sendFunc func(req *Req) error
	switch ws.Mode {
	case StreamingBidirectional:
		sendFunc = ws.BSC.Send
	case StreamingClient:
		sendFunc = ws.CSC.Send
	default:
		panic("not implemented")
	}

	ws.Injector.Execute(ctx, "Send")

	return sendFunc(req)
}

func (ws *WrapClientStream[Req, Resp]) Recv(ctx context.Context) (*Resp, error) {
	var recvFunc func() (*Resp, error)

	switch ws.Mode {
	case StreamingBidirectional:
		recvFunc = ws.BSC.Recv
	case StreamingServer:
		recvFunc = ws.SSC.Recv
	default:
		panic("not implemented")
	}

	ws.Injector.Execute(ctx, "Recv")
	return recvFunc()
}

func (ws *WrapClientStream[Req, Resp]) CloseAndRecv(ctx context.Context) (*Resp, error) {
	var closeAndRecvFunc func() (*Resp, error)
	switch ws.Mode {
	case StreamingClient:
		closeAndRecvFunc = ws.CSC.CloseAndRecv
	default:
		panic("not implemented")
	}

	ws.Injector.Execute(ctx, "CloseAndRecv")
	return closeAndRecvFunc()
}

func (ws *WrapClientStream[Req, Resp]) Close(ctx context.Context) error {
	var closeFunc func() error
	switch ws.Mode {
	case StreamingClient:
		closeFunc = ws.CSC.Close
	case StreamingServer:
		closeFunc = ws.SSC.Close
	case StreamingBidirectional:
		closeFunc = ws.BSC.Close
	}

	ws.Injector.Execute(ctx, "Close")
	return closeFunc()
}

type WrapServerStream[Req any, Resp any] struct {
	// mode
	mode serviceinfo.StreamingMode
	side Side

	// server streaming
	sss ServerStreamingServer[Resp]

	// client streaming
	css ClientStreamingServer[Req, Resp]

	// bidi streaming
	bss BidiStreamingServer[Req, Resp]
}

func (ws *WrapServerStream[Req, Resp]) Send(ctx context.Context, resp *Resp) error {
	var sendFunc func(resp *Resp) error
	switch ws.mode {
	case StreamingBidirectional:
		sendFunc = ws.bss.Send
	case StreamingServer:
		sendFunc = ws.sss.Send
	default:
		panic("not implemented")
	}
	return sendFunc(resp)
}

func (ws *WrapServerStream[Req, Resp]) Recv(ctx context.Context) (*Req, error) {
	var recvFunc func() (*Req, error)

	switch ws.mode {
	case StreamingBidirectional:
		recvFunc = ws.bss.Recv
	case StreamingClient:
		recvFunc = ws.css.Recv
	default:
		panic("not implemented")
	}

	return recvFunc()
}

func (ws *WrapServerStream[Req, Resp]) SendAndClose(ctx context.Context, resp *Resp) error {
	var closeAndRecvFunc func(resp *Resp) error
	switch ws.mode {
	case StreamingClient:
		closeAndRecvFunc = ws.css.SendAndClose
	default:
		panic("not implemented")
	}
	return closeAndRecvFunc(resp)
}

type Injector struct {
	Settings []Setting
}

func (i *Injector) Execute(ctx context.Context, key string) {
	for _, s := range i.Settings {
		err := s.Execute(ctx)
		if err != nil {
			panic(err)
		}
	}
}

type Setting struct {
	Match  func(ctx context.Context) (bool, error)
	Action func(ctx context.Context) error
}

func (s Setting) Execute(ctx context.Context) error {
	if s.Match == nil || s.Action == nil {
		return nil
	}

	b, err := s.Match(ctx)
	if err != nil {
		return err
	}
	if b {
		return s.Action(ctx)
	}
	return nil
}

func NewDummyInjector() Injector {
	return Injector{
		Settings: []Setting{
			NewDummylSetting(),
		},
	}
}

func NewDummylSetting() Setting {
	return Setting{}
}

func NewSelfCancelSetting(cancelFunc context.CancelFunc) Setting {
	return Setting{
		Match: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		Action: func(ctx context.Context) error {
			if cnt := ctx.Value("cnt"); cnt != nil {
				if cnt.(int) == 1 {
					cancelFunc()
					return nil
				}
			}
			return nil
		},
	}
}

func NewCloseConnSetting() Setting {
	return Setting{
		Match: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		Action: func(ctx context.Context) error {
			if cnt := ctx.Value("cnt"); cnt != nil {
				if cnt.(int) == 1 {
					if conn, err := controller.GetConn(""); err == nil {
						conn.Close()
						return nil
					}
				}
			}
			return nil
		},
	}
}

var controller = &ConnController{
	ConnMap: map[string]net.Conn{},
}

type ConnController struct {
	ConnMap map[string]net.Conn
}

func (c *ConnController) GetConn(key string) (net.Conn, error) {
	for s := range c.ConnMap {
		if c.ConnMap[s] != nil {
			return c.ConnMap[s], nil
		}
	}
	return nil, fmt.Errorf("no such connection")
}

func (c *ConnController) PutConn(conn net.Conn) error {
	key := c.connKey(conn)
	c.ConnMap[key] = conn
	return nil
}

func (c *ConnController) connKey(conn net.Conn) string {
	return conn.LocalAddr().String() + "|" + conn.RemoteAddr().String()
}

var _ remote.Dialer = &WrapDialer{}

type WrapDialer struct {
	remote.Dialer
	connController *ConnController
}

func NewWrapDialer() *WrapDialer {
	npDialer := netpoll.NewDialer()
	return &WrapDialer{
		Dialer:         npDialer,
		connController: controller,
	}
}

func (d *WrapDialer) DialTimeout(network string, address string, timeout time.Duration) (net.Conn, error) {
	conn, err := d.Dialer.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}

	d.connController.PutConn(conn)
	return conn, nil
}
