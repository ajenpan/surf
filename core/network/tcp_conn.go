package network

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/ajenpan/surf/core/auth"
)

type FuncOnConnPacket func(Conn, *HVPacket)
type FuncOnConnEnable func(Conn, bool)
type FuncOnConnAuth func(data []byte) (auth.User, error)

var sid uint64 = 0

func GenConnID() string {
	return fmt.Sprintf("%d_%d", atomic.AddUint64(&sid, 1), time.Now().Unix())
}

func newTcpConn(id string, uinfo auth.User, imp net.Conn, rwtimeout time.Duration) *TcpConn {
	return &TcpConn{
		User:       uinfo,
		id:         id,
		imp:        imp,
		timeOut:    rwtimeout,
		status:     Initing,
		chClosed:   make(chan struct{}),
		chWrite:    make(chan *HVPacket, 10),
		chRead:     make(chan *HVPacket, 10),
		lastSendAt: time.Now().Unix(),
		lastRecvAt: time.Now().Unix(),
	}
}

type TcpConn struct {
	auth.User

	imp net.Conn
	id  string

	chWrite  chan *HVPacket
	chRead   chan *HVPacket
	chClosed chan struct{}

	timeOut time.Duration

	lastSendAt int64
	lastRecvAt int64

	status ConnStatus

	writeSize int64
	readSize  int64
	userdata  any
}

func (s *TcpConn) SetUserData(d any) {
	s.userdata = d
}

func (s *TcpConn) GetUserData() any {
	return s.userdata
}

func (s *TcpConn) ConnID() string {
	return s.id
}

func (s *TcpConn) Send(p *HVPacket) error {
	if !s.Enable() {
		return ErrDisconn
	}
	select {
	case <-s.chClosed:
		return ErrDisconn
	case s.chWrite <- p:
		return nil
	}
}

func (c *TcpConn) Close() error {
	select {
	case <-c.chClosed:
		return nil
	default:
		atomic.StoreInt32((*int32)(&c.status), int32(Closed))

		close(c.chClosed)
		close(c.chWrite)
		close(c.chRead)
		return nil
	}
}

func (s *TcpConn) RemoteAddr() string {
	if !s.Enable() {
		return ""
	}
	return s.imp.RemoteAddr().String()
}

func (s *TcpConn) LocalAddr() net.Addr {
	if !s.Enable() {
		return nil
	}
	return s.imp.LocalAddr()
}

func (s *TcpConn) Enable() bool {
	return s.Status() == Connected
}

func (s *TcpConn) Status() ConnStatus {
	return ConnStatus(atomic.LoadInt32((*int32)(&s.status)))
}

func (c *TcpConn) writeWork() error {
	for {
		select {
		case <-c.chClosed:
			return nil
		case p, ok := <-c.chWrite:
			if !ok {
				return nil
			}
			c.imp.SetWriteDeadline(time.Now().Add(c.timeOut))
			n, err := p.WriteTo(c.imp)
			if err != nil {
				return err
			}
			c.writeSize += n
			atomic.SwapInt64(&c.lastSendAt, time.Now().UnixMilli())
		}
	}
}

func (s *TcpConn) readWork() error {
	for {
		s.imp.SetReadDeadline(time.Now().Add(s.timeOut))

		pk := NewHVPacket()
		n, err := pk.ReadFrom(s.imp)
		if err != nil {
			return err
		}

		s.readSize += n
		atomic.SwapInt64(&s.lastRecvAt, time.Now().UnixMilli())
		select {
		case <-s.chClosed:
			return nil
		case s.chRead <- pk:
		}
	}
}
