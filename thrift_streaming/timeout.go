// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package thrift_streaming

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/streamclient"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/server"
)

// TODO: move to cloudwego/kitex? or cloudwego/kitex-contrib?

type ctxCancelKey struct{}

type contextCancelMetaHandler struct{}

func (c contextCancelMetaHandler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}

func (c contextCancelMetaHandler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}

// OnConnectStream is called by client and will return a new context with a cancel function
func (c contextCancelMetaHandler) OnConnectStream(ctx context.Context) (context.Context, error) {
	newCtx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(newCtx, ctxCancelKey{}, cancel)
	return ctx, nil
}

// OnReadStream is called by server and will return a new context with a cancel function
func (c contextCancelMetaHandler) OnReadStream(ctx context.Context) (context.Context, error) {
	newCtx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(newCtx, ctxCancelKey{}, cancel)
	return ctx, nil
}

type withCancelServerSuite struct{}

func (c *withCancelServerSuite) Options() []server.Option {
	return []server.Option{
		server.WithMetaHandler(&contextCancelMetaHandler{}),
		server.WithMiddleware(func(endpoint endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				cancel, ok := ctx.Value(ctxCancelKey{}).(context.CancelFunc)
				if ok {
					defer cancel()
				}
				err = endpoint(ctx, req, resp)
				return
			}
		}),
	}
}

type withCancelClientSuite struct{}

func (c *withCancelClientSuite) Options() []client.Option {
	return []client.Option{
		client.WithMetaHandler(&contextCancelMetaHandler{}),
		client.WithMiddleware(func(endpoint endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				cancel, ok := ctx.Value(ctxCancelKey{}).(context.CancelFunc)
				if ok {
					defer cancel()
				}
				err = endpoint(ctx, req, resp)
				return
			}
		}),
	}
}

// WithServerContextCancel sets the context cancel function for server, and invoke it with middleware.
func WithServerContextCancel() server.Option {
	return server.WithSuite(&withCancelServerSuite{})
}

// WithClientContextCancel sets the context cancel function for client, and invoke it with middleware.
func WithClientContextCancel() streamclient.Option {
	return streamclient.ConvertOptionFrom(client.WithSuite(&withCancelClientSuite{}))
}

func CallWithCtx(ctx context.Context, timeout time.Duration, f func() (err error)) error {
	cancel, _ := ctx.Value(ctxCancelKey{}).(context.CancelFunc)
	return streaming.CallWithTimeout(timeout, cancel, f)
}
