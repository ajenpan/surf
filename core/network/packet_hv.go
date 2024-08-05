package network

import (
	"io"
)

// | pktype | subflag | bodylen | body |
// | 1      | 1       | 2       | n    |
const PacketHeadLen = 1 + 1 + 2
const PacketMaxBodySize uint16 = (0xFFFF - 1)

type Packet_MsgType uint8

const (
	PacketType_Inner     Packet_MsgType = 0
	PacketType_Route     Packet_MsgType = 1
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
	return Packet_MsgType((*hr)[0])
}

func (hr *PacketHead) SetType(f Packet_MsgType) {
	(*hr)[0] = uint8(f)
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

func (hr *PacketHead) Reset() {
	(*hr) = NewHead()
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
