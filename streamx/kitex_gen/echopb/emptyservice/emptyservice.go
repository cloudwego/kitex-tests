// Code generated by Kitex v0.12.3. DO NOT EDIT.

package emptyservice

import (
	"errors"
	echopb "github.com/cloudwego/kitex-tests/streamx/kitex_gen/echopb"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{}

var (
	emptyServiceServiceInfo = NewServiceInfo()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return emptyServiceServiceInfo
}

// NewServiceInfo creates a new ServiceInfo
func NewServiceInfo() *kitex.ServiceInfo {
	return newServiceInfo()
}

func newServiceInfo() *kitex.ServiceInfo {
	serviceName := "EmptyService"
	handlerType := (*echopb.EmptyService)(nil)
	extra := map[string]interface{}{
		"PackageName": "echopb",
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         serviceMethods,
		PayloadCodec:    kitex.Protobuf,
		KiteXGenVersion: "v0.12.3",
		Extra:           extra,
	}
	return svcInfo
}

type kClient struct {
	c  client.Client
	sc client.Streaming
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c:  c,
		sc: c.(client.Streaming),
	}
}
