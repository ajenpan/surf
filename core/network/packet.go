package network

import (
	"io"
)

const (
	RoutePacketType byte = 1 // route
	HVPacketType    byte = 0x10
)

type Packet interface {
	PacketType() byte
	io.ReaderFrom
	io.WriterTo
}

func GetUint24(b []byte) uint32 {
	_ = b[2] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
}

func PutUint24(b []byte, v uint32) {
	_ = b[2] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
}

type hvPacketFlag = uint8

const (
	// inner 224
	HVPacketTypeInnerStartAt_   hvPacketFlag = 224
	HVPacketFlagHandShake       hvPacketFlag = 225
	HVPacketFlagCmd             hvPacketFlag = 226
	HVPacketFlagCmdResult       hvPacketFlag = 227
	HVPacketFlagHandShakeResult hvPacketFlag = 228
	HVPacketFlagHeartbeat       hvPacketFlag = 229
	HVPacketFlagPacket          hvPacketFlag = 230
	HVPcketTypeInnerEndAt_      hvPacketFlag = 255
)

// const hvPacketMaxBodySize = 0x7FFFFF

const hvPackMetaLen = 4

type hvHead []byte

type HVPacket struct {
	head hvHead
	body []byte
}

func newHead() hvHead {
	return make([]byte, hvPackMetaLen)
}

func NewHVPacket() *HVPacket {
	return &HVPacket{
		head: newHead(),
	}
}

func (hr hvHead) getFlag() byte {
	return hr[0]
}

func (hr hvHead) setFlag(f byte) {
	hr[0] = f
}

func (hr hvHead) getSubFlag() byte {
	return hr[1]
}

func (hr hvHead) setSubFlag(f byte) {
	hr[1] = f
}

func (hr hvHead) getBodyLen() uint16 {
	return uint16(hr[2]) | uint16(hr[3])<<8
}

func (hr hvHead) setBodyLen(l uint16) {
	hr[2] = byte(l)
	hr[3] = byte(l >> 8)
}

func (hr hvHead) reset() {
	for i := 0; i < len(hr); i++ {
		hr[i] = 0
	}
}

func (p *HVPacket) PacketType() byte {
	return HVPacketType
}

func (p *HVPacket) ReadFrom(reader io.Reader) (int64, error) {
	var err error
	if _, err = io.ReadFull(reader, p.head); err != nil {
		return 0, err
	}

	bodylen := p.head.getBodyLen()

	if bodylen > 0 {
		p.body = make([]byte, bodylen)
		_, err = io.ReadFull(reader, p.body)
		if err != nil {
			return 0, err
		}
	}
	return int64(hvPackMetaLen + int(bodylen)), nil
}

func (p *HVPacket) WriteTo(writer io.Writer) (int64, error) {
	_, err := writer.Write(p.head)
	if err != nil {
		return 0, err
	}

	if len(p.body) > 0 {
		_, err = writer.Write(p.body)
		if err != nil {
			return 0, err
		}
	}

	return int64(hvPackMetaLen + len(p.body)), nil
}

func (p *HVPacket) SetFlag(h byte) {
	p.head.setFlag(h)
}

func (p *HVPacket) GetFlag() byte {
	return p.head.getFlag()
}

func (p *HVPacket) SetSubFlag(h byte) {
	p.head.setSubFlag(h)
}

func (p *HVPacket) GetSubFlag() byte {
	return p.head.getSubFlag()
}

func (p *HVPacket) SetBody(b []byte) {
	p.body = b
	p.head.setBodyLen(uint16(len(b)))
}

func (p *HVPacket) GetBody() []byte {
	return p.body
}

func (p *HVPacket) Reset() {
	p.head.reset()
	p.body = nil
}
