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

type UserMeta interface {
	LoadAndDelete(key any) (value any, loaded bool)
	Swap(key, value any) (previous any, loaded bool)
	Load(key any) (value any, ok bool)
	Store(key, value any)
}

type Conn interface {
	auth.User
	UserMeta

	ConnID() string
	Send(*HVPacket) error
	Close() error
	Enable() bool
	RemoteAddr() string
}
