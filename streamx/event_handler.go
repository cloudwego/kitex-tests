// Copyright 2026 CloudWeGo Authors
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
	"runtime/debug"
	"sync/atomic"
	"testing"

	"github.com/cloudwego/kitex/pkg/rpcinfo"

	"github.com/cloudwego/kitex-tests/pkg/test"
)

type ClientCalledTimes struct {
	startTimes      uint32
	recvHeaderTimes uint32
	recvTimes       uint32
	sendTimes       uint32
	finishTimes     uint32
}

type ServerCalledTimes struct {
	startTimes  uint32
	recvTimes   uint32
	sendTimes   uint32
	finishTimes uint32
}

type ClientMockEventHandler struct {
	handleStreamStartEvent      func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamStartEvent)
	startTimes                  uint32
	handleStreamRecvHeaderEvent func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvHeaderEvent)
	recvHeaderTimes             uint32
	handleStreamRecvEvent       func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvEvent)
	recvTimes                   uint32
	handleStreamSendEvent       func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamSendEvent)
	sendTimes                   uint32
	handleStreamFinishEvent     func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent)
	finishTimes                 uint32
}

func newClientMockEventHandler() *ClientMockEventHandler {
	return &ClientMockEventHandler{
		handleStreamStartEvent:      func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamStartEvent) {},
		handleStreamRecvHeaderEvent: func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvHeaderEvent) {},
		handleStreamRecvEvent:       func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvEvent) {},
		handleStreamSendEvent:       func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamSendEvent) {},
		handleStreamFinishEvent:     func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {},
	}
}

func (hdl *ClientMockEventHandler) HandleStreamStartEvent(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamStartEvent) {
	atomic.AddUint32(&hdl.startTimes, 1)
	hdl.handleStreamStartEvent(ctx, ri, evt)
}

func (hdl *ClientMockEventHandler) HandleStreamRecvHeaderEvent(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvHeaderEvent) {
	atomic.AddUint32(&hdl.recvHeaderTimes, 1)
	hdl.handleStreamRecvHeaderEvent(ctx, ri, evt)
}

func (hdl *ClientMockEventHandler) HandleStreamRecvEvent(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvEvent) {
	atomic.AddUint32(&hdl.recvTimes, 1)
	hdl.handleStreamRecvEvent(ctx, ri, evt)
}

func (hdl *ClientMockEventHandler) HandleStreamSendEvent(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamSendEvent) {
	atomic.AddUint32(&hdl.sendTimes, 1)
	hdl.handleStreamSendEvent(ctx, ri, evt)
}

func (hdl *ClientMockEventHandler) HandleStreamFinishEvent(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {
	atomic.AddUint32(&hdl.finishTimes, 1)
	hdl.handleStreamFinishEvent(ctx, ri, evt)
}

func (hdl *ClientMockEventHandler) AssertCalledTimes(t *testing.T, calledTimes ClientCalledTimes) {
	hdl.assertHandleStreamStartEventTimes(t, calledTimes.startTimes)
	hdl.assertHandleStreamRecvHeaderEventTimes(t, calledTimes.recvHeaderTimes)
	hdl.assertHandleStreamRecvEventTimes(t, calledTimes.recvTimes)
	hdl.assertHandleStreamSendEventTimes(t, calledTimes.sendTimes)
	hdl.assertHandleStreamFinishEventTimes(t, calledTimes.finishTimes)
}

func (hdl *ClientMockEventHandler) assertHandleStreamStartEventTimes(t *testing.T, times uint32) {
	actual := atomic.LoadUint32(&hdl.startTimes)
	test.Assert(t, actual == times, actual, times)
}

func (hdl *ClientMockEventHandler) assertHandleStreamRecvHeaderEventTimes(t *testing.T, times uint32) {
	actual := atomic.LoadUint32(&hdl.recvHeaderTimes)
	test.Assert(t, actual == times, actual, times)
}

func (hdl *ClientMockEventHandler) assertHandleStreamRecvEventTimes(t *testing.T, times uint32) {
	actual := atomic.LoadUint32(&hdl.recvTimes)
	test.Assert(t, actual == times, actual, times)
}

func (hdl *ClientMockEventHandler) assertHandleStreamSendEventTimes(t *testing.T, times uint32) {
	actual := atomic.LoadUint32(&hdl.sendTimes)
	test.Assert(t, actual == times, actual, times, string(debug.Stack()))
}

func (hdl *ClientMockEventHandler) assertHandleStreamFinishEventTimes(t *testing.T, times uint32) {
	actual := atomic.LoadUint32(&hdl.finishTimes)
	test.Assert(t, actual == times, actual, times)
}

func (hdl *ClientMockEventHandler) Reset() {
	*hdl = ClientMockEventHandler{
		handleStreamStartEvent:      func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamStartEvent) {},
		handleStreamRecvHeaderEvent: func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvHeaderEvent) {},
		handleStreamRecvEvent:       func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvEvent) {},
		handleStreamSendEvent:       func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamSendEvent) {},
		handleStreamFinishEvent:     func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {},
	}
}

type ServerMockEventHandler struct {
	handleStreamStartEvent  func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamStartEvent)
	startTimes              uint32
	handleStreamRecvEvent   func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvEvent)
	recvTimes               uint32
	handleStreamSendEvent   func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamSendEvent)
	sendTimes               uint32
	handleStreamFinishEvent func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent)
	finishTimes             uint32
	testChan                chan eventHandleTestSuite
}

func newServerMockEventHandler(testChan chan eventHandleTestSuite) *ServerMockEventHandler {
	return &ServerMockEventHandler{
		handleStreamStartEvent:  func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamStartEvent) {},
		handleStreamRecvEvent:   func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvEvent) {},
		handleStreamSendEvent:   func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamSendEvent) {},
		handleStreamFinishEvent: func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {},
		testChan:                testChan,
	}
}

func (hdl *ServerMockEventHandler) HandleStreamStartEvent(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamStartEvent) {
	atomic.AddUint32(&hdl.startTimes, 1)
	hdl.handleStreamStartEvent(ctx, ri, evt)
}

func (hdl *ServerMockEventHandler) HandleStreamRecvEvent(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvEvent) {
	atomic.AddUint32(&hdl.recvTimes, 1)
	hdl.handleStreamRecvEvent(ctx, ri, evt)
}

func (hdl *ServerMockEventHandler) HandleStreamSendEvent(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamSendEvent) {
	atomic.AddUint32(&hdl.sendTimes, 1)
	hdl.handleStreamSendEvent(ctx, ri, evt)
}

func (hdl *ServerMockEventHandler) HandleStreamFinishEvent(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {
	atomic.AddUint32(&hdl.finishTimes, 1)
	hdl.handleStreamFinishEvent(ctx, ri, evt)
}

func (hdl *ServerMockEventHandler) AssertCalledTimes(t *testing.T, calledTimes ServerCalledTimes) {
	// in server-side, EventHandler test should be invoked in Tracer.Finish so that FinishEvent could be verified
	hdl.testChan <- struct {
		hdl         *ServerMockEventHandler
		t           *testing.T
		calledTimes ServerCalledTimes
	}{
		hdl:         hdl,
		t:           t,
		calledTimes: calledTimes}
}

func (hdl *ServerMockEventHandler) assertCalledTimes(t *testing.T, calledTimes ServerCalledTimes) {
	hdl.assertHandleStreamStartEventTimes(t, calledTimes.startTimes)
	hdl.assertHandleStreamRecvEventTimes(t, calledTimes.recvTimes)
	hdl.assertHandleStreamSendEventTimes(t, calledTimes.sendTimes)
	hdl.assertHandleStreamFinishEventTimes(t, calledTimes.finishTimes)
}

func (hdl *ServerMockEventHandler) assertHandleStreamStartEventTimes(t *testing.T, times uint32) {
	actual := atomic.LoadUint32(&hdl.startTimes)
	test.Assert(t, actual == times, actual, times, string(debug.Stack()))
}

func (hdl *ServerMockEventHandler) assertHandleStreamRecvEventTimes(t *testing.T, times uint32) {
	actual := atomic.LoadUint32(&hdl.recvTimes)
	test.Assert(t, actual == times, actual, times)
}

func (hdl *ServerMockEventHandler) assertHandleStreamSendEventTimes(t *testing.T, times uint32) {
	actual := atomic.LoadUint32(&hdl.sendTimes)
	test.Assert(t, actual == times, actual, times)
}

func (hdl *ServerMockEventHandler) assertHandleStreamFinishEventTimes(t *testing.T, times uint32) {
	actual := atomic.LoadUint32(&hdl.finishTimes)
	test.Assert(t, actual == times, actual, times)
}

func (hdl *ServerMockEventHandler) Reset() {
	ch := hdl.testChan
	*hdl = ServerMockEventHandler{
		handleStreamStartEvent:  func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamStartEvent) {},
		handleStreamRecvEvent:   func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamRecvEvent) {},
		handleStreamSendEvent:   func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamSendEvent) {},
		handleStreamFinishEvent: func(ctx context.Context, ri rpcinfo.RPCInfo, evt rpcinfo.StreamFinishEvent) {},
		testChan:                ch,
	}
}

type eventHandleTestSuite struct {
	hdl         *ServerMockEventHandler
	t           *testing.T
	calledTimes ServerCalledTimes
}

type ServerEventHandlerTracer struct {
	testChan chan eventHandleTestSuite
}

func newServerEventHandlerTracer() (*ServerEventHandlerTracer, chan eventHandleTestSuite) {
	ch := make(chan eventHandleTestSuite, 1)
	return &ServerEventHandlerTracer{testChan: ch}, ch
}

func (s *ServerEventHandlerTracer) Start(ctx context.Context) context.Context {
	return ctx
}

func (s *ServerEventHandlerTracer) Finish(ctx context.Context) {
	suite := <-s.testChan
	suite.hdl.assertCalledTimes(suite.t, suite.calledTimes)
	suite.hdl.Reset()
}
