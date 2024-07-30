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
