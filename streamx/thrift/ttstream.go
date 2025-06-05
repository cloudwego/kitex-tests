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

package streamx_thrift

import (
	"context"
	"sync/atomic"

	"github.com/cloudwego/kitex/pkg/remote/trans/ttstream"
	"github.com/cloudwego/kitex/pkg/streaming"
)

type metaFrameHandler struct {
	executed uint32
}

func (m *metaFrameHandler) OnMetaFrame(ctx context.Context, intHeader ttstream.IntHeader, header streaming.Header, payload []byte) error {
	atomic.StoreUint32(&m.executed, 1)
	return nil
}

type headerFrameHandler struct {
	executed uint32
}

func (h *headerFrameHandler) OnReadStream(ctx context.Context, ihd ttstream.IntHeader, shd ttstream.StrHeader) (context.Context, error) {
	atomic.StoreUint32(&h.executed, 1)
	return ctx, nil
}

var ttmh, tthh = &metaFrameHandler{}, &headerFrameHandler{}
