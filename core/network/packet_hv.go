package network

import (
	"io"
)

// | pktype | subtype | headlen | bodylen | body |
// | 1      | 1       | 1       | 2       | n    |
const PacketHeadLen = 1 + 1 + 1 + 2
const PacketMaxBodySize uint16 = (0xFFFF - 1)

type Packet_MsgType uint8

const (
	PacketType_Route     Packet_MsgType = 0
	PacketType_Inner     Packet_MsgType = 1
	PacketType_NodeInner Packet_MsgType = 2
)

type Packet_Inner = uint8

const (
	PacketType_Inner_HandShake       Packet_Inner = 5
	PacketType_Inner_Cmd             Packet_Inner = 6
	PacketType_Inner_CmdResult       Packet_Inner = 7
	PacketType_Inner_HandShakeFinish Packet_Inner = 8
	PacketType_Inner_Heartbeat       Packet_Inner = 9
)

type PacketMeta []uint8

type HVPacket struct {
	Meta PacketMeta
	Head []uint8
	Body []uint8
}

func NewHead() PacketMeta {
	return make([]uint8, PacketHeadLen)
}

func NewHVPacket() *HVPacket {
	return &HVPacket{
		Meta: NewHead(),
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
	(*hr)[2] = uint8(l)
	(*hr)[3] = uint8(l >> 8)
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
		_, err = io.ReadFull(reader, p.Body)
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
	return int64(PacketHeadLen + int(bodylen)), nil
}

func (p *HVPacket) WriteTo(writer io.Writer) (int64, error) {
	_, err := writer.Write(p.Meta)
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
	p.Meta.SetBodyLen(uint16(len(b)))
}

func (p *HVPacket) GetBody() []uint8 {
	return p.Body
}

func (p *HVPacket) Reset() {
	p.Meta.Reset()
	p.Meta = nil
	p.Body = nil
}
