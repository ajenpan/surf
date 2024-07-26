// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.18.1
// source: service/battle/proto/battle.proto

package proto

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

type JoinBattleRequest_MSGID int32

const (
	JoinBattleRequest_INVALID_MSGID JoinBattleRequest_MSGID = 0
	JoinBattleRequest_ID            JoinBattleRequest_MSGID = 2000
)

// Enum value maps for JoinBattleRequest_MSGID.
var (
	JoinBattleRequest_MSGID_name = map[int32]string{
		0:    "INVALID_MSGID",
		2000: "ID",
	}
	JoinBattleRequest_MSGID_value = map[string]int32{
		"INVALID_MSGID": 0,
		"ID":            2000,
	}
)

func (x JoinBattleRequest_MSGID) Enum() *JoinBattleRequest_MSGID {
	p := new(JoinBattleRequest_MSGID)
	*p = x
	return p
}

func (x JoinBattleRequest_MSGID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (JoinBattleRequest_MSGID) Descriptor() protoreflect.EnumDescriptor {
	return file_service_battle_proto_battle_proto_enumTypes[0].Descriptor()
}

func (JoinBattleRequest_MSGID) Type() protoreflect.EnumType {
	return &file_service_battle_proto_battle_proto_enumTypes[0]
}

func (x JoinBattleRequest_MSGID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use JoinBattleRequest_MSGID.Descriptor instead.
func (JoinBattleRequest_MSGID) EnumDescriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{0, 0}
}

type JoinBattleResponse_MSGID int32

const (
	JoinBattleResponse_INVALID_MSGID JoinBattleResponse_MSGID = 0
	JoinBattleResponse_ID            JoinBattleResponse_MSGID = 2001
)

// Enum value maps for JoinBattleResponse_MSGID.
var (
	JoinBattleResponse_MSGID_name = map[int32]string{
		0:    "INVALID_MSGID",
		2001: "ID",
	}
	JoinBattleResponse_MSGID_value = map[string]int32{
		"INVALID_MSGID": 0,
		"ID":            2001,
	}
)

func (x JoinBattleResponse_MSGID) Enum() *JoinBattleResponse_MSGID {
	p := new(JoinBattleResponse_MSGID)
	*p = x
	return p
}

func (x JoinBattleResponse_MSGID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (JoinBattleResponse_MSGID) Descriptor() protoreflect.EnumDescriptor {
	return file_service_battle_proto_battle_proto_enumTypes[1].Descriptor()
}

func (JoinBattleResponse_MSGID) Type() protoreflect.EnumType {
	return &file_service_battle_proto_battle_proto_enumTypes[1]
}

func (x JoinBattleResponse_MSGID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use JoinBattleResponse_MSGID.Descriptor instead.
func (JoinBattleResponse_MSGID) EnumDescriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{1, 0}
}

type PlayerReadyRequest_MSGID int32

const (
	PlayerReadyRequest_INVALID_MSGID PlayerReadyRequest_MSGID = 0
	PlayerReadyRequest_ID            PlayerReadyRequest_MSGID = 2002
)

// Enum value maps for PlayerReadyRequest_MSGID.
var (
	PlayerReadyRequest_MSGID_name = map[int32]string{
		0:    "INVALID_MSGID",
		2002: "ID",
	}
	PlayerReadyRequest_MSGID_value = map[string]int32{
		"INVALID_MSGID": 0,
		"ID":            2002,
	}
)

func (x PlayerReadyRequest_MSGID) Enum() *PlayerReadyRequest_MSGID {
	p := new(PlayerReadyRequest_MSGID)
	*p = x
	return p
}

func (x PlayerReadyRequest_MSGID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PlayerReadyRequest_MSGID) Descriptor() protoreflect.EnumDescriptor {
	return file_service_battle_proto_battle_proto_enumTypes[2].Descriptor()
}

func (PlayerReadyRequest_MSGID) Type() protoreflect.EnumType {
	return &file_service_battle_proto_battle_proto_enumTypes[2]
}

func (x PlayerReadyRequest_MSGID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PlayerReadyRequest_MSGID.Descriptor instead.
func (PlayerReadyRequest_MSGID) EnumDescriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{2, 0}
}

type PlayerReadyResonse_MSGID int32

const (
	PlayerReadyResonse_INVALID_MSGID PlayerReadyResonse_MSGID = 0
	PlayerReadyResonse_ID            PlayerReadyResonse_MSGID = 2003
)

// Enum value maps for PlayerReadyResonse_MSGID.
var (
	PlayerReadyResonse_MSGID_name = map[int32]string{
		0:    "INVALID_MSGID",
		2003: "ID",
	}
	PlayerReadyResonse_MSGID_value = map[string]int32{
		"INVALID_MSGID": 0,
		"ID":            2003,
	}
)

func (x PlayerReadyResonse_MSGID) Enum() *PlayerReadyResonse_MSGID {
	p := new(PlayerReadyResonse_MSGID)
	*p = x
	return p
}

func (x PlayerReadyResonse_MSGID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PlayerReadyResonse_MSGID) Descriptor() protoreflect.EnumDescriptor {
	return file_service_battle_proto_battle_proto_enumTypes[3].Descriptor()
}

func (PlayerReadyResonse_MSGID) Type() protoreflect.EnumType {
	return &file_service_battle_proto_battle_proto_enumTypes[3]
}

func (x PlayerReadyResonse_MSGID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PlayerReadyResonse_MSGID.Descriptor instead.
func (PlayerReadyResonse_MSGID) EnumDescriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{3, 0}
}

type LoigcMessageWrap_MSGID int32

const (
	LoigcMessageWrap_INVALID_MSGID LoigcMessageWrap_MSGID = 0
	LoigcMessageWrap_ID            LoigcMessageWrap_MSGID = 2004
)

// Enum value maps for LoigcMessageWrap_MSGID.
var (
	LoigcMessageWrap_MSGID_name = map[int32]string{
		0:    "INVALID_MSGID",
		2004: "ID",
	}
	LoigcMessageWrap_MSGID_value = map[string]int32{
		"INVALID_MSGID": 0,
		"ID":            2004,
	}
)

func (x LoigcMessageWrap_MSGID) Enum() *LoigcMessageWrap_MSGID {
	p := new(LoigcMessageWrap_MSGID)
	*p = x
	return p
}

func (x LoigcMessageWrap_MSGID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LoigcMessageWrap_MSGID) Descriptor() protoreflect.EnumDescriptor {
	return file_service_battle_proto_battle_proto_enumTypes[4].Descriptor()
}

func (LoigcMessageWrap_MSGID) Type() protoreflect.EnumType {
	return &file_service_battle_proto_battle_proto_enumTypes[4]
}

func (x LoigcMessageWrap_MSGID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LoigcMessageWrap_MSGID.Descriptor instead.
func (LoigcMessageWrap_MSGID) EnumDescriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{4, 0}
}

type JoinBattleRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BattleId   string `protobuf:"bytes,1,opt,name=battle_id,json=battleId,proto3" json:"battle_id,omitempty"`
	SeatId     uint32 `protobuf:"varint,2,opt,name=seat_id,json=seatId,proto3" json:"seat_id,omitempty"`
	ReadyState int32  `protobuf:"varint,3,opt,name=ready_state,json=readyState,proto3" json:"ready_state,omitempty"`
}

func (x *JoinBattleRequest) Reset() {
	*x = JoinBattleRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_battle_proto_battle_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinBattleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinBattleRequest) ProtoMessage() {}

func (x *JoinBattleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_battle_proto_battle_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinBattleRequest.ProtoReflect.Descriptor instead.
func (*JoinBattleRequest) Descriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{0}
}

func (x *JoinBattleRequest) GetBattleId() string {
	if x != nil {
		return x.BattleId
	}
	return ""
}

func (x *JoinBattleRequest) GetSeatId() uint32 {
	if x != nil {
		return x.SeatId
	}
	return 0
}

func (x *JoinBattleRequest) GetReadyState() int32 {
	if x != nil {
		return x.ReadyState
	}
	return 0
}

type JoinBattleResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BattleId   string `protobuf:"bytes,1,opt,name=battle_id,json=battleId,proto3" json:"battle_id,omitempty"`
	SeatId     uint32 `protobuf:"varint,2,opt,name=seat_id,json=seatId,proto3" json:"seat_id,omitempty"`
	ReadyState int32  `protobuf:"varint,3,opt,name=ready_state,json=readyState,proto3" json:"ready_state,omitempty"`
}

func (x *JoinBattleResponse) Reset() {
	*x = JoinBattleResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_battle_proto_battle_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinBattleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinBattleResponse) ProtoMessage() {}

func (x *JoinBattleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_battle_proto_battle_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinBattleResponse.ProtoReflect.Descriptor instead.
func (*JoinBattleResponse) Descriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{1}
}

func (x *JoinBattleResponse) GetBattleId() string {
	if x != nil {
		return x.BattleId
	}
	return ""
}

func (x *JoinBattleResponse) GetSeatId() uint32 {
	if x != nil {
		return x.SeatId
	}
	return 0
}

func (x *JoinBattleResponse) GetReadyState() int32 {
	if x != nil {
		return x.ReadyState
	}
	return 0
}

type PlayerReadyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BattleId   string `protobuf:"bytes,1,opt,name=battle_id,json=battleId,proto3" json:"battle_id,omitempty"`
	ReadyState int32  `protobuf:"varint,2,opt,name=ready_state,json=readyState,proto3" json:"ready_state,omitempty"`
}

func (x *PlayerReadyRequest) Reset() {
	*x = PlayerReadyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_battle_proto_battle_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PlayerReadyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlayerReadyRequest) ProtoMessage() {}

func (x *PlayerReadyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_battle_proto_battle_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlayerReadyRequest.ProtoReflect.Descriptor instead.
func (*PlayerReadyRequest) Descriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{2}
}

func (x *PlayerReadyRequest) GetBattleId() string {
	if x != nil {
		return x.BattleId
	}
	return ""
}

func (x *PlayerReadyRequest) GetReadyState() int32 {
	if x != nil {
		return x.ReadyState
	}
	return 0
}

type PlayerReadyResonse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReadyState int32 `protobuf:"varint,1,opt,name=ready_state,json=readyState,proto3" json:"ready_state,omitempty"`
}

func (x *PlayerReadyResonse) Reset() {
	*x = PlayerReadyResonse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_battle_proto_battle_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PlayerReadyResonse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlayerReadyResonse) ProtoMessage() {}

func (x *PlayerReadyResonse) ProtoReflect() protoreflect.Message {
	mi := &file_service_battle_proto_battle_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlayerReadyResonse.ProtoReflect.Descriptor instead.
func (*PlayerReadyResonse) Descriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{3}
}

func (x *PlayerReadyResonse) GetReadyState() int32 {
	if x != nil {
		return x.ReadyState
	}
	return 0
}

type LoigcMessageWrap struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BattleId string `protobuf:"bytes,1,opt,name=battle_id,json=battleId,proto3" json:"battle_id,omitempty"`
	Msgid    uint32 `protobuf:"varint,2,opt,name=msgid,proto3" json:"msgid,omitempty"`
	Data     []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *LoigcMessageWrap) Reset() {
	*x = LoigcMessageWrap{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_battle_proto_battle_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LoigcMessageWrap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LoigcMessageWrap) ProtoMessage() {}

func (x *LoigcMessageWrap) ProtoReflect() protoreflect.Message {
	mi := &file_service_battle_proto_battle_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LoigcMessageWrap.ProtoReflect.Descriptor instead.
func (*LoigcMessageWrap) Descriptor() ([]byte, []int) {
	return file_service_battle_proto_battle_proto_rawDescGZIP(), []int{4}
}

func (x *LoigcMessageWrap) GetBattleId() string {
	if x != nil {
		return x.BattleId
	}
	return ""
}

func (x *LoigcMessageWrap) GetMsgid() uint32 {
	if x != nil {
		return x.Msgid
	}
	return 0
}

func (x *LoigcMessageWrap) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_service_battle_proto_battle_proto protoreflect.FileDescriptor

var file_service_battle_proto_battle_proto_rawDesc = []byte{
	0x0a, 0x21, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x06, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x22, 0x8f, 0x01, 0x0a, 0x11,
	0x4a, 0x6f, 0x69, 0x6e, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1b, 0x0a, 0x09, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x49, 0x64, 0x12, 0x17,
	0x0a, 0x07, 0x73, 0x65, 0x61, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x06, 0x73, 0x65, 0x61, 0x74, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x61, 0x64, 0x79,
	0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72, 0x65,
	0x61, 0x64, 0x79, 0x53, 0x74, 0x61, 0x74, 0x65, 0x22, 0x23, 0x0a, 0x05, 0x4d, 0x53, 0x47, 0x49,
	0x44, 0x12, 0x11, 0x0a, 0x0d, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x4d, 0x53, 0x47,
	0x49, 0x44, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x02, 0x49, 0x44, 0x10, 0xd0, 0x0f, 0x22, 0x90, 0x01,
	0x0a, 0x12, 0x4a, 0x6f, 0x69, 0x6e, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x49,
	0x64, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x65, 0x61, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x06, 0x73, 0x65, 0x61, 0x74, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65,
	0x61, 0x64, 0x79, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0a, 0x72, 0x65, 0x61, 0x64, 0x79, 0x53, 0x74, 0x61, 0x74, 0x65, 0x22, 0x23, 0x0a, 0x05, 0x4d,
	0x53, 0x47, 0x49, 0x44, 0x12, 0x11, 0x0a, 0x0d, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f,
	0x4d, 0x53, 0x47, 0x49, 0x44, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x02, 0x49, 0x44, 0x10, 0xd1, 0x0f,
	0x22, 0x77, 0x0a, 0x12, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x52, 0x65, 0x61, 0x64, 0x79, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x62, 0x61, 0x74, 0x74, 0x6c,
	0x65, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x61, 0x64, 0x79, 0x5f, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72, 0x65, 0x61, 0x64, 0x79, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x22, 0x23, 0x0a, 0x05, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x12, 0x11, 0x0a,
	0x0d, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x10, 0x00,
	0x12, 0x07, 0x0a, 0x02, 0x49, 0x44, 0x10, 0xd2, 0x0f, 0x22, 0x5a, 0x0a, 0x12, 0x50, 0x6c, 0x61,
	0x79, 0x65, 0x72, 0x52, 0x65, 0x61, 0x64, 0x79, 0x52, 0x65, 0x73, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x61, 0x64, 0x79, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72, 0x65, 0x61, 0x64, 0x79, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x22, 0x23, 0x0a, 0x05, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x12, 0x11, 0x0a, 0x0d, 0x49, 0x4e, 0x56,
	0x41, 0x4c, 0x49, 0x44, 0x5f, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x02,
	0x49, 0x44, 0x10, 0xd3, 0x0f, 0x22, 0x7e, 0x0a, 0x10, 0x4c, 0x6f, 0x69, 0x67, 0x63, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x57, 0x72, 0x61, 0x70, 0x12, 0x1b, 0x0a, 0x09, 0x62, 0x61, 0x74,
	0x74, 0x6c, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x62, 0x61,
	0x74, 0x74, 0x6c, 0x65, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x73, 0x67, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x6d, 0x73, 0x67, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x22, 0x23, 0x0a, 0x05, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x12, 0x11, 0x0a, 0x0d, 0x49, 0x4e, 0x56,
	0x41, 0x4c, 0x49, 0x44, 0x5f, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x02,
	0x49, 0x44, 0x10, 0xd4, 0x0f, 0x42, 0x16, 0x5a, 0x14, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2f, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_service_battle_proto_battle_proto_rawDescOnce sync.Once
	file_service_battle_proto_battle_proto_rawDescData = file_service_battle_proto_battle_proto_rawDesc
)

func file_service_battle_proto_battle_proto_rawDescGZIP() []byte {
	file_service_battle_proto_battle_proto_rawDescOnce.Do(func() {
		file_service_battle_proto_battle_proto_rawDescData = protoimpl.X.CompressGZIP(file_service_battle_proto_battle_proto_rawDescData)
	})
	return file_service_battle_proto_battle_proto_rawDescData
}

var file_service_battle_proto_battle_proto_enumTypes = make([]protoimpl.EnumInfo, 5)
var file_service_battle_proto_battle_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_service_battle_proto_battle_proto_goTypes = []interface{}{
	(JoinBattleRequest_MSGID)(0),  // 0: battle.JoinBattleRequest.MSGID
	(JoinBattleResponse_MSGID)(0), // 1: battle.JoinBattleResponse.MSGID
	(PlayerReadyRequest_MSGID)(0), // 2: battle.PlayerReadyRequest.MSGID
	(PlayerReadyResonse_MSGID)(0), // 3: battle.PlayerReadyResonse.MSGID
	(LoigcMessageWrap_MSGID)(0),   // 4: battle.LoigcMessageWrap.MSGID
	(*JoinBattleRequest)(nil),     // 5: battle.JoinBattleRequest
	(*JoinBattleResponse)(nil),    // 6: battle.JoinBattleResponse
	(*PlayerReadyRequest)(nil),    // 7: battle.PlayerReadyRequest
	(*PlayerReadyResonse)(nil),    // 8: battle.PlayerReadyResonse
	(*LoigcMessageWrap)(nil),      // 9: battle.LoigcMessageWrap
}
var file_service_battle_proto_battle_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_service_battle_proto_battle_proto_init() }
func file_service_battle_proto_battle_proto_init() {
	if File_service_battle_proto_battle_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_service_battle_proto_battle_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinBattleRequest); i {
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
		file_service_battle_proto_battle_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinBattleResponse); i {
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
		file_service_battle_proto_battle_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PlayerReadyRequest); i {
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
		file_service_battle_proto_battle_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PlayerReadyResonse); i {
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
		file_service_battle_proto_battle_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LoigcMessageWrap); i {
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
			RawDescriptor: file_service_battle_proto_battle_proto_rawDesc,
			NumEnums:      5,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_service_battle_proto_battle_proto_goTypes,
		DependencyIndexes: file_service_battle_proto_battle_proto_depIdxs,
		EnumInfos:         file_service_battle_proto_battle_proto_enumTypes,
		MessageInfos:      file_service_battle_proto_battle_proto_msgTypes,
	}.Build()
	File_service_battle_proto_battle_proto = out.File
	file_service_battle_proto_battle_proto_rawDesc = nil
	file_service_battle_proto_battle_proto_goTypes = nil
	file_service_battle_proto_battle_proto_depIdxs = nil
}
