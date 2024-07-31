package network

import "encoding/binary"

const RoutePackMetaLen = 4

type RoutePacket []byte

// | from | to | seqid | type | ttl | code | body |
// | 4    | 4  | 4     | 1    | 1   | 2    | n    |

func (r RoutePacket) GetFrom() uint32 {
	return binary.LittleEndian.Uint32(r[0:4])
}

func (r RoutePacket) GetTo() uint32 {
	return binary.LittleEndian.Uint32(r[4:8])
}

func (r RoutePacket) GetSeqid() uint32 {
	return binary.LittleEndian.Uint32(r[8:12])
}

func (r RoutePacket) GetType() uint8 {
	return r[12]
}

func (r RoutePacket) GetCode() uint16 {
	return binary.LittleEndian.Uint16(r[14:16])
}

func (r RoutePacket) GetBody() []byte {
	return r[16:]
}
