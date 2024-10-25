package network

import (
	"io"
)

// | Packet Meta                          | Packet Body       |
// | pktype | subtype | headlen | bodylen | head    | body    |
// | 1      | 1       | 1       | 2       | headlen | bodylen |
const PacketMetaLen = 1 + 1 + 1 + 2

const PacketMaxBodySize uint16 = (0xFFFF - 1)
const PacketMaxHeadSize uint8 = (0xFF - 1)

type Packet_MsgType uint8

const (
	PacketType_Route     Packet_MsgType = 1
	PacketType_RouteNode Packet_MsgType = 254
	PacketType_Inner     Packet_MsgType = 255
)

type PacketInnerSubType = uint8

const (
	PacketInnerSubType_HandShakeStart  PacketInnerSubType = 5
	PacketInnerSubType_Cmd             PacketInnerSubType = 6
	PacketInnerSubType_CmdResult       PacketInnerSubType = 7
	PacketInnerSubType_HandShakeFailed PacketInnerSubType = 8
	PacketInnerSubType_HandShakeFinish PacketInnerSubType = 9
	PacketInnerSubType_Heartbeat       PacketInnerSubType = 10
)

type PacketMeta []uint8

type HVPacket struct {
	Meta PacketMeta
	Head []uint8
	Body []uint8
}

func NewMeta() PacketMeta {
	return make([]uint8, PacketMetaLen)
}

func NewHVPacket() *HVPacket {
	return &HVPacket{
		Meta: NewMeta(),
		Head: nil,
		Body: nil,
	}
}

func (hr *PacketMeta) GetType() Packet_MsgType {
	return Packet_MsgType((*hr)[0])
}

func (hr *PacketMeta) SetType(f Packet_MsgType) {
	(*hr)[0] = uint8(f)
}

func (hr *PacketMeta) GetSubFlag() uint8 {
	return (*hr)[1]
}

func (hr *PacketMeta) SetSubFlag(f uint8) {
	(*hr)[1] = f
}

func (hr *PacketMeta) GetHeadLen() uint8 {
	return (*hr)[2]
}

func (hr *PacketMeta) SetHeadLen(l uint8) {
	(*hr)[2] = (l)
}

func (hr *PacketMeta) GetBodyLen() uint16 {
	return uint16((*hr)[3]) | uint16((*hr)[4])<<8
}

func (hr *PacketMeta) SetBodyLen(l uint16) {
	(*hr)[3] = uint8(l)
	(*hr)[4] = uint8(l >> 8)
}

func (hr *PacketMeta) Reset() {
	(*hr) = (*hr)[:]
}

func (p *HVPacket) ReadFrom(reader io.Reader) (int64, error) {
	var err error
	if _, err = io.ReadFull(reader, p.Meta); err != nil {
		return 0, err
	}

	headlen := p.Meta.GetHeadLen()
	if headlen > 0 {
		p.Head = make([]uint8, headlen)
		_, err = io.ReadFull(reader, p.Head)
		if err != nil {
			return 0, err
		}
	}

	bodylen := p.Meta.GetBodyLen()
	if bodylen > 0 {
		p.Body = make([]uint8, bodylen)
		_, err = io.ReadFull(reader, p.Body)
		if err != nil {
			return 0, err
		}
	}
	return int64(PacketMetaLen + int(bodylen)), nil
}

func (p *HVPacket) WriteTo(writer io.Writer) (int64, error) {
	_, err := writer.Write(p.Meta)
	if err != nil {
		return 0, err
	}
	if len(p.Head) > 0 {
		_, err = writer.Write(p.Head)
		if err != nil {
			return 0, err
		}
	}
	if len(p.Body) > 0 {
		_, err = writer.Write(p.Body)
		if err != nil {
			return 0, err
		}
	}

	return int64(PacketMetaLen + len(p.Head) + len(p.Body)), nil
}

func (p *HVPacket) SetHead(b []uint8) {
	p.Meta.SetHeadLen(uint8(len(b)))
	p.Head = b
}

func (p *HVPacket) GetHead() []uint8 {
	return p.Head
}

func (p *HVPacket) SetBody(b []uint8) {
	p.Meta.SetBodyLen(uint16(len(b)))
	p.Body = b
}

func (p *HVPacket) GetBody() []uint8 {
	return p.Body
}

func (p *HVPacket) Reset() {
	p.Meta.Reset()
	p.Head = nil
	p.Body = nil
}
