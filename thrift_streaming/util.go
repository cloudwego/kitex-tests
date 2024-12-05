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
	"errors"
	"strconv"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
)

const (
	KeyError               = "ERROR"
	KeyMessage             = "MESSAGE"
	KeyCount               = "COUNT"
	KeyServerRecvTimeoutMS = "RECV_TIMEOUT_MS"
	KeyServerSendTimeoutMS = "SEND_TIMEOUT_MS"
	KeyGetServerConn       = "GET_SERVER_CONN"
	KeyInspectMWCtx        = "INSPECT_MW_CTX"
)

func GetError(ctx context.Context) error {
	if err, exists := metainfo.GetValue(ctx, KeyError); exists {
		return errors.New(err)
	}
	return nil
}

func GetErrorHTTP2(ctx context.Context) error {
	err := GetKeyHTTP2(ctx, KeyError, "")
	if err != "" {
		return errors.New(err)
	}
	return nil
}

func GetIntHTTP2(ctx context.Context, key string, defaultValue int) int {
	if i, err := strconv.Atoi(GetKeyHTTP2(ctx, key, "")); err == nil {
		return i
	}
	return defaultValue
}

func GetKeyHTTP2(ctx context.Context, key, defaultValue string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return defaultValue
	}
	if v := md.Get(key); len(v) > 0 {
		return v[0]
	}
	return defaultValue
}

func AddKeyValueHTTP2(ctx context.Context, key, value string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, key, value)
}

func GetValue(ctx context.Context, key, defaultValue string) string {
	if v, exists := metainfo.GetValue(ctx, key); exists {
		return v
	}
	return defaultValue
}

func GetInt(ctx context.Context, key string, defaultValue int) int {
	if i, err := strconv.Atoi(GetValue(ctx, key, "")); err == nil {
		return i
	}
	return defaultValue
}

func GetBool(ctx context.Context, key string, defaultBool bool) bool {
	if b, err := strconv.ParseBool(GetValue(ctx, key, "")); err == nil {
		return b
	}
	return defaultBool
}
