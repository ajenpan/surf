package gate

import (
	"time"

	"github.com/ajenpan/surf/core/network"
)

type ClientConnUserData struct {
	CreateAt          time.Time
	serverType2NodeID map[uint16]uint32

	nodeids map[uint32]struct{}
}

func NewClientConnUserData(s network.Conn) *ClientConnUserData {
	ret := &ClientConnUserData{
		CreateAt:          time.Now(),
		serverType2NodeID: make(map[uint16]uint32),
		nodeids:           make(map[uint32]struct{}),
	}
	return ret
}
