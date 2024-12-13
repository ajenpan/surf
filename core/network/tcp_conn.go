package network

import (
	"net"
	"sync/atomic"
	"time"
)

func newTcpConn(id string, uinfo User, imp net.Conn, rwtimeout time.Duration) *TcpConn {
	if rwtimeout.Seconds() < float64(DefaultHeartbeatSec) {
		rwtimeout = time.Duration(DefaultHeartbeatSec*2) * time.Second
	}
	ret := &TcpConn{
		id:         id,
		imp:        imp,
		timeOut:    rwtimeout,
		status:     Initing,
		chClosed:   make(chan struct{}),
		chWrite:    make(chan *HVPacket, 10),
		chRead:     make(chan *HVPacket, 10),
		lastSendAt: time.Now().UnixMilli(),
		lastRecvAt: time.Now().UnixMilli(),
	}
	if uinfo != nil {
		ret.userInfo.fromUser(uinfo)
	}
	return ret
}

type TcpConn struct {
	userInfo
	SYNGenerator

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

func (s *TcpConn) ConnId() string {
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

func (s *TcpConn) LocalAddr() string {
	if !s.Enable() {
		return ""
	}
	return s.imp.LocalAddr().String()
}

func (s *TcpConn) Enable() bool {
	return s.Status() == Connected
}

func (s *TcpConn) Status() ConnStatus {
	return ConnStatus(atomic.LoadInt32((*int32)(&s.status)))
}

func (s *TcpConn) ReadPacket() (*HVPacket, error) {
	pk := NewHVPacket()
	_, err := pk.ReadFrom(s.imp)
	return pk, err
}

func (s *TcpConn) WritePacket(hv *HVPacket) error {
	_, err := hv.WriteTo(s.imp)
	return err
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
