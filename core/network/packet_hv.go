package network

import (
	"io"
)

const (
	hvpFlagInit uint8 = 255
	// inner 224
	hvpSubFlagInnerStartAt_   uint8 = 224
	hvpSubFlagHandShake       uint8 = 225
	hvpSubFlagCmd             uint8 = 226
	hvpSubFlagCmdResult       uint8 = 227
	hvpSubFlagHandShakeFinish uint8 = 228
	hvpSubFlagHeartbeat       uint8 = 229
	hvpSubFlagInnerEndAt_     uint8 = 255
)

// | flag | subflag | len |
// | 1    | 1       | 2   |
const HVPackMetaLen = 1 + 1 + 2
const HVPacketMaxBodySize uint16 = (0xFFFF - 1)

type hvHead []byte

type HVPacket struct {
	head hvHead
	body []byte
}

func newHead() hvHead {
	return make([]byte, HVPackMetaLen)
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
	return int64(HVPackMetaLen + int(bodylen)), nil
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

	return int64(HVPackMetaLen + len(p.body)), nil
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
