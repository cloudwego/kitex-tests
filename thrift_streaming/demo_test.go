package thrift_streaming

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo"
	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/echo/echoservice"
	"github.com/cloudwego/kitex/pkg/endpoint"
	sep "github.com/cloudwego/kitex/pkg/endpoint/server"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/server"
)

func Test_Demo(t *testing.T) {
	svr := echoservice.NewServer(&thriftServerTimeoutImpl{
		Handler: func(stream echo.EchoService_EchoBidirectionalServer) (err error) {
			if _, err = stream.Recv(); err != nil {
				return err
			}
			return stream.Send(&echo.EchoResponse{Message: "ok"})
		},
	}, WithServerAddr(":8080"), server.WithExitWaitTime(time.Millisecond*10),
		server.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				// do something
				return next(ctx, req, resp)
			}
		}),
		server.WithUnaryOptions(server.WithUnaryRPCTimeout(time.Second),
			server.WithUnaryMiddleware(func(next sep.UnaryEndpoint) sep.UnaryEndpoint {
				return func(ctx context.Context, req, resp interface{}) (err error) {
					// do something
					return next(ctx, req, resp)
				}
			})),
		server.WithStreamOptions(server.WithStreamMiddleware(func(next sep.StreamEndpoint) sep.StreamEndpoint {
			return func(ctx context.Context, st streaming.ServerStream) (err error) {
				// do something
				return next(ctx, st)
			}
		}), server.WithStreamRecvMiddleware(func(next sep.StreamRecvEndpoint) sep.StreamRecvEndpoint {
			return func(stream streaming.ServerStream, message interface{}) (err error) {
				// do something
				return next(stream, message)
			}
		})),
	)
	// svr.RegisterService(xxxService, xxxHandler)
	if err := svr.Run(); err != nil {
		panic(err)
	}
}
