// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v4.23.4
// source: route.proto

package route

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

type ReqLinkSession struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Target int32 `protobuf:"varint,1,opt,name=target,proto3" json:"target,omitempty"`
}

func (x *ReqLinkSession) Reset() {
	*x = ReqLinkSession{}
	if protoimpl.UnsafeEnabled {
		mi := &file_route_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqLinkSession) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqLinkSession) ProtoMessage() {}

func (x *ReqLinkSession) ProtoReflect() protoreflect.Message {
	mi := &file_route_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqLinkSession.ProtoReflect.Descriptor instead.
func (*ReqLinkSession) Descriptor() ([]byte, []int) {
	return file_route_proto_rawDescGZIP(), []int{0}
}

func (x *ReqLinkSession) GetTarget() int32 {
	if x != nil {
		return x.Target
	}
	return 0
}

type RespLinkSession struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Flag   int32 `protobuf:"varint,1,opt,name=flag,proto3" json:"flag,omitempty"`
	Target int32 `protobuf:"varint,2,opt,name=target,proto3" json:"target,omitempty"`
}

func (x *RespLinkSession) Reset() {
	*x = RespLinkSession{}
	if protoimpl.UnsafeEnabled {
		mi := &file_route_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RespLinkSession) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RespLinkSession) ProtoMessage() {}

func (x *RespLinkSession) ProtoReflect() protoreflect.Message {
	mi := &file_route_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RespLinkSession.ProtoReflect.Descriptor instead.
func (*RespLinkSession) Descriptor() ([]byte, []int) {
	return file_route_proto_rawDescGZIP(), []int{1}
}

func (x *RespLinkSession) GetFlag() int32 {
	if x != nil {
		return x.Flag
	}
	return 0
}

func (x *RespLinkSession) GetTarget() int32 {
	if x != nil {
		return x.Target
	}
	return 0
}

type NotifyLinkStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid    uint32 `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Status int32  `protobuf:"varint,2,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *NotifyLinkStatus) Reset() {
	*x = NotifyLinkStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_route_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyLinkStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyLinkStatus) ProtoMessage() {}

func (x *NotifyLinkStatus) ProtoReflect() protoreflect.Message {
	mi := &file_route_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyLinkStatus.ProtoReflect.Descriptor instead.
func (*NotifyLinkStatus) Descriptor() ([]byte, []int) {
	return file_route_proto_rawDescGZIP(), []int{2}
}

func (x *NotifyLinkStatus) GetUid() uint32 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *NotifyLinkStatus) GetStatus() int32 {
	if x != nil {
		return x.Status
	}
	return 0
}

var File_route_proto protoreflect.FileDescriptor

var file_route_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x73,
	0x75, 0x72, 0x66, 0x22, 0x28, 0x0a, 0x0e, 0x52, 0x65, 0x71, 0x4c, 0x69, 0x6e, 0x6b, 0x53, 0x65,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x22, 0x3d, 0x0a,
	0x0f, 0x52, 0x65, 0x73, 0x70, 0x4c, 0x69, 0x6e, 0x6b, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x12, 0x0a, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04,
	0x66, 0x6c, 0x61, 0x67, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x22, 0x3c, 0x0a, 0x10,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x4c, 0x69, 0x6e, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x75,
	0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x1d, 0x5a, 0x0c, 0x2f, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x3b, 0x72, 0x6f, 0x75, 0x74, 0x65, 0xaa, 0x02, 0x0c, 0x73, 0x72, 0x63,
	0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x73, 0x75, 0x72, 0x66, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_route_proto_rawDescOnce sync.Once
	file_route_proto_rawDescData = file_route_proto_rawDesc
)

func file_route_proto_rawDescGZIP() []byte {
	file_route_proto_rawDescOnce.Do(func() {
		file_route_proto_rawDescData = protoimpl.X.CompressGZIP(file_route_proto_rawDescData)
	})
	return file_route_proto_rawDescData
}

var file_route_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_route_proto_goTypes = []interface{}{
	(*ReqLinkSession)(nil),   // 0: surf.ReqLinkSession
	(*RespLinkSession)(nil),  // 1: surf.RespLinkSession
	(*NotifyLinkStatus)(nil), // 2: surf.NotifyLinkStatus
}
var file_route_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_route_proto_init() }
func file_route_proto_init() {
	if File_route_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_route_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqLinkSession); i {
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
		file_route_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RespLinkSession); i {
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
		file_route_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyLinkStatus); i {
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
			RawDescriptor: file_route_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_route_proto_goTypes,
		DependencyIndexes: file_route_proto_depIdxs,
		MessageInfos:      file_route_proto_msgTypes,
	}.Build()
	File_route_proto = out.File
	file_route_proto_rawDesc = nil
	file_route_proto_goTypes = nil
	file_route_proto_depIdxs = nil
}
