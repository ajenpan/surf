package tcp

import (
	"errors"
	"io"
)

// Codec constants.
const (
	//8MB 8388607
	MaxPacketBodySize = 0x7FFFFF

	//255
	MaxPacketHeadSize = 0xFF
)

var ErrWrongPacketType = errors.New("wrong packet type")
var ErrBodySizeWrong = errors.New("packet body size error")
var ErrHeadSizeWrong = errors.New("packet head size error")
var ErrParseHead = errors.New("parse packet error")
var ErrDisconn = errors.New("socket disconnected")

// |-Meta---------------------------------|-Body------------|
// |-<PacketType>-|-<HeadLen>-|-<BodyLen>-|-<Head>-|-<Body>-|
// |-1------------|-1---------|-3---------|--------|--------|

type PacketType = uint8

const (
	// inner
	PacketTypeInnerStartAt_ PacketType = iota
	PacketTypeAck
	PacketTypeHeartbeat
	PacketTypeEcho
	PacketTypeAuth
	PacketTypeInnerEndAt_
)

type Packet interface {
	io.ReaderFrom
	io.WriterTo
}

func GetUint24(b []uint8) uint32 {
	_ = b[2] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
}

func PutUint24(b []uint8, v uint32) {
	_ = b[2] // early bounds check to guarantee safety of writes below
	b[0] = uint8(v)
	b[1] = uint8(v >> 8)
	b[2] = uint8(v >> 16)
}

const PackMetaLen = 5

type THVPacketMeta []uint8

func NewPackMeta() THVPacketMeta {
	return make([]uint8, PackMetaLen)
}

func (hr THVPacketMeta) GetType() uint8 {
	return hr[0]
}

func (h THVPacketMeta) GetHeadLen() uint8 {
	return h[1]
}

func (hr THVPacketMeta) GetBodyLen() uint32 {
	return GetUint24(hr[2:5])
}

func (hr THVPacketMeta) SetType(t uint8) {
	hr[0] = t
}

func (h THVPacketMeta) SetHeadLen(t uint8) {
	h[1] = t
}

func (hr THVPacketMeta) SetBodyLen(l uint32) {
	PutUint24(hr[2:5], l)
}

func (hr THVPacketMeta) Reset() {
	for i := 0; i < len(hr); i++ {
		hr[i] = 0
	}
}

func NewEmptyTHVPacket() *THVPacket {
	return &THVPacket{
		Meta: NewPackMeta(),
	}
}

func NewPackFrame(t uint8, h []uint8, b []uint8) *THVPacket {
	p := NewEmptyTHVPacket()
	p.SetType(t)
	p.SetHead(h)
	p.SetBody(b)
	return p
}

type THVPacket struct {
	Meta THVPacketMeta
	Head []uint8
	Body []uint8
}

func (p *THVPacket) ReadFrom(reader io.Reader) (int64, error) {
	var err error
	metalen, err := io.ReadFull(reader, p.Meta)
	if err != nil {
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
	return int64(metalen + int(headlen) + int(bodylen)), nil
}

func (p *THVPacket) WriteTo(writer io.Writer) (int64, error) {
	ret := int64(0)
	n, err := writer.Write(p.Meta)
	ret += int64(n)
	if err != nil {
		return -1, err
	}
	if len(p.Head) > 0 {
		n, err = writer.Write(p.Body)
		ret += int64(n)
		if err != nil {
			return -1, err
		}
	}
	if len(p.Body) > 0 {
		n, err = writer.Write(p.Body)
		ret += int64(n)
		if err != nil {
			return -1, err
		}
	}
	return ret, nil
}

func (p *THVPacket) Name() string {
	return "tcp-binary"
}

func (p *THVPacket) Reset() {
	p.Meta.Reset()
	if p.Head != nil {
		p.Head = p.Head[:0]
	}
	if p.Body != nil {
		p.Body = p.Body[:0]
	}
}

func (p *THVPacket) Clone() *THVPacket {
	ret := &THVPacket{
		Meta: p.Meta[:],
	}
	if p.Head != nil {
		ret.Head = p.Head[:]
	}
	if p.Body != nil {
		ret.Body = p.Body[:]
	}
	return ret
}

func (p *THVPacket) SetType(t uint8) {
	p.Meta.SetType(t)
}

func (p *THVPacket) GetType() uint8 {
	return p.Meta.GetType()
}

func (p *THVPacket) SetHead(b []uint8) {
	if len(b) > MaxPacketHeadSize {
		panic(ErrHeadSizeWrong)
	}
	p.Head = b
	p.Meta.SetHeadLen(uint8(len(b)))
}

func (p *THVPacket) GetHead() []uint8 {
	return p.Head
}

func (p *THVPacket) SetBody(b []uint8) {
	if len(b) > MaxPacketBodySize {
		panic(ErrBodySizeWrong)
	}
	p.Body = b
	p.Meta.SetBodyLen(uint32(len(b)))
}

func (p *THVPacket) GetBody() []uint8 {
	return p.Body
}

func (p *THVPacket) GetMeta() THVPacketMeta {
	return p.Meta
}
