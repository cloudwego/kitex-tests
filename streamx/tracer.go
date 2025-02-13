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
	if event.Event() == stats.StreamSend {
		if ri.Invocation().Extra("test_stream_send_event") == nil {
			var count int
			ri.Invocation().(rpcinfo.InvocationSetter).SetExtra("test_stream_send_event", &count)
		}
		count := ri.Invocation().Extra("test_stream_send_event").(*int)
		*count++
	} else {
		if ri.Invocation().Extra("test_stream_recv_event") == nil {
			var count int
			ri.Invocation().(rpcinfo.InvocationSetter).SetExtra("test_stream_recv_event", &count)
		}
		count := ri.Invocation().Extra("test_stream_recv_event").(*int)
		*count++
	}
	return
}
