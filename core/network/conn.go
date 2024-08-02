package network

import (
	"errors"

	"github.com/ajenpan/surf/core/auth"
)

var ErrDisconn = errors.New("socket disconnected")
var ErrInvalidPacket = errors.New("invalid packet")

var DefaultTimeoutSec = 30
var DefaultMinTimeoutSec = 10

type ConnStatus = int32

const (
	Disconnected ConnStatus = iota
	Connectting  ConnStatus = iota
	Connected    ConnStatus = iota
)

type Conn interface {
	auth.User

	SetUserData(any)
	GetUserData() any

	ConnID() string
	Send(*HVPacket) error
	Close() error
	Enable() bool
	RemoteAddr() string
}
