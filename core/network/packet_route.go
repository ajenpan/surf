package network

import "encoding/binary"

// | client | nodeid | msgid | syn | svrtype | marshal |errcode | body |
// | 4      | 4      | 4     | 4   | 2       | 2       |2       | n    |
const RoutePackHeadLen = 4 + 4 + 4 + 4 + 2 + 2 + 2 // 24

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

func NewRoutePacket(subflag uint8, head, body []uint8) *HVPacket {
	ret := NewHVPacket()
	ret.SetBody(body)
	ret.SetHead(head)
	ret.Meta.SetSubFlag(subflag)
	ret.Meta.SetType(PacketType_Route)
	return ret
}

func NewRoutePacketHead() RoutePacketHead {
	return make(RoutePacketHead, RoutePackHeadLen)
}

func ParseRoutePacket(pk *HVPacket) *RoutePacket {
	if pk.Meta.GetType() != PacketType_Route || pk.Meta.GetHeadLen() != RoutePackHeadLen || len(pk.Body) < RoutePackHeadLen {
		return nil
	}

	return &RoutePacket{
		RoutePacketHead: pk.GetHead(),
		Body:            pk.GetBody(),
	}
}

type RoutePacketHead []uint8

type RoutePacket struct {
	RoutePacketHead
	Body []byte
}

func (r RoutePacketHead) GetClientId() uint32 {
	return binary.LittleEndian.Uint32(r[0:4])
}

func (r RoutePacketHead) GetNodeId() uint32 {
	return binary.LittleEndian.Uint32(r[4:8])
}

func (r RoutePacketHead) GetMsgId() uint32 {
	return binary.LittleEndian.Uint32(r[8:12])
}

func (r RoutePacketHead) GetSYN() uint32 {
	return binary.LittleEndian.Uint32(r[12:16])
}

func (r RoutePacketHead) GetSvrType() uint16 {
	return binary.LittleEndian.Uint16(r[16:18])
}

func (r RoutePacketHead) GetMarshalType() uint16 {
	return binary.LittleEndian.Uint16(r[18:20])
}

func (r RoutePacketHead) GetErrCode() int16 {
	return int16(binary.LittleEndian.Uint16(r[20:22]))
}

func (r RoutePacketHead) GetHead() []byte {
	return r[:RoutePackHeadLen]
}

func (r RoutePacketHead) GetBody() []byte {
	return r[RoutePackHeadLen:]
}

func (r RoutePacketHead) SetClientId(d uint32) {
	binary.LittleEndian.PutUint32(r[0:4], d)
}

func (r RoutePacketHead) SetNodeId(d uint32) {
	binary.LittleEndian.PutUint32(r[4:8], d)
}

func (r RoutePacketHead) SetMsgId(d uint32) {
	binary.LittleEndian.PutUint32(r[8:12], d)
}

func (r RoutePacketHead) SetSYN(d uint32) {
	binary.LittleEndian.PutUint32(r[12:16], d)
}

func (r RoutePacketHead) SetSvrType(d uint16) {
	binary.LittleEndian.PutUint16(r[16:18], d)
}

func (r RoutePacketHead) SetMarshalType(d int16) {
	binary.LittleEndian.PutUint16(r[18:20], uint16(d))
}

func (r RoutePacketHead) SetErrCode(d int16) {
	binary.LittleEndian.PutUint16(r[20:22], uint16(d))
}

func (r RoutePacketHead) GenHVPacket(subflag uint8) *HVPacket {
	ret := NewHVPacket()
	ret.Meta.SetType(PacketType_Route)
	ret.Meta.SetSubFlag(subflag)
	return ret
}
