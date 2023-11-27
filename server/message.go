package server

import (
	"encoding/binary"
	"io"

	"github.com/ajenpan/surf/server/tcp"
)

const (
	PacketBinaryRouteType uint8 = 1 // route
)

type MsgType = uint8

const (
	MsgTypeAsync    MsgType = 0
	MsgTypeRequest  MsgType = 1
	MsgTypeResponse MsgType = 2
)

func init() {
	tcp.RegPacket(PacketBinaryRouteType, func() tcp.Packet {
		return NewMessage()
	})
}

const msgHeadSize = 8

type msgHead []byte

type Message struct {
	msgHead
	body []byte
}

func NewMessage() *Message {
	return &Message{
		msgHead: make([]byte, msgHeadSize),
	}
}
func (m *Message) PacketType() uint8 {
	return PacketBinaryRouteType
}

func (m *Message) GetMsgtype() MsgType {
	return m.msgHead[0]
}

func (m *Message) SetMsgtype(typ MsgType) {
	m.msgHead[0] = typ
}

func (m *Message) SetBodyLen(len uint32) {
	tcp.PutUint24(m.msgHead[1:4], len)
}
func (m *Message) GetBodyLen() uint32 {
	return tcp.GetUint24(m.msgHead[1:4])
}

func (m *Message) GetUid() uint32 {
	return binary.LittleEndian.Uint32(m.msgHead[4:8])
}

func (m *Message) SetUid(uid uint32) {
	binary.LittleEndian.PutUint32(m.msgHead[4:8], uid)
}

func (m *Message) GetBody() []byte {
	return m.body
}
func (m *Message) SetBody(b []byte) {
	m.body = b
	m.SetBodyLen(uint32(len(b)))
}

func (m *Message) ReadFrom(r io.Reader) (int64, error) {
	// 3+1+4 + bodylen
	var err error
	_, err = io.ReadFull(r, m.msgHead)
	if err != nil {
		return 0, err
	}
	bodylen := m.GetBodyLen()
	if bodylen > 0 {
		m.body = make([]byte, bodylen)
		_, err = io.ReadFull(r, m.body)
	}
	return (8 + int64(bodylen)), err
}

func (m *Message) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(m.msgHead)
	if err != nil {
		return 0, err
	}
	if len(m.body) > 0 {
		bn, err := w.Write(m.body)
		if err != nil {
			return 0, err
		}
		n += bn
	}
	return int64(n), nil
}
