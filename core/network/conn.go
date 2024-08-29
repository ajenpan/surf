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
	Initing      ConnStatus = iota
	Connectting  ConnStatus = iota
	Connected    ConnStatus = iota
	Disconnected ConnStatus = iota
	Closed       ConnStatus = iota
)

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
