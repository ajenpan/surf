package tcp

import (
	"errors"
	"io"
	"sync"
)

type Packet interface {
	PacketType() uint8
	io.ReaderFrom
	io.WriterTo
}

var maplock sync.RWMutex
var packetMap = make(map[uint8]func() Packet)

func NewPacket(typ uint8) Packet {
	maplock.RLock()
	defer maplock.RUnlock()
	if f, ok := packetMap[typ]; ok {
		return f()
	}
	return newHVPacket()
}
func RegPacket(typ uint8, f func() Packet) error {
	maplock.Lock()
	defer maplock.Unlock()
	if _, ok := packetMap[typ]; ok {
		return errors.New("packet type already registered")
	}
	packetMap[typ] = f
	return nil
}

const HVPacketType = 0xE0

func init() {
	RegPacket(HVPacketType, func() Packet { return newHVPacket() })
}

type hvPacketSubType = uint8

var ErrWrongPacketType = errors.New("wrong packet type")
var ErrWronMetaData = errors.New("wrong packet meta data")
var ErrBodySizeWrong = errors.New("packet body size error")

const (
	// inner 224
	PacketTypeInnerStartAt_ hvPacketSubType = 0xE0
	PacketTypeHandShake     hvPacketSubType = PacketTypeInnerStartAt_ + iota
	PacketTypeActionRequire hvPacketSubType = PacketTypeInnerStartAt_ + iota
	PacketTypeDoAction      hvPacketSubType = PacketTypeInnerStartAt_ + iota
	PacketTypeAckFailure    hvPacketSubType = PacketTypeInnerStartAt_ + iota
	PacketTypeAckSuccess    hvPacketSubType = PacketTypeInnerStartAt_ + iota
	PacketTypeHeartbeat     hvPacketSubType = PacketTypeInnerStartAt_ + iota
	PacketTypeEcho          hvPacketSubType = PacketTypeInnerStartAt_ + iota
	PacketTypeMessage       hvPacketSubType = PacketTypeInnerStartAt_ + iota
	PacketTypeInnerEndAt_   hvPacketSubType = PacketTypeInnerStartAt_ + iota
)

var MaxPacketBodySize = 0x7FFFFF

// LittleEndian
func GetUint24(b []uint8) uint32 {
	_ = b[2] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
}

// LittleEndian
func PutUint24(b []uint8, v uint32) {
	_ = b[2] // early bounds check to guarantee safety of writes below
	b[0] = uint8(v)
	b[1] = uint8(v >> 8)
	b[2] = uint8(v >> 16)
}

// LittleEndian
func GetUint16(b []uint8) uint16 {
	_ = b[1]
	return uint16(b[0]) | uint16(b[1])<<8
}

// LittleEndian
func PutUint16(b []uint8, v uint16) {
	_ = b[1]
	b[0] = uint8(v)
	b[1] = uint8(v >> 8)
}

const PackMetaLen = 4

type head []uint8

func newHead() head {
	return make([]uint8, PackMetaLen)
}

func (hr head) getType() uint8 {
	return hr[0]
}

func (hr head) setType(l uint8) {
	hr[0] = l
}

func (h head) getBodyLen() uint32 {
	return GetUint24(h[1:4])
}

func (hr head) setBodyLen(l uint32) {
	PutUint24(hr[1:4], l)
}

func (hr head) reset() {
	for i := 0; i < len(hr); i++ {
		hr[i] = 0
	}
}

func (hr head) check() error {
	if hr.getType() <= PacketTypeInnerStartAt_ || hr.getType() >= PacketTypeInnerEndAt_ {
		return ErrInvalidPacket
	}
	if hr.getBodyLen() > uint32(MaxPacketBodySize) {
		return ErrBodySizeWrong
	}
	return nil
}

type hvPacket struct {
	head head
	body []uint8
}

func newHVPacket() *hvPacket {
	return &hvPacket{
		head: newHead(),
	}
}

func (p *hvPacket) PacketType() uint8 {
	return HVPacketType
}

func (p *hvPacket) ReadFrom(reader io.Reader) (int64, error) {
	var err error
	if _, err = io.ReadFull(reader, p.head); err != nil {
		return 0, err
	}

	bodylen := p.head.getBodyLen()

	if bodylen > 0 {
		p.body = make([]uint8, bodylen)
		_, err = io.ReadFull(reader, p.body)
		if err != nil {
			return 0, err
		}
	}
	return int64(PackMetaLen + int(bodylen)), nil
}

func (p *hvPacket) WriteTo(writer io.Writer) (int64, error) {
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
	return int64(PackMetaLen + len(p.body)), nil
}

func (p *hvPacket) SetType(h uint8) {
	p.head.setType(h)
}

func (p *hvPacket) GetType() uint8 {
	return p.head.getType()
}

func (p *hvPacket) SetBody(b []uint8) {
	p.body = b
	p.head.setBodyLen(uint32(len(b)))
}

func (p *hvPacket) GetBody() []uint8 {
	return p.body
}

func (p *hvPacket) Reset() {
	p.head.reset()
	p.body = nil
}
