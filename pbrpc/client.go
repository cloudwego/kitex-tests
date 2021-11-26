// Copyright 2021 CloudWeGo Authors
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

package pbrpc

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/protobuf/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/utils"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/connpool"
	"github.com/cloudwego/kitex/transport"
)

// ConnectionMode .
type ConnectionMode int

// Modes .
const (
	ShortConnection ConnectionMode = iota
	LongConnection
	ConnectionMultiplexed
)

// ClientInitParam .
type ClientInitParam struct {
	TargetServiceName string
	HostPorts         []string
	Protocol          transport.Protocol
	ConnMode          ConnectionMode
}

// CreateKitexClient .
func CreateKitexClient(param *ClientInitParam, opts ...client.Option) stservice.Client {
	if len(param.HostPorts) > 0 {
		opts = append(opts, client.WithHostPorts(param.HostPorts...))
	}

	if param.Protocol != transport.PurePayload {
		opts = append(opts, client.WithTransportProtocol(param.Protocol))
	}

	switch param.ConnMode {
	case LongConnection:
		opts = append(opts, client.WithLongConnection(
			connpool.IdleConfig{
				MaxIdlePerAddress: 1000,
				MaxIdleGlobal:     1000 * 10,
				MaxIdleTimeout:    30 * time.Second,
			}))
	case ConnectionMultiplexed:
		opts = append(opts, client.WithMuxConnection(4))
	default:
	}

	return stservice.MustNewClient(param.TargetServiceName, opts...)
}

// CreateSTRequest .
func CreateSTRequest(ctx context.Context) (context.Context, *stability.STRequest) {
	req := &stability.STRequest{}
	req.Name = "byted"
	req.On = true
	req.D = 0.0
	req.Str = utils.RandomString(100)
	req.StringMap = map[string]string{
		"key1": utils.RandomString(100),
		"key2": utils.RandomString(10),
	}
	req.StringList = []string{utils.RandomString(10), utils.RandomString(20), utils.RandomString(30)}
	req.E = stability.TestEnum_FIRST

	ctx = metainfo.WithValue(ctx, "Key1", "Val1")
	ctx = metainfo.WithPersistentValue(ctx, "Key1", "Val1")
	return ctx, req
}

// CreateObjReq .
func CreateObjReq(ctx context.Context) (context.Context, *instparam.ObjReq) {
	id := int64(rand.Intn(100))
	subMsg1 := &instparam.SubMessage{
		Id:    id,
		Value: (utils.RandomString(100)),
	}
	subMsg2 := &instparam.SubMessage{
		Id:    (math.MaxInt64),
		Value: (utils.RandomString(10)),
	}
	subMsgList := []*instparam.SubMessage{subMsg1, subMsg2}

	msg := &instparam.Message{}
	msg.Id = id
	msg.Value = (utils.RandomString(100))
	msg.SubMessages = subMsgList

	req := &instparam.ObjReq{}
	req.Msg = msg
	req.MsgMap = map[string]*instparam.SubMessage{
		utils.RandomString(10): subMsg1,
	}
	req.SubMsgs = subMsgList

	ctx = metainfo.WithValue(ctx, "Key1", "Val1")
	ctx = metainfo.WithPersistentValue(ctx, "Key1", "Val1")
	return ctx, req
}
