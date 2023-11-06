package server

import (
	"fmt"
	"sync/atomic"
	"time"
)

// todo list :
// 1. tcp socket session
// 2. web socket session

type User interface {
	UID() uint64
	UserName() string
	UserRole() string
}

type Session interface {
	User
	ID() string
	Valid() bool
	Close() error
	Send(msg *Message) error
	SessionType() string
}

type FuncOnSessionMessage func(Session, *Message)
type FuncOnSessionStatus func(Session, bool)

type FuncNewSessionID func() string

var sid int64 = 0

func NewSessionID() string {
	return fmt.Sprintf("%d_%d", atomic.AddInt64(&sid, 1), time.Now().Unix())
}
