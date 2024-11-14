package network

import "encoding/binary"

// | msgid | syn | errcode | msgtype | marshal |
// | 4     | 4   | 2       | 1       | 1       |
const NodePackHeadLen = 4 + 4 + 2 + 1 + 1 // 12

func NewNodePacketHead() nodePacketHead {
	return make(nodePacketHead, NodePackHeadLen)
}

type nodePacketHead []uint8

func NewEmptyNodePacket() *NodePacket {
	return &NodePacket{
		Head: NewNodePacketHead(),
		Body: nil,
	}
}

type NodePacket struct {
	Head nodePacketHead
	Body []byte
}

func (r *NodePacket) GetMsgId() uint32 {
	return binary.LittleEndian.Uint32(r.Head[0:4])
}

func (r *NodePacket) SetMsgId(d uint32) {
	binary.LittleEndian.PutUint32(r.Head[0:4], d)
}

func (r *NodePacket) GetSyn() uint32 {
	return binary.LittleEndian.Uint32(r.Head[4:8])
}

func (r *NodePacket) SetSyn(d uint32) {
	binary.LittleEndian.PutUint32(r.Head[4:8], d)
}

func (r *NodePacket) SetErrCode(d int16) {
	binary.LittleEndian.PutUint16(r.Head[8:10], uint16(d))
}

func (r *NodePacket) GetErrCode() int16 {
	return int16(binary.LittleEndian.Uint16(r.Head[8:10]))
}

func (r *NodePacket) GetMsgType() uint8 {
	return r.Head[10]
}

func (r *NodePacket) SetMsgType(d uint8) {
	r.Head[10] = d
}

func (r *NodePacket) GetMarshal() uint8 {
	return r.Head[11]
}

func (r *NodePacket) SetMarshal(d uint8) {
	r.Head[11] = d
}

func (r *NodePacket) SetBody(b []byte) {
	r.Body = b
}

func (r *NodePacket) GetBody() []byte {
	return r.Body
}

func (r *NodePacket) ToHVPacket() *HVPacket {
	ret := NewHVPacket()
	ret.Meta.SetType(PacketType_Node)
	ret.SetBody(r.Body)
	return ret
}

func (r *NodePacket) FromHVPacket(pk *HVPacket) {
	r.Head = pk.GetHead()
	r.Body = pk.Body
}
