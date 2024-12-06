package core

import (
	"encoding/binary"
	"errors"

	"github.com/ajenpan/surf/core/network"
)

// | msgid | syn | from_uid | to_uid | from_urole | to_urole | errcode | msgtyp | marshal |
// | 4     | 4   | 4        | 4      | 2          | 2        | 2       | 1      | 1       |
const RoutePackHeadBytesLen = 4 + 4 + 4 + 4 + 2 + 2 + 2 + 1 + 1 // 24

var ErrInvalidRoutePacketHeadBytesLen = errors.New("invalid route packet head bytes length")

const (
	RoutePackType_SubFlag_RouteFail uint8 = 1
)

const (
	RoutePackMsgType_Async    = 0
	RoutePackMsgType_Request  = 1
	RoutePackMsgType_Response = 2
)

type routePacketHeadBytes []uint8

func NewRoutePacket(body []byte) *RoutePacket {
	return &RoutePacket{
		Head: make(routePacketHeadBytes, RoutePackHeadBytesLen),
		Body: body,
	}
}

func NewRouteFailedPacket(subtype uint8) *RoutePacket {
	return &RoutePacket{
		subtype: RoutePackType_SubFlag_RouteFail,
		Head:    make(routePacketHeadBytes, RoutePackHeadBytesLen),
	}
}

func (dst routePacketHeadBytes) CopyFrom(src routePacketHeadBytes) {
	copy(dst, src)
}

type RoutePacketHead struct {
	MsgId       uint32 `json:"msgid"`
	SYN         uint32 `json:"syn"`
	FromUID     uint32 `json:"from_uid"`
	ToUID       uint32 `json:"to_uid"`
	FromURole   uint16 `json:"from_urole"`
	ToURole     uint16 `json:"to_urole"`
	ErrCode     int16  `json:"errcode"`
	MsgType     uint8  `json:"msg_type"`
	MarshalType uint8  `json:"marshal"`
}

func (r *RoutePacketHead) FromBytes(b []byte) error {
	if len(b) != RoutePackHeadBytesLen {
		return ErrInvalidRoutePacketHeadBytesLen
	}
	r.MsgId = binary.LittleEndian.Uint32(b[0:4])
	r.SYN = binary.LittleEndian.Uint32(b[4:8])
	r.FromUID = binary.LittleEndian.Uint32(b[8:12])
	r.ToUID = binary.LittleEndian.Uint32(b[12:16])
	r.FromURole = binary.LittleEndian.Uint16(b[16:18])
	r.ToURole = binary.LittleEndian.Uint16(b[18:20])
	r.ErrCode = (int16)(binary.LittleEndian.Uint16(b[20:22]))
	r.MsgType = b[22]
	r.MarshalType = b[23]
	return nil
}

func (r *RoutePacketHead) ToBytes() []byte {
	buf := make([]byte, RoutePackHeadBytesLen)
	binary.LittleEndian.PutUint32(buf[0:4], r.MsgId)
	binary.LittleEndian.PutUint32(buf[4:8], r.SYN)
	binary.LittleEndian.PutUint32(buf[8:12], r.FromUID)
	binary.LittleEndian.PutUint32(buf[12:16], r.ToUID)
	binary.LittleEndian.PutUint16(buf[16:18], r.FromURole)
	binary.LittleEndian.PutUint16(buf[18:20], r.ToURole)
	binary.LittleEndian.PutUint16(buf[20:22], uint16(r.ErrCode))
	buf[22] = r.MsgType
	buf[23] = r.MarshalType
	return buf
}

type RoutePacket struct {
	subtype uint8

	Head routePacketHeadBytes
	Body []byte
}

func (r *RoutePacket) FromHVPacket(hv *network.HVPacket) *RoutePacket {
	r.subtype = hv.Meta.GetSubFlag()
	r.Head = hv.GetHead()
	if len(r.Head) != RoutePackHeadBytesLen {
		return nil
	}
	r.Body = hv.GetBody()
	return r
}

func (r *RoutePacket) ToHVPacket() *network.HVPacket {
	ret := network.NewHVPacket()
	ret.Meta.SetType(network.PacketType_Route)
	ret.Meta.SetSubFlag(r.subtype)
	ret.SetHead(r.Head)
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

func (r *RoutePacket) GetFromUID() uint32 {
	return binary.LittleEndian.Uint32(r.Head[8:12])
}

func (r *RoutePacket) GetToUID() uint32 {
	return binary.LittleEndian.Uint32(r.Head[12:16])
}

func (r *RoutePacket) GetFromURole() uint16 {
	return binary.LittleEndian.Uint16(r.Head[16:18])
}

func (r *RoutePacket) GetToURole() uint16 {
	return binary.LittleEndian.Uint16(r.Head[18:20])
}

func (r *RoutePacket) GetErrCode() int16 {
	return (int16)(binary.LittleEndian.Uint16(r.Head[20:22]))
}

func (r *RoutePacket) GetMsgType() uint8 {
	return r.Head[22]
}

func (r *RoutePacket) GetMarshalType() uint8 {
	return r.Head[23]
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

func (r *RoutePacket) SetFromUID(d uint32) {
	binary.LittleEndian.PutUint32(r.Head[8:12], d)
}

func (r *RoutePacket) SetToUID(d uint32) {
	binary.LittleEndian.PutUint32(r.Head[12:16], d)
}

func (r *RoutePacket) SetFromURole(d uint16) {
	binary.LittleEndian.PutUint16(r.Head[16:18], d)
}

func (r *RoutePacket) SetToURole(d uint16) {
	binary.LittleEndian.PutUint16(r.Head[18:20], d)
}

func (r *RoutePacket) SetErrCode(d int16) {
	binary.LittleEndian.PutUint16(r.Head[20:22], uint16(d))
}

func (r *RoutePacket) SetMsgType(d uint8) {
	r.Head[22] = d
}

func (r *RoutePacket) SetMarshalType(d uint8) {
	r.Head[23] = d
}

func (r *RoutePacket) SetBody(d []byte) {
	r.Body = d
}
