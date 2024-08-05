package network

import (
	"encoding/binary"
	"io"
)

// | pktype | subflag | bodylen | clientid | svrid | svrtype | msgid | syn | body |
// | 1      | 1       | 2       | 4        | 4     | 2       | 4     | 4   | n    |
const PacketHeadLen = 1 + 1 + 2 + 4 + 4 + 2 + 4 + 4
const PacketMaxBodySize uint16 = (0xFFFF - 1)

type Packet_MsgType = uint8

const (
	PacketType_Async     Packet_MsgType = 0
	PacketType_Request   Packet_MsgType = 1
	PacketType_Response  Packet_MsgType = 2
	PacketType_RouteErr  Packet_MsgType = 3
	PacketType_HandleErr Packet_MsgType = 4
	PacketType_Inner     Packet_MsgType = 255
)

type Packet_Inner = uint8

const (
	PacketType_Inner_HandShake       Packet_Inner = 225
	PacketType_Inner_Cmd             Packet_Inner = 226
	PacketType_Inner_CmdResult       Packet_Inner = 227
	PacketType_Inner_HandShakeFinish Packet_Inner = 228
	PacketType_Inner_Heartbeat       Packet_Inner = 229
)

type Packet_RouteErr = uint8

const (
	PacketType_RouteErr_NodeNotFound Packet_RouteErr = 1
)

type Packet_HandleErrCode = uint8

const (
	PacketType_HandleErr_MethodNotFound Packet_HandleErrCode = 1
	PacketType_HandleErr_MethodParseErr Packet_HandleErrCode = 2
)

type PacketHead []uint8

type HVPacket struct {
	Head PacketHead
	Body []uint8
}

func NewHead() PacketHead {
	return make([]uint8, PacketHeadLen)
}

func NewHVPacket() *HVPacket {
	return &HVPacket{
		Head: NewHead(),
	}
}

func (hr *PacketHead) GetType() Packet_MsgType {
	return (*hr)[0]
}

func (hr *PacketHead) SetType(f Packet_MsgType) {
	(*hr)[0] = f
}

func (hr PacketHead) GetSubFlag() uint8 {
	return hr[1]
}

func (hr PacketHead) SetSubFlag(f uint8) {
	hr[1] = f
}

func (hr PacketHead) GetBodyLen() uint16 {
	return uint16(hr[2]) | uint16(hr[3])<<8
}

func (hr PacketHead) SetBodyLen(l uint16) {
	hr[2] = uint8(l)
	hr[3] = uint8(l >> 8)
}

func (r *PacketHead) GetClientId() uint32 {
	return binary.LittleEndian.Uint32((*r)[4:8])
}
func (r PacketHead) SetClientId(d uint32) {
	binary.LittleEndian.PutUint32(r[4:8], d)
}

func (r PacketHead) GetSvrId() uint32 {
	return binary.LittleEndian.Uint32(r[8:12])
}
func (r PacketHead) SetSvrId(d uint32) {
	binary.LittleEndian.PutUint32(r[8:12], d)
}

func (r PacketHead) GetSvrType() uint16 {
	return binary.LittleEndian.Uint16(r[12:14])
}

func (r PacketHead) SetSvrType(d uint16) {
	binary.LittleEndian.PutUint16(r[12:14], d)
}

func (r PacketHead) GetMsgId() uint32 {
	return binary.LittleEndian.Uint32(r[14:18])
}
func (r PacketHead) SetMsgId(d uint32) {
	binary.LittleEndian.PutUint32(r[14:18], d)
}
func (r PacketHead) GetSYN() uint32 {
	return binary.LittleEndian.Uint32(r[18:22])
}

func (r PacketHead) SetSYN(d uint32) {
	binary.LittleEndian.PutUint32(r[18:22], d)
}

func (hr *PacketHead) Reset() {
	(*hr) = NewHead()
}

func (p *HVPacket) PacketType() uint8 {
	return HVPacketType
}

func (p *HVPacket) ReadFrom(reader io.Reader) (int64, error) {
	var err error
	if _, err = io.ReadFull(reader, p.Head); err != nil {
		return 0, err
	}

	bodylen := p.Head.GetBodyLen()

	if bodylen > 0 {
		p.Body = make([]uint8, bodylen)
		_, err = io.ReadFull(reader, p.Body)
		if err != nil {
			return 0, err
		}
	}
	return int64(PacketHeadLen + int(bodylen)), nil
}

func (p *HVPacket) WriteTo(writer io.Writer) (int64, error) {
	_, err := writer.Write(p.Head)
	if err != nil {
		return 0, err
	}

	if len(p.Body) > 0 {
		_, err = writer.Write(p.Body)
		if err != nil {
			return 0, err
		}
	}

	return int64(PacketHeadLen + len(p.Body)), nil
}

func (p *HVPacket) SetBody(b []uint8) {
	p.Body = b
	p.Head.SetBodyLen(uint16(len(b)))
}

func (p *HVPacket) GetBody() []uint8 {
	return p.Body
}

func (p *HVPacket) Reset() {
	p.Head.Reset()
	p.Body = nil
}
