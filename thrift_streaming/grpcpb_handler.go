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
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"

	"github.com/cloudwego/kitex-tests/thrift_streaming/kitex_gen/grpc_pb"
)

type GRPCPBServiceImpl struct{}

func (G *GRPCPBServiceImpl) Echo(stream grpc_pb.PBService_EchoServer) (err error) {
	klog.Infof("[GRPCPBServiceImpl.Echo] Echo called")
	count := GetInt(stream.Context(), KeyCount, 0)
	klog.Infof("count: %d", count)
	for i := 0; i < count; i++ {
		var req *grpc_pb.Request
		if req, err = stream.Recv(); err != nil {
			klog.Infof("[GRPCPBServiceImpl.Echo] recv error: %v", err)
			return err
		}
		klog.Infof("[GRPCPBServiceImpl.Echo] recv success: %s", req.Message)

		resp := &grpc_pb.Response{
			Message: req.Message,
		}
		if err = stream.Send(resp); err != nil {
			klog.Infof("[GRPCPBServiceImpl.Echo] send error: %v", err)
			return
		}
		klog.Infof("[GRPCPBServiceImpl.Echo] send success: %v", resp)
	}
	return
}

func (G *GRPCPBServiceImpl) EchoClient(stream grpc_pb.PBService_EchoClientServer) (err error) {
	klog.Infof("[GRPCPBServiceImpl.EchoClient] EchoClient called")
	count := GetInt(stream.Context(), KeyCount, 0)
	klog.Infof("count: %d", count)
	for i := 0; i < count; i++ {
		var req *grpc_pb.Request
		if req, err = stream.Recv(); err != nil {
			klog.Infof("[GRPCPBServiceImpl.EchoClient] recv error: %v", err)
			return err
		}
		klog.Infof("[GRPCPBServiceImpl.EchoClient] recv success: %s", req.Message)
	}

	resp := &grpc_pb.Response{
		Message: strconv.Itoa(count),
	}
	err = stream.SendAndClose(resp)
	klog.Infof("[GRPCPBServiceImpl.EchoClient] send: err = %v, resp = %v", err, resp)
	return err
}

func (G *GRPCPBServiceImpl) EchoServer(req *grpc_pb.Request, stream grpc_pb.PBService_EchoServerServer) (err error) {
	klog.Infof("[GRPCPBServiceImpl.EchoServer] recv req = %s", req.Message)
	count := GetInt(stream.Context(), KeyCount, 0)
	klog.Infof("count: %d", count)
	for i := 0; i < count; i++ {
		resp := &grpc_pb.Response{
			Message: req.Message + "-" + strconv.Itoa(i),
		}
		if err = stream.Send(resp); err != nil {
			klog.Infof("[GRPCPBServiceImpl.EchoServer] send error: %v", err)
			return err
		}
		klog.Infof("[GRPCPBServiceImpl.EchoServer] send success: %s", resp.Message)
	}
	return nil
}

func (G *GRPCPBServiceImpl) EchoPingPong(ctx context.Context, req *grpc_pb.Request) (resp *grpc_pb.Response, err error) {
	klog.Infof("[GRPCPBServiceImpl.EchoPingPong] recv: %s", req.Message)
	resp = &grpc_pb.Response{
		Message: req.Message,
	}
	klog.Infof("[GRPCPBServiceImpl.EchoPingPong] send: %s", resp.Message)
	return resp, nil
}
