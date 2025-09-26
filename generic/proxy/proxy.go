// Copyright 2025 CloudWeGo Authors
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

package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/server/genericserver"
	"github.com/cloudwego/kitex/transport"
)

const targetDownstreamService = "cloud.kitex.service"

var (
	thriftClientsMu sync.RWMutex
	thriftClients   = make(map[string]genericclient.Client)
	pbClientsMu     sync.RWMutex
	pbClients       = make(map[string]genericclient.Client)
)

func getOrCreateThriftGenericClient(ri rpcinfo.RPCInfo, genCodeAddr net.Addr) (genericclient.Client, error) {
	serviceName := ri.Invocation().ServiceName()
	thriftClientsMu.RLock()
	if cli, ok := thriftClients[serviceName]; ok {
		thriftClientsMu.RUnlock()
		return cli, nil
	}
	thriftClientsMu.RUnlock()
	thriftClientsMu.Lock()
	defer thriftClientsMu.Unlock()
	if cli, ok := thriftClients[serviceName]; ok {
		return cli, nil
	}
	options := []client.Option{
		client.WithHostPorts(genCodeAddr.String()),
		client.WithTransportProtocol(transport.TTHeader | transport.TTHeaderStreaming),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithMetaHandler(transmeta.ClientHTTP2Handler),
	}
	cli, err := genericclient.NewClient(targetDownstreamService, generic.BinaryThriftGenericV2(serviceName), options...)
	if err != nil {
		return nil, err
	}
	thriftClients[serviceName] = cli
	return cli, nil
}

func getOrCreatePbGenericClient(ri rpcinfo.RPCInfo, genCodeAddr net.Addr) (genericclient.Client, error) {
	serviceName, packageName := ri.Invocation().ServiceName(), ri.Invocation().PackageName()
	name := fmt.Sprintf("%s:%s", packageName, serviceName)
	pbClientsMu.RLock()
	if cli, ok := pbClients[name]; ok {
		pbClientsMu.RUnlock()
		return cli, nil
	}
	pbClientsMu.RUnlock()
	pbClientsMu.Lock()
	defer pbClientsMu.Unlock()
	if cli, ok := pbClients[name]; ok {
		return cli, nil
	}
	options := []client.Option{
		client.WithHostPorts(genCodeAddr.String()),
		client.WithTransportProtocol(transport.TTHeader | transport.TTHeaderStreaming),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithMetaHandler(transmeta.ClientHTTP2Handler),
	}
	cli, err := genericclient.NewClient(targetDownstreamService, generic.BinaryPbGeneric(serviceName, packageName), options...)
	if err != nil {
		return nil, err
	}
	pbClients[name] = cli
	return cli, nil
}

func getOrCreateGenericClient(ri rpcinfo.RPCInfo, genCodeAddr net.Addr) (genericclient.Client, error) {
	switch ri.Config().PayloadCodec() {
	case serviceinfo.Thrift:
		return getOrCreateThriftGenericClient(ri, genCodeAddr)
	case serviceinfo.Protobuf:
		return getOrCreatePbGenericClient(ri, genCodeAddr)
	default:
		return nil, fmt.Errorf("unknown payload codec: %v", ri.Config().PayloadCodec())
	}
}

func newGenericProxyServer(ln net.Listener, backendAddr net.Addr, opts ...server.Option) server.Server {
	opts = append(opts, server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler))
	svr := server.NewServer(opts...)
	err := genericserver.RegisterUnknownServiceOrMethodHandler(svr, &genericserver.UnknownServiceOrMethodHandler{
		DefaultHandler:   defaultUnknownHandler(backendAddr),
		StreamingHandler: streamingUnknownHandler(backendAddr),
	})
	if err != nil {
		panic(err)
	}
	go func() {
		err := svr.Run()
		if err != nil {
			panic(err)
		}
	}()
	// wait for server starting to avoid data race
	time.Sleep(100 * time.Millisecond)
	return svr
}

func defaultUnknownHandler(genCodeAddr net.Addr) func(ctx context.Context, service, method string, request interface{}) (response interface{}, err error) {
	return func(ctx context.Context, service, method string, request interface{}) (response interface{}, err error) {
		cli, err := getOrCreateGenericClient(rpcinfo.GetRPCInfo(ctx), genCodeAddr)
		if err != nil {
			return nil, err
		}
		resp, err := cli.GenericCall(ctx, method, request)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func streamingUnknownHandler(genCodeAddr net.Addr) func(ctx context.Context, service, method string, streamSvr generic.BidiStreamingServer) (err error) {
	return func(ctx context.Context, service, method string, streamSvr generic.BidiStreamingServer) (err error) {
		cli, err := getOrCreateGenericClient(rpcinfo.GetRPCInfo(ctx), genCodeAddr)
		if err != nil {
			return err
		}
		var cancel context.CancelCauseFunc
		ctx, cancel = context.WithCancelCause(ctx)
		streamCli, err := cli.BidirectionalStreaming(ctx, method)
		if err != nil {
			return err
		}
		go func() {
			for {
				req, err := streamSvr.Recv(ctx)
				if err == io.EOF {
					if err = streamCli.CloseSend(streamCli.Context()); err != nil {
						klog.CtxErrorf(ctx, "generic proxy close send to downstream failed: %v", err)
					}
					return
				}
				if err != nil {
					cancel(err)
					klog.CtxErrorf(ctx, "generic proxy recv from upstream failed: %w", err)
					return
				}
				if err := streamCli.Send(streamCli.Context(), req); err != nil {
					klog.CtxErrorf(ctx, "generic proxy send to downstream failed: %w", err)
					return
				}
			}
		}()
		for {
			resp, err := streamCli.Recv(streamCli.Context())
			if err == io.EOF {
				return nil
			}
			if bizErr, ok := kerrors.FromBizStatusError(err); ok {
				return bizErr
			}
			if err != nil {
				return fmt.Errorf("generic proxy recv from downstream failed: %w", err)
			}
			if err := streamSvr.Send(ctx, resp); err != nil {
				cancel(err)
				return fmt.Errorf("generic proxy send to upstream failed: %w", err)
			}
		}
	}
}
