package network

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
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

type User interface {
	UserID() uint32
	UserName() string
	UserRole() uint16
}

type userInfo struct {
	UId   uint32 `json:"uid"`
	UName string `json:"uname"`
	URole uint16 `json:"urid"`
}

func (u *userInfo) UserID() uint32 {
	return u.UId
}

func (u *userInfo) UserRole() uint16 {
	return u.URole
}

func (u *userInfo) UserName() string {
	return u.UName
}

func (u *userInfo) fromUser(user User) {
	u.UId = user.UserID()
	u.UName = user.UserName()
	u.URole = user.UserRole()
}

type SYNGenerator struct {
	synIdx uint32
}

func (s *SYNGenerator) NextSYN() uint32 {
	ret := atomic.AddUint32(&s.synIdx, 1)
	if ret == 0 {
		return atomic.AddUint32(&s.synIdx, 1)
	}
	return ret
}

type FuncOnConnAuth func(data []byte) (User, error)
type FuncOnConnEnable func(Conn, bool)
type FuncOnConnPacket func(Conn, *HVPacket)

var sid uint64 = 0

func GenConnID() string {
	return fmt.Sprintf("%d_%d", atomic.AddUint64(&sid, 1), time.Now().Unix())
}

type Conn interface {
	User

	ConnID() string

	SetUserData(any)
	GetUserData() any

	NextSYN() uint32

	Send(*HVPacket) error

	Close() error
	Enable() bool
	RemoteAddr() string
}
