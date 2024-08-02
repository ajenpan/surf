package network

import "encoding/binary"

type RoutePacketRaw []byte

// | client | server | nodeid | msgid | syn | errcode | msgtype | ttl | body |
// | 4      | 4      | 4      | 4     | 4   | 2       | 1       | 1   | n    |
const RoutePackHeadLen = 4 + 4 + 4 + 4 + 4 + 2 + 1 + 1 // 24

type RouteMsgType uint8

const (
	RouteMsgType_Async     = 0
	RouteMsgType_Request   = 1
	RouteMsgType_Response  = 2
	RouteMsgType_RouteErr  = 3
	RouteMsgType_HandleErr = 4
)

type RouteMsgErrCode uint16

const (
	RouteMsgErrCode_NodeNotFound = 1
)

type RouteHandleErrCode uint16

const (
	RouteHandleErrCode_MethodNotFound = 1
	RouteHandleErrCode_MethodParseErr = 1
)

func (r RoutePacketRaw) GetClientId() uint32 {
	return binary.LittleEndian.Uint32(r[0:4])
}

func (r RoutePacketRaw) GetServerId() uint32 {
	return binary.LittleEndian.Uint32(r[4:8])
}

func (r RoutePacketRaw) GetNodeId() uint32 {
	return binary.LittleEndian.Uint32(r[8:12])
}

func (r RoutePacketRaw) GetMsgId() uint32 {
	return binary.LittleEndian.Uint32(r[12:16])
}

func (r RoutePacketRaw) GetSYN() uint32 {
	return binary.LittleEndian.Uint32(r[16:20])
}

func (r RoutePacketRaw) GetErrCode() int16 {
	return (int16)(binary.LittleEndian.Uint16(r[20:22]))
}

func (r RoutePacketRaw) GetMsgType() uint8 {
	return r[23]
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

func (r RoutePacketRaw) SetServerId(d uint32) {
	binary.LittleEndian.PutUint32(r[4:8], d)
}

func (r RoutePacketRaw) SetNodeId(d uint32) {
	binary.LittleEndian.PutUint32(r[8:12], d)
}

func (r RoutePacketRaw) SetMsgId(d uint32) {
	binary.LittleEndian.PutUint32(r[12:16], d)
}

func (r RoutePacketRaw) SetSYN(d uint32) {
	binary.LittleEndian.PutUint32(r[16:20], d)
}

func (r RoutePacketRaw) SetErrCode(d int16) {
	binary.LittleEndian.PutUint16(r[20:22], uint16(d))
}

func (r RoutePacketRaw) SetMsgType(d uint8) {
	r[23] = d
}
