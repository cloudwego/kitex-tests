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

package utils

import (
	"context"

	"github.com/cloudwego/kitex/pkg/generic"
)

// ServiceV2Iface defines an interface that extends and implements all the methods required by ServiceV2,
// which can simplify the construction of the ServiceV2 class when used in conjunction with the ServiceV2Iface2ServiceV2 method.
type ServiceV2Iface interface {
	// GenericCall handle the generic call
	GenericCall(ctx context.Context, service, method string, request interface{}) (response interface{}, err error)
	// ClientStreaming handle the client streaming call
	ClientStreaming(ctx context.Context, service, method string, stream generic.ClientStreamingServer) (err error)
	// ServerStreaming handle the server streaming call
	ServerStreaming(ctx context.Context, service, method string, request interface{}, stream generic.ServerStreamingServer) (err error)
	// BidiStreaming handle the bidi streaming call
	BidiStreaming(ctx context.Context, service, method string, stream generic.BidiStreamingServer) (err error)
}

// ServiceV2Iface2ServiceV2 converts ServiceV2Iface to ServiceV2.
func ServiceV2Iface2ServiceV2(iface ServiceV2Iface) *generic.ServiceV2 {
	return &generic.ServiceV2{
		GenericCall:     iface.GenericCall,
		ClientStreaming: iface.ClientStreaming,
		ServerStreaming: iface.ServerStreaming,
		BidiStreaming:   iface.BidiStreaming,
	}
}
