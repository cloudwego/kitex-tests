// Copyright 2024 CloudWeGo Authors
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

package normalcall

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
	remote_transmeta "github.com/cloudwego/kitex/pkg/remote/transmeta"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/pkg/test"
	"github.com/cloudwego/kitex-tests/pkg/utils/serverutils"
	"github.com/cloudwego/kitex-tests/thriftrpc"
)

var (
	rpcTimeout = 10 * time.Millisecond
	sleepTime  = 20 * time.Millisecond
)

type timeoutHandler struct {
	stability.STService
	testSTReq func(context.Context, *stability.STRequest) (*stability.STResponse, error)
}

func (t *timeoutHandler) TestSTReq(ctx context.Context, req *stability.STRequest) (r *stability.STResponse, err error) {
	return t.testSTReq(ctx, req)
}

var _ remote.MetaHandler = (*timeoutMetaHandler)(nil)

type timeoutMetaHandler struct {
	timeout time.Duration
}

// WriteMeta implements remote.MetaHandler, for client to set timeout.
func (t *timeoutMetaHandler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	if t.timeout > 0 {
		hd := map[uint16]string{
			remote_transmeta.RPCTimeout: strconv.Itoa(int(t.timeout.Milliseconds())),
		}
		msg.TransInfo().PutTransIntInfo(hd)
	}
	return ctx, nil
}

// ReadMeta implements remote.MetaHandler, for server to get timeout.
func (t *timeoutMetaHandler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	if t.timeout > 0 {
		ri := msg.RPCInfo()
		if cfg := rpcinfo.AsMutableRPCConfig(ri.Config()); cfg != nil {
			_ = cfg.SetRPCTimeout(t.timeout)
		}
	}
	return ctx, nil
}

func TestServerProcessingTimeout(t *testing.T) {
	t.Run("both-ttheader-meta-handler/not-enabled", func(t *testing.T) {
		addr := serverutils.NextListenAddr()
		svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{Network: "tcp", Address: addr},
			&timeoutHandler{
				testSTReq: func(ctx context.Context, request *stability.STRequest) (*stability.STResponse, error) {
					_, ok := ctx.Deadline()
					test.Assert(t, !ok)
					time.Sleep(sleepTime)
					return &stability.STResponse{Str: request.Str}, nil
				},
			},
			server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
			// server.WithEnableContextTimeout(false), // disable by default
		)
		serverutils.Wait(addr)
		defer svr.Stop()

		cli := getKitexClient(transport.TTHeaderFramed, client.WithMetaHandler(transmeta.ClientTTHeaderHandler))
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())

		_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(addr))
		test.Assert(t, err == nil, err)
	})

	t.Run("both-ttheader-meta-handler/enabled", func(t *testing.T) {
		addr := serverutils.NextListenAddr()
		svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{Network: "tcp", Address: addr},
			&timeoutHandler{
				testSTReq: func(ctx context.Context, request *stability.STRequest) (*stability.STResponse, error) {
					ddl, ok := ctx.Deadline()
					test.Assert(t, ok)
					test.Assert(t, ddl.After(time.Now()))
					time.Sleep(sleepTime)
					return &stability.STResponse{Str: request.Str}, nil
				},
			},
			server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
			server.WithEnableContextTimeout(true),
		)
		serverutils.Wait(addr)
		defer svr.Stop()

		cli := getKitexClient(transport.TTHeaderFramed, client.WithMetaHandler(transmeta.ClientTTHeaderHandler))
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())

		_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(addr), callopt.WithRPCTimeout(rpcTimeout))
		test.Assert(t, errors.Is(err, kerrors.ErrRPCTimeout), err)
	})

	t.Run("client-set-ttheader/server-ttheader-meta-handler/not-enabled", func(t *testing.T) {
		addr := serverutils.NextListenAddr()
		svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{Network: "tcp", Address: addr},
			&timeoutHandler{
				testSTReq: func(ctx context.Context, request *stability.STRequest) (*stability.STResponse, error) {
					_, ok := ctx.Deadline()
					test.Assert(t, !ok)
					time.Sleep(sleepTime)
					return &stability.STResponse{Str: request.Str}, nil
				},
			},
			server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
			// server.WithEnableContextTimeout(false), // default is false
		)
		serverutils.Wait(addr)
		defer svr.Stop()

		cli := getKitexClient(transport.TTHeaderFramed, client.WithMetaHandler(&timeoutMetaHandler{
			timeout: rpcTimeout,
		}))
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())

		_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(addr))
		test.Assert(t, err == nil, err)
	})

	t.Run("client-set-timeout/server-ttheader-meta-handler/enabled", func(t *testing.T) {
		addr := serverutils.NextListenAddr()
		svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{Network: "tcp", Address: addr},
			&timeoutHandler{
				testSTReq: func(ctx context.Context, request *stability.STRequest) (*stability.STResponse, error) {
					ddl, ok := ctx.Deadline()
					test.Assert(t, ok)
					test.Assert(t, ddl.After(time.Now()))
					time.Sleep(sleepTime)
					return &stability.STResponse{Str: request.Str}, nil
				},
			},
			server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
			server.WithEnableContextTimeout(true),
		)
		serverutils.Wait(addr)
		defer svr.Stop()

		cli := getKitexClient(transport.TTHeaderFramed, client.WithMetaHandler(&timeoutMetaHandler{
			timeout: rpcTimeout,
		}))
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())

		_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(addr))
		test.Assert(t, err == nil, err) // no timeout control in both client and server
	})

	t.Run("client-not-set/server-set-timeout/enabled", func(t *testing.T) {
		addr := serverutils.NextListenAddr()
		svr := thriftrpc.RunServer(&thriftrpc.ServerInitParam{Network: "tcp", Address: addr},
			&timeoutHandler{
				testSTReq: func(ctx context.Context, request *stability.STRequest) (*stability.STResponse, error) {
					ddl, ok := ctx.Deadline()
					test.Assert(t, ok)
					test.Assert(t, ddl.After(time.Now()))
					time.Sleep(sleepTime)
					return &stability.STResponse{Str: request.Str}, nil
				},
			},
			server.WithMetaHandler(&timeoutMetaHandler{
				timeout: rpcTimeout,
			}),
			server.WithEnableContextTimeout(true),
		)
		serverutils.Wait(addr)
		defer svr.Stop()

		cli := getKitexClient(transport.TTHeaderFramed)
		ctx, stReq := thriftrpc.CreateSTRequest(context.Background())

		_, err := cli.TestSTReq(ctx, stReq, callopt.WithHostPort(addr))
		test.Assert(t, err == nil, err) // no timeout control in both client and server
	})
}
