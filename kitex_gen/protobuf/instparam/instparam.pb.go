//
// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.13.0
// source: instparam.proto

package instparam

import (
	context "context"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SubMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *SubMessage) Reset() {
	*x = SubMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_instparam_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubMessage) ProtoMessage() {}

func (x *SubMessage) ProtoReflect() protoreflect.Message {
	mi := &file_instparam_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubMessage.ProtoReflect.Descriptor instead.
func (*SubMessage) Descriptor() ([]byte, []int) {
	return file_instparam_proto_rawDescGZIP(), []int{0}
}

func (x *SubMessage) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *SubMessage) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          int64         `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Value       string        `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	SubMessages []*SubMessage `protobuf:"bytes,3,rep,name=subMessages,proto3" json:"subMessages,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_instparam_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_instparam_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_instparam_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Message) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Message) GetSubMessages() []*SubMessage {
	if x != nil {
		return x.SubMessages
	}
	return nil
}

// 复杂参数
type ObjReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg      *Message               `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	MsgMap   map[string]*SubMessage `protobuf:"bytes,2,rep,name=msgMap,proto3" json:"msgMap,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	SubMsgs  []*SubMessage          `protobuf:"bytes,3,rep,name=subMsgs,proto3" json:"subMsgs,omitempty"`
	Msgs     []*Message             `protobuf:"bytes,4,rep,name=msgs,proto3" json:"msgs,omitempty"`
	FlagMsg  string                 `protobuf:"bytes,5,opt,name=flagMsg,proto3" json:"flagMsg,omitempty"`
	MockCost string                 `protobuf:"bytes,6,opt,name=mockCost,proto3" json:"mockCost,omitempty"`
}

func (x *ObjReq) Reset() {
	*x = ObjReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_instparam_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObjReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObjReq) ProtoMessage() {}

func (x *ObjReq) ProtoReflect() protoreflect.Message {
	mi := &file_instparam_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObjReq.ProtoReflect.Descriptor instead.
func (*ObjReq) Descriptor() ([]byte, []int) {
	return file_instparam_proto_rawDescGZIP(), []int{2}
}

func (x *ObjReq) GetMsg() *Message {
	if x != nil {
		return x.Msg
	}
	return nil
}

func (x *ObjReq) GetMsgMap() map[string]*SubMessage {
	if x != nil {
		return x.MsgMap
	}
	return nil
}

func (x *ObjReq) GetSubMsgs() []*SubMessage {
	if x != nil {
		return x.SubMsgs
	}
	return nil
}

func (x *ObjReq) GetMsgs() []*Message {
	if x != nil {
		return x.Msgs
	}
	return nil
}

func (x *ObjReq) GetFlagMsg() string {
	if x != nil {
		return x.FlagMsg
	}
	return ""
}

func (x *ObjReq) GetMockCost() string {
	if x != nil {
		return x.MockCost
	}
	return ""
}

type ObjResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg     *Message               `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	MsgMap  map[string]*SubMessage `protobuf:"bytes,2,rep,name=msgMap,proto3" json:"msgMap,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	SubMsgs []*SubMessage          `protobuf:"bytes,3,rep,name=subMsgs,proto3" json:"subMsgs,omitempty"`
	Msgs    []*Message             `protobuf:"bytes,4,rep,name=msgs,proto3" json:"msgs,omitempty"`
	FlagMsg string                 `protobuf:"bytes,5,opt,name=flagMsg,proto3" json:"flagMsg,omitempty"`
}

func (x *ObjResp) Reset() {
	*x = ObjResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_instparam_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObjResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObjResp) ProtoMessage() {}

func (x *ObjResp) ProtoReflect() protoreflect.Message {
	mi := &file_instparam_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObjResp.ProtoReflect.Descriptor instead.
func (*ObjResp) Descriptor() ([]byte, []int) {
	return file_instparam_proto_rawDescGZIP(), []int{3}
}

func (x *ObjResp) GetMsg() *Message {
	if x != nil {
		return x.Msg
	}
	return nil
}

func (x *ObjResp) GetMsgMap() map[string]*SubMessage {
	if x != nil {
		return x.MsgMap
	}
	return nil
}

func (x *ObjResp) GetSubMsgs() []*SubMessage {
	if x != nil {
		return x.SubMsgs
	}
	return nil
}

func (x *ObjResp) GetMsgs() []*Message {
	if x != nil {
		return x.Msgs
	}
	return nil
}

func (x *ObjResp) GetFlagMsg() string {
	if x != nil {
		return x.FlagMsg
	}
	return ""
}

var File_instparam_proto protoreflect.FileDescriptor

var file_instparam_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x09, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x22, 0x32, 0x0a, 0x0a,
	0x53, 0x75, 0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x22, 0x68, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x37, 0x0a, 0x0b, 0x73, 0x75, 0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72,
	0x61, 0x6d, 0x2e, 0x53, 0x75, 0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x0b, 0x73,
	0x75, 0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x22, 0xc6, 0x02, 0x0a, 0x06, 0x4f,
	0x62, 0x6a, 0x52, 0x65, 0x71, 0x12, 0x24, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x35, 0x0a, 0x06, 0x6d,
	0x73, 0x67, 0x4d, 0x61, 0x70, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x69, 0x6e,
	0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x4f, 0x62, 0x6a, 0x52, 0x65, 0x71, 0x2e, 0x4d,
	0x73, 0x67, 0x4d, 0x61, 0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x6d, 0x73, 0x67, 0x4d,
	0x61, 0x70, 0x12, 0x2f, 0x0a, 0x07, 0x73, 0x75, 0x62, 0x4d, 0x73, 0x67, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e,
	0x53, 0x75, 0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x73, 0x75, 0x62, 0x4d,
	0x73, 0x67, 0x73, 0x12, 0x26, 0x0a, 0x04, 0x6d, 0x73, 0x67, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x12, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x04, 0x6d, 0x73, 0x67, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x66,
	0x6c, 0x61, 0x67, 0x4d, 0x73, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x66, 0x6c,
	0x61, 0x67, 0x4d, 0x73, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x73,
	0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6d, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x73,
	0x74, 0x1a, 0x50, 0x0a, 0x0b, 0x4d, 0x73, 0x67, 0x4d, 0x61, 0x70, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x2b, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x15, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x53, 0x75,
	0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a,
	0x02, 0x38, 0x01, 0x22, 0xac, 0x02, 0x0a, 0x07, 0x4f, 0x62, 0x6a, 0x52, 0x65, 0x73, 0x70, 0x12,
	0x24, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x69,
	0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x36, 0x0a, 0x06, 0x6d, 0x73, 0x67, 0x4d, 0x61, 0x70, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61,
	0x6d, 0x2e, 0x4f, 0x62, 0x6a, 0x52, 0x65, 0x73, 0x70, 0x2e, 0x4d, 0x73, 0x67, 0x4d, 0x61, 0x70,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x6d, 0x73, 0x67, 0x4d, 0x61, 0x70, 0x12, 0x2f, 0x0a,
	0x07, 0x73, 0x75, 0x62, 0x4d, 0x73, 0x67, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15,
	0x2e, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x53, 0x75, 0x62, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x73, 0x75, 0x62, 0x4d, 0x73, 0x67, 0x73, 0x12, 0x26,
	0x0a, 0x04, 0x6d, 0x73, 0x67, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x69,
	0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x52, 0x04, 0x6d, 0x73, 0x67, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x66, 0x6c, 0x61, 0x67, 0x4d, 0x73,
	0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x66, 0x6c, 0x61, 0x67, 0x4d, 0x73, 0x67,
	0x1a, 0x50, 0x0a, 0x0b, 0x4d, 0x73, 0x67, 0x4d, 0x61, 0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x2b, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x15, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x53, 0x75, 0x62,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x42, 0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x77, 0x65, 0x67, 0x6f, 0x2f, 0x6b, 0x69, 0x74, 0x65, 0x78,
	0x2d, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2f, 0x6b, 0x69, 0x74, 0x65, 0x78, 0x5f, 0x67, 0x65, 0x6e,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x69, 0x6e, 0x73, 0x74, 0x70, 0x61,
	0x72, 0x61, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_instparam_proto_rawDescOnce sync.Once
	file_instparam_proto_rawDescData = file_instparam_proto_rawDesc
)

func file_instparam_proto_rawDescGZIP() []byte {
	file_instparam_proto_rawDescOnce.Do(func() {
		file_instparam_proto_rawDescData = protoimpl.X.CompressGZIP(file_instparam_proto_rawDescData)
	})
	return file_instparam_proto_rawDescData
}

var file_instparam_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_instparam_proto_goTypes = []interface{}{
	(*SubMessage)(nil), // 0: instparam.SubMessage
	(*Message)(nil),    // 1: instparam.Message
	(*ObjReq)(nil),     // 2: instparam.ObjReq
	(*ObjResp)(nil),    // 3: instparam.ObjResp
	nil,                // 4: instparam.ObjReq.MsgMapEntry
	nil,                // 5: instparam.ObjResp.MsgMapEntry
}
var file_instparam_proto_depIdxs = []int32{
	0,  // 0: instparam.Message.subMessages:type_name -> instparam.SubMessage
	1,  // 1: instparam.ObjReq.msg:type_name -> instparam.Message
	4,  // 2: instparam.ObjReq.msgMap:type_name -> instparam.ObjReq.MsgMapEntry
	0,  // 3: instparam.ObjReq.subMsgs:type_name -> instparam.SubMessage
	1,  // 4: instparam.ObjReq.msgs:type_name -> instparam.Message
	1,  // 5: instparam.ObjResp.msg:type_name -> instparam.Message
	5,  // 6: instparam.ObjResp.msgMap:type_name -> instparam.ObjResp.MsgMapEntry
	0,  // 7: instparam.ObjResp.subMsgs:type_name -> instparam.SubMessage
	1,  // 8: instparam.ObjResp.msgs:type_name -> instparam.Message
	0,  // 9: instparam.ObjReq.MsgMapEntry.value:type_name -> instparam.SubMessage
	0,  // 10: instparam.ObjResp.MsgMapEntry.value:type_name -> instparam.SubMessage
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_instparam_proto_init() }
func file_instparam_proto_init() {
	if File_instparam_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_instparam_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_instparam_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_instparam_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObjReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_instparam_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObjResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_instparam_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_instparam_proto_goTypes,
		DependencyIndexes: file_instparam_proto_depIdxs,
		MessageInfos:      file_instparam_proto_msgTypes,
	}.Build()
	File_instparam_proto = out.File
	file_instparam_proto_rawDesc = nil
	file_instparam_proto_goTypes = nil
	file_instparam_proto_depIdxs = nil
}

var _ context.Context
