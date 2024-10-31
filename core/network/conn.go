package network

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ajenpan/surf/core/auth"
)

var ErrDisconn = errors.New("conn disconnected")
var ErrInvalidPacket = errors.New("invalid packet")

var DefaultTimeoutSec = 30
var DefaultHeartbeatSec = 10

type ConnStatus = int32

const (
	Initing      ConnStatus = iota
	Connectting  ConnStatus = iota
	Connected    ConnStatus = iota
	Disconnected ConnStatus = iota
	Closed       ConnStatus = iota
)

type FuncOnConnAuth func(data []byte) (auth.User, error)
type FuncOnConnStatus func(Conn, bool)
type FuncOnConnPacket func(Conn, *HVPacket)

var sid uint64 = 0

func GenConnID() string {
	return fmt.Sprintf("%d_%d", atomic.AddUint64(&sid, 1), time.Now().Unix())
}

type Conn interface {
	auth.User

	ConnID() string

	SetUserData(any)
	GetUserData() any

	Send(*HVPacket) error

	Close() error
	Enable() bool
	RemoteAddr() string
}
