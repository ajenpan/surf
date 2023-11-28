package server

import (
	"encoding/binary"
	"io"

	"github.com/ajenpan/surf/msg"
	"github.com/ajenpan/surf/server/tcp"
)

const (
	PacketTypeRouteMsgWraper uint8 = 1 // route
)

type MsgType = uint8

const (
	MsgTypeAsync    MsgType = 0
	MsgTypeRequest  MsgType = 1
	MsgTypeResponse MsgType = 2
)

func init() {
	tcp.RegPacket(PacketTypeRouteMsgWraper, func() tcp.Packet {
		return NewMsgWraper()
	})
}

type AsyncMsg = msg.AsyncMsgWrap
type RequestMsg = msg.RequestMsgWrap
type ResponseMsg = msg.ResponseMsgWrap
type ResponseError = msg.Error

// typ - bodylen - uid
// 1 - 3 - 4
const msgHeadSize = 8

type msgHead []byte

type MsgWraper struct {
	msgHead
	body []byte
}

func NewMsgWraper() *MsgWraper {
	return &MsgWraper{
		msgHead: make([]byte, msgHeadSize),
	}
}
func (m *MsgWraper) PacketType() uint8 {
	return PacketTypeRouteMsgWraper
}

func (m *MsgWraper) GetMsgtype() MsgType {
	return m.msgHead[0]
}

func (m *MsgWraper) SetMsgtype(typ MsgType) {
	m.msgHead[0] = typ
}

func (m *MsgWraper) SetBodyLen(len uint32) {
	tcp.PutUint24(m.msgHead[1:4], len)
}
func (m *MsgWraper) GetBodyLen() uint32 {
	return tcp.GetUint24(m.msgHead[1:4])
}

func (m *MsgWraper) GetUid() uint32 {
	return binary.LittleEndian.Uint32(m.msgHead[4:8])
}

func (m *MsgWraper) SetUid(uid uint32) {
	binary.LittleEndian.PutUint32(m.msgHead[4:8], uid)
}

func (m *MsgWraper) GetBody() []byte {
	return m.body
}
func (m *MsgWraper) SetBody(b []byte) {
	m.body = b
	m.SetBodyLen(uint32(len(b)))
}

func (m *MsgWraper) ReadFrom(r io.Reader) (int64, error) {
	// 1+3+4 + bodylen
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

func (m *MsgWraper) WriteTo(w io.Writer) (int64, error) {
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
