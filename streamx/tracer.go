// Copyright 2025 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package streamx

import (
	"context"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
)

func NewTracer() stats.Tracer {
	return &tracer{}
}

type tracer struct {
}

func (t *tracer) Start(ctx context.Context) context.Context {
	return ctx
}

func (t *tracer) Finish(ctx context.Context) {
	return
}

func (t *tracer) ReportStreamEvent(ctx context.Context, ri rpcinfo.RPCInfo, event rpcinfo.Event) {
	switch event.Event() {
	case stats.StreamSend:
		if ri.Invocation().Extra("test_stream_send_event") == nil {
			var count int
			ri.Invocation().(rpcinfo.InvocationSetter).SetExtra("test_stream_send_event", &count)
		}
		count := ri.Invocation().Extra("test_stream_send_event").(*int)
		*count++
	case stats.StreamRecv:
		if ri.Invocation().Extra("test_stream_recv_event") == nil {
			var count int
			ri.Invocation().(rpcinfo.InvocationSetter).SetExtra("test_stream_recv_event", &count)
		}
		count := ri.Invocation().Extra("test_stream_recv_event").(*int)
		*count++
	}
	return
}
