package network

import "encoding/binary"

type RoutePacketRaw []byte

// | client | nodeid | msgid | syn | svrtype | errcode | body |
// | 4      | 4      | 4     | 4   | 2       | 2       | n    |
const RoutePackHeadLen = 4 + 2 + 4 + 4 + 4 + 2 // 20

type RoutePackType = uint8

const (
	RoutePackType_SubFlag_Async    RoutePackType = 0
	RoutePackType_SubFlag_Request  RoutePackType = 1
	RoutePackType_SubFlag_Response RoutePackType = 2
	RoutePackType_SubFlag_RouteErr RoutePackType = 3
)

type RoutePackType_SubFlag_RouteErrCode uint16

const (
	RoutePackType_SubFlag_RouteErrCode_NodeNotFound   = 1
	RoutePackType_SubFlag_RouteErrCode_MethodNotFound = 2
	RoutePackType_SubFlag_RouteErrCode_MethodParseErr = 3
)

func (r RoutePacketRaw) GetClientId() uint32 {
	return binary.LittleEndian.Uint32(r[0:4])
}

func (r RoutePacketRaw) GetNodeId() uint32 {
	return binary.LittleEndian.Uint32(r[4:8])
}

func (r RoutePacketRaw) GetMsgId() uint32 {
	return binary.LittleEndian.Uint32(r[8:12])
}

func (r RoutePacketRaw) GetSYN() uint32 {
	return binary.LittleEndian.Uint32(r[12:16])
}

func (r RoutePacketRaw) GetSvrType() uint32 {
	return binary.LittleEndian.Uint32(r[16:18])
}

func (r RoutePacketRaw) GetErrCode() int16 {
	return (int16)(binary.LittleEndian.Uint16(r[18:20]))
}

func (r RoutePacketRaw) GetHead() []byte {
	return r[:RoutePackHeadLen]
}

func (r RoutePacketRaw) GetBody() []byte {
	return r[RoutePackHeadLen:]
}

func (r RoutePacketRaw) SetClientId(d uint32) {
	binary.LittleEndian.PutUint32(r[0:4], d)
}

func (r RoutePacketRaw) SetNodeId(d uint32) {
	binary.LittleEndian.PutUint32(r[4:8], d)
}

func (r RoutePacketRaw) SetMsgId(d uint32) {
	binary.LittleEndian.PutUint32(r[8:12], d)
}

func (r RoutePacketRaw) SetSYN(d uint32) {
	binary.LittleEndian.PutUint32(r[12:16], d)
}

func (r RoutePacketRaw) SetSvrType(d uint32) {
	binary.LittleEndian.PutUint32(r[16:18], d)
}

func (r RoutePacketRaw) SetErrCode(d int16) {
	binary.LittleEndian.PutUint16(r[18:20], uint16(d))
}

func (r RoutePacketRaw) GenHVPacket(subflag uint8) *HVPacket {
	ret := NewHVPacket()
	ret.Meta.SetType(PacketType_Route)
	ret.Meta.SetSubFlag(subflag)
	return ret
}
