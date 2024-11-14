package gate

import (
	"time"

	"github.com/ajenpan/surf/core/network"
)

type ConnUserData struct {
	CreateAt          time.Time
	serverType2NodeID map[uint16]uint32

	nodeids map[uint32]struct{}
}

func NewConnUserData(s network.Conn) *ConnUserData {
	ret := &ConnUserData{
		CreateAt:          time.Now(),
		serverType2NodeID: make(map[uint16]uint32),
		nodeids:           make(map[uint32]struct{}),
	}
	return ret
}
