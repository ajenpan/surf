// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.18.1
// source: core.proto

package core

import (
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

type NotifyClientDisconnect_MSGID int32

const (
	NotifyClientDisconnect___invilid__MSGID NotifyClientDisconnect_MSGID = 0
	NotifyClientDisconnect_ID               NotifyClientDisconnect_MSGID = 100001
)

// Enum value maps for NotifyClientDisconnect_MSGID.
var (
	NotifyClientDisconnect_MSGID_name = map[int32]string{
		0:      "__invilid__MSGID",
		100001: "ID",
	}
	NotifyClientDisconnect_MSGID_value = map[string]int32{
		"__invilid__MSGID": 0,
		"ID":               100001,
	}
)

func (x NotifyClientDisconnect_MSGID) Enum() *NotifyClientDisconnect_MSGID {
	p := new(NotifyClientDisconnect_MSGID)
	*p = x
	return p
}

func (x NotifyClientDisconnect_MSGID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NotifyClientDisconnect_MSGID) Descriptor() protoreflect.EnumDescriptor {
	return file_core_proto_enumTypes[0].Descriptor()
}

func (NotifyClientDisconnect_MSGID) Type() protoreflect.EnumType {
	return &file_core_proto_enumTypes[0]
}

func (x NotifyClientDisconnect_MSGID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NotifyClientDisconnect_MSGID.Descriptor instead.
func (NotifyClientDisconnect_MSGID) EnumDescriptor() ([]byte, []int) {
	return file_core_proto_rawDescGZIP(), []int{0, 0}
}

type NotifyClientDisconnect_Reason int32

const (
	NotifyClientDisconnect_Disconnect NotifyClientDisconnect_Reason = 0
	NotifyClientDisconnect_BeKicked   NotifyClientDisconnect_Reason = 1
	NotifyClientDisconnect_GateClosed NotifyClientDisconnect_Reason = 2
)

// Enum value maps for NotifyClientDisconnect_Reason.
var (
	NotifyClientDisconnect_Reason_name = map[int32]string{
		0: "Disconnect",
		1: "BeKicked",
		2: "GateClosed",
	}
	NotifyClientDisconnect_Reason_value = map[string]int32{
		"Disconnect": 0,
		"BeKicked":   1,
		"GateClosed": 2,
	}
)

func (x NotifyClientDisconnect_Reason) Enum() *NotifyClientDisconnect_Reason {
	p := new(NotifyClientDisconnect_Reason)
	*p = x
	return p
}

func (x NotifyClientDisconnect_Reason) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NotifyClientDisconnect_Reason) Descriptor() protoreflect.EnumDescriptor {
	return file_core_proto_enumTypes[1].Descriptor()
}

func (NotifyClientDisconnect_Reason) Type() protoreflect.EnumType {
	return &file_core_proto_enumTypes[1]
}

func (x NotifyClientDisconnect_Reason) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NotifyClientDisconnect_Reason.Descriptor instead.
func (NotifyClientDisconnect_Reason) EnumDescriptor() ([]byte, []int) {
	return file_core_proto_rawDescGZIP(), []int{0, 1}
}

type NotifyClientConnect_MSGID int32

const (
	NotifyClientConnect___invilid__MSGID NotifyClientConnect_MSGID = 0
	NotifyClientConnect_ID               NotifyClientConnect_MSGID = 100002
)

// Enum value maps for NotifyClientConnect_MSGID.
var (
	NotifyClientConnect_MSGID_name = map[int32]string{
		0:      "__invilid__MSGID",
		100002: "ID",
	}
	NotifyClientConnect_MSGID_value = map[string]int32{
		"__invilid__MSGID": 0,
		"ID":               100002,
	}
)

func (x NotifyClientConnect_MSGID) Enum() *NotifyClientConnect_MSGID {
	p := new(NotifyClientConnect_MSGID)
	*p = x
	return p
}

func (x NotifyClientConnect_MSGID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NotifyClientConnect_MSGID) Descriptor() protoreflect.EnumDescriptor {
	return file_core_proto_enumTypes[2].Descriptor()
}

func (NotifyClientConnect_MSGID) Type() protoreflect.EnumType {
	return &file_core_proto_enumTypes[2]
}

func (x NotifyClientConnect_MSGID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NotifyClientConnect_MSGID.Descriptor instead.
func (NotifyClientConnect_MSGID) EnumDescriptor() ([]byte, []int) {
	return file_core_proto_rawDescGZIP(), []int{1, 0}
}

type NotifyClientDisconnect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid        uint32                        `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	GateNodeId uint32                        `protobuf:"varint,2,opt,name=gate_node_id,json=gateNodeId,proto3" json:"gate_node_id,omitempty"`
	Reason     NotifyClientDisconnect_Reason `protobuf:"varint,3,opt,name=reason,proto3,enum=openproto.core.NotifyClientDisconnect_Reason" json:"reason,omitempty"`
}

func (x *NotifyClientDisconnect) Reset() {
	*x = NotifyClientDisconnect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyClientDisconnect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyClientDisconnect) ProtoMessage() {}

func (x *NotifyClientDisconnect) ProtoReflect() protoreflect.Message {
	mi := &file_core_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyClientDisconnect.ProtoReflect.Descriptor instead.
func (*NotifyClientDisconnect) Descriptor() ([]byte, []int) {
	return file_core_proto_rawDescGZIP(), []int{0}
}

func (x *NotifyClientDisconnect) GetUid() uint32 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *NotifyClientDisconnect) GetGateNodeId() uint32 {
	if x != nil {
		return x.GateNodeId
	}
	return 0
}

func (x *NotifyClientDisconnect) GetReason() NotifyClientDisconnect_Reason {
	if x != nil {
		return x.Reason
	}
	return NotifyClientDisconnect_Disconnect
}

type NotifyClientConnect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid        uint32 `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	GateNodeId uint32 `protobuf:"varint,2,opt,name=gate_node_id,json=gateNodeId,proto3" json:"gate_node_id,omitempty"`
	IpAddr     string `protobuf:"bytes,3,opt,name=ip_addr,json=ipAddr,proto3" json:"ip_addr,omitempty"`
}

func (x *NotifyClientConnect) Reset() {
	*x = NotifyClientConnect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyClientConnect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyClientConnect) ProtoMessage() {}

func (x *NotifyClientConnect) ProtoReflect() protoreflect.Message {
	mi := &file_core_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyClientConnect.ProtoReflect.Descriptor instead.
func (*NotifyClientConnect) Descriptor() ([]byte, []int) {
	return file_core_proto_rawDescGZIP(), []int{1}
}

func (x *NotifyClientConnect) GetUid() uint32 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *NotifyClientConnect) GetGateNodeId() uint32 {
	if x != nil {
		return x.GateNodeId
	}
	return 0
}

func (x *NotifyClientConnect) GetIpAddr() string {
	if x != nil {
		return x.IpAddr
	}
	return ""
}

var File_core_proto protoreflect.FileDescriptor

var file_core_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x6f, 0x70,
	0x65, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x22, 0xf4, 0x01, 0x0a,
	0x16, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x44, 0x69, 0x73,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x20, 0x0a, 0x0c, 0x67, 0x61, 0x74,
	0x65, 0x5f, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x0a, 0x67, 0x61, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x45, 0x0a, 0x06, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2d, 0x2e, 0x6f, 0x70,
	0x65, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x4e, 0x6f, 0x74,
	0x69, 0x66, 0x79, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x2e, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x52, 0x06, 0x72, 0x65, 0x61, 0x73,
	0x6f, 0x6e, 0x22, 0x27, 0x0a, 0x05, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x12, 0x14, 0x0a, 0x10, 0x5f,
	0x5f, 0x69, 0x6e, 0x76, 0x69, 0x6c, 0x69, 0x64, 0x5f, 0x5f, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x10,
	0x00, 0x12, 0x08, 0x0a, 0x02, 0x49, 0x44, 0x10, 0xa1, 0x8d, 0x06, 0x22, 0x36, 0x0a, 0x06, 0x52,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x0a, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x42, 0x65, 0x4b, 0x69, 0x63, 0x6b, 0x65,
	0x64, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x47, 0x61, 0x74, 0x65, 0x43, 0x6c, 0x6f, 0x73, 0x65,
	0x64, 0x10, 0x02, 0x22, 0x8b, 0x01, 0x0a, 0x13, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x43, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x75,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x20, 0x0a,
	0x0c, 0x67, 0x61, 0x74, 0x65, 0x5f, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x0a, 0x67, 0x61, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12,
	0x17, 0x0a, 0x07, 0x69, 0x70, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x69, 0x70, 0x41, 0x64, 0x64, 0x72, 0x22, 0x27, 0x0a, 0x05, 0x4d, 0x53, 0x47, 0x49,
	0x44, 0x12, 0x14, 0x0a, 0x10, 0x5f, 0x5f, 0x69, 0x6e, 0x76, 0x69, 0x6c, 0x69, 0x64, 0x5f, 0x5f,
	0x4d, 0x53, 0x47, 0x49, 0x44, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x02, 0x49, 0x44, 0x10, 0xa2, 0x8d,
	0x06, 0x42, 0x19, 0x5a, 0x08, 0x6d, 0x73, 0x67, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0xaa, 0x02, 0x0c,
	0x73, 0x72, 0x63, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_core_proto_rawDescOnce sync.Once
	file_core_proto_rawDescData = file_core_proto_rawDesc
)

func file_core_proto_rawDescGZIP() []byte {
	file_core_proto_rawDescOnce.Do(func() {
		file_core_proto_rawDescData = protoimpl.X.CompressGZIP(file_core_proto_rawDescData)
	})
	return file_core_proto_rawDescData
}

var file_core_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_core_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_core_proto_goTypes = []any{
	(NotifyClientDisconnect_MSGID)(0),  // 0: openproto.core.NotifyClientDisconnect.MSGID
	(NotifyClientDisconnect_Reason)(0), // 1: openproto.core.NotifyClientDisconnect.Reason
	(NotifyClientConnect_MSGID)(0),     // 2: openproto.core.NotifyClientConnect.MSGID
	(*NotifyClientDisconnect)(nil),     // 3: openproto.core.NotifyClientDisconnect
	(*NotifyClientConnect)(nil),        // 4: openproto.core.NotifyClientConnect
}
var file_core_proto_depIdxs = []int32{
	1, // 0: openproto.core.NotifyClientDisconnect.reason:type_name -> openproto.core.NotifyClientDisconnect.Reason
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_core_proto_init() }
func file_core_proto_init() {
	if File_core_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_core_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*NotifyClientDisconnect); i {
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
		file_core_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*NotifyClientConnect); i {
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
			RawDescriptor: file_core_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_core_proto_goTypes,
		DependencyIndexes: file_core_proto_depIdxs,
		EnumInfos:         file_core_proto_enumTypes,
		MessageInfos:      file_core_proto_msgTypes,
	}.Build()
	File_core_proto = out.File
	file_core_proto_rawDesc = nil
	file_core_proto_goTypes = nil
	file_core_proto_depIdxs = nil
}
