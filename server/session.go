package server

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

// todo list :
// 1. tcp socket session
// 2. web socket session

type User interface {
	UserID() uint32
	UserRole() string
	UserName() string
}

type Session interface {
	User

	SessionID() string
	SessionType() string

	IsValid() bool
	Close() error
	Send(msg *Message) error

	RemoteAddr() net.Addr
}

type FuncOnSessionMessage func(Session, *Message)
type FuncOnSessionStatus func(Session, bool)

type FuncNewSessionID func() string

var sid int64 = 0

func NewSessionID() string {
	return fmt.Sprintf("%d_%d", atomic.AddInt64(&sid, 1), time.Now().Unix())
}
