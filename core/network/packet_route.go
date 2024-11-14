package network

import (
	"encoding/binary"
)

// | msgid | syn | client | nodeid | svrtype | errcode | msgtyp | marshal |
// | 4     | 4   | 4      | 4      | 2       | 2       | 1      | 1       |
const RoutePackHeadLen = 4 + 4 + 4 + 4 + 2 + 2 + 1 + 1 // 22

type routePacketHead []uint8

type MyStruct struct {
	Field1 uint32
	Field2 uint16
	Field3 uint8
}

type RoutePackType = uint8

const (
	RoutePackType_SubFlag_RouteFail RoutePackType = 1
)

type RoutePackMsgType = uint8

const (
	RoutePackMsgType_Async    RoutePackMsgType = 0
	RoutePackMsgType_Request  RoutePackMsgType = 1
	RoutePackMsgType_Response RoutePackMsgType = 2
)

// type RoutePackType_SubFlag_RouteErrCode uint16
// const (
// 	RoutePackType_SubFlag_RouteErrCode_NodeNotFound   = 1
// 	RoutePackType_SubFlag_RouteErrCode_MethodNotFound = 2
// 	RoutePackType_SubFlag_RouteErrCode_MethodParseErr = 3
// )

func NewRoutePacket(body []byte) *RoutePacket {
	return &RoutePacket{
		Head: make(routePacketHead, RoutePackHeadLen),
		Body: body,
	}
}

func NewRouteFailedPacket(subtype uint8) *RoutePacket {
	return &RoutePacket{
		subtype: RoutePackType_SubFlag_RouteFail,
		Head:    make(routePacketHead, RoutePackHeadLen),
	}
}

func (routePacketHead) CopyFrom(source routePacketHead) {
	copy(source, source)
}

type RoutePacket struct {
	subtype uint8

	Head routePacketHead
	Body []byte
}

func (r *RoutePacket) FromHVPacket(hv *HVPacket) *RoutePacket {
	r.subtype = hv.Meta.GetSubFlag()
	r.Head = hv.GetHead()
	r.Body = hv.GetBody()
	return r
}

func (r *RoutePacket) ToHVPacket() *HVPacket {
	ret := NewHVPacket()
	ret.Meta.SetType(PacketType_Route)
	ret.Meta.SetSubFlag(r.subtype)
	ret.SetBody(r.Body)
	return ret
}

func (r *RoutePacket) GetSubType() uint8 {
	return r.subtype
}

func (r *RoutePacket) SetSubType(t uint8) {
	r.subtype = t
}

func (r *RoutePacket) GetMsgId() uint32 {
	return binary.LittleEndian.Uint32(r.Head[0:4])
}

func (r *RoutePacket) GetSYN() uint32 {
	return binary.LittleEndian.Uint32(r.Head[4:8])
}

func (r *RoutePacket) GetClientId() uint32 {
	return binary.LittleEndian.Uint32(r.Head[8:12])
}

func (r *RoutePacket) GetNodeId() uint32 {
	return binary.LittleEndian.Uint32(r.Head[8:16])
}

func (r *RoutePacket) GetSvrType() uint16 {
	return binary.LittleEndian.Uint16(r.Head[16:18])
}

func (r *RoutePacket) GetErrCode() int16 {
	return (int16)(binary.LittleEndian.Uint16(r.Head[18:20]))
}

func (r *RoutePacket) GetMsgType() uint8 {
	return r.Head[20]
}

func (r *RoutePacket) GetMarshalType() uint8 {
	return r.Head[21]
}

func (r *RoutePacket) GetBody() []byte {
	return r.Body
}

func (r *RoutePacket) SetMsgId(d uint32) {
	binary.LittleEndian.PutUint32(r.Head[0:4], d)
}

func (r *RoutePacket) SetSYN(d uint32) {
	binary.LittleEndian.PutUint32(r.Head[4:8], d)
}

func (r *RoutePacket) SetClientId(d uint32) {
	binary.LittleEndian.PutUint32(r.Head[8:12], d)
}

func (r *RoutePacket) SetNodeId(d uint32) {
	binary.LittleEndian.PutUint32(r.Head[12:16], d)
}

func (r *RoutePacket) SetSvrType(d uint16) {
	binary.LittleEndian.PutUint16(r.Head[16:18], d)
}

func (r *RoutePacket) SetErrCode(d int16) {
	binary.LittleEndian.PutUint16(r.Head[18:20], uint16(d))
}

func (r *RoutePacket) SetMsgType(d uint8) {
	r.Head[20] = d
}

func (r *RoutePacket) SetMarshalType(d uint8) {
	r.Head[21] = d
}

func (r *RoutePacket) SetBody(d []byte) {
	r.Body = d
}
