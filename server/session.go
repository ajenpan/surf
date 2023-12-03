package server

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"
)

type FuncRespCallback = func(Session, *ResponseMsg)
type FuncOnSessionMessage func(Session, *MsgWraper)
type FuncOnSessionStatus func(Session, bool)
type FuncNewSessionID func() string

var sid int64 = 0

func NewSessionID() string {
	return fmt.Sprintf("%d_%d", atomic.AddInt64(&sid, 1), time.Now().Unix())
}

// todo list :
// 1. tcp socket session
// 2. web socket session

type Session interface {
	UserID() uint32

	SessionID() string
	SessionType() string

	IsValid() bool
	Close()

	Send(msg *MsgWraper) error
	SendAsync(uid uint32, m proto.Message) error
	SendRequest(uid uint32, m proto.Message, cb FuncRespCallback) error
	SendResponse(uid uint32, req *RequestMsg, resp proto.Message, err error) error

	RemoteAddr() net.Addr
}
