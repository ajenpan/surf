package server

import (
	"encoding/binary"

	"github.com/ajenpan/surf/server/tcp"
)

const (
	// customer
	PacketTypStartAt_ tcp.PacketType = tcp.PacketTypeInnerEndAt_ + iota
	PacketTypRoute
	PacketTypEndAt_
)

const (
	RouteTypAsync RouteMsgTyp = iota
	RouteTypRequest
	RouteTypResponse
	RouteTypRespErr
)

type RouteHead []uint8

const RouteHeadLen = 17

type RouteMsgTyp uint8

func CastRoutHead(head []uint8) (RouteHead, error) {
	if len(head) != RouteHeadLen {
		return nil, tcp.ErrHeadSizeWrong
	}
	return RouteHead(head), nil
}

func NewRoutHead() RouteHead {
	return make([]uint8, RouteHeadLen)
}

func (h RouteHead) GetTargetUID() uint32 {
	return binary.LittleEndian.Uint32(h[0:4])
}

func (h RouteHead) GetSrouceUID() uint32 {
	return binary.LittleEndian.Uint32(h[4:8])
}

func (h RouteHead) GetAskID() uint32 {
	return binary.LittleEndian.Uint32(h[8:12])
}

func (h RouteHead) GetMsgID() uint32 {
	return binary.LittleEndian.Uint32(h[12:16])
}

func (h RouteHead) GetMsgTyp() RouteMsgTyp {
	return RouteMsgTyp(h[16])
}

func (h RouteHead) SetTargetUID(u uint32) {
	binary.LittleEndian.PutUint32(h[0:4], u)
}

func (h RouteHead) SetSrouceUID(u uint32) {
	binary.LittleEndian.PutUint32(h[4:8], u)
}

func (h RouteHead) SetAskID(id uint32) {
	binary.LittleEndian.PutUint32(h[8:12], id)
}

func (h RouteHead) SetMsgID(id uint32) {
	binary.LittleEndian.PutUint32(h[12:16], id)
}

func (h RouteHead) SetMsgTyp(typ RouteMsgTyp) {
	h[16] = uint8(typ)
}

type RouteErrHead RouteHead
