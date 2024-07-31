package network

import (
	"fmt"
	"net"
	"sync"
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

type TcpConn struct {
	auth.User
	sync.Map

	conn net.Conn
	id   string

	chWrite  chan *HVPacket
	chRead   chan *HVPacket
	chClosed chan struct{}

	timeOut time.Duration

	lastSendAt int64
	lastRecvAt int64

	status ConnStatus

	writeSize int64
	readSize  int64
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
	old := atomic.SwapInt32((*int32)(&c.status), int32(Disconnected))
	if old != Connected {
		return nil
	}

	select {
	case <-c.chClosed:
		return nil
	default:
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
	return s.conn.RemoteAddr().String()
}

func (s *TcpConn) LocalAddr() net.Addr {
	if !s.Enable() {
		return nil
	}
	return s.conn.LocalAddr()
}

func (s *TcpConn) Enable() bool {
	return s.Status() == Connected
}

func (s *TcpConn) Status() ConnStatus {
	return ConnStatus(atomic.LoadInt32((*int32)(&s.status)))
}

func (s *TcpConn) writeWork() error {
	for {
		select {
		case <-s.chClosed:
			return nil
		case p, ok := <-s.chWrite:
			if !ok {
				return nil
			}
			s.conn.SetWriteDeadline(time.Now().Add(s.timeOut))
			n, err := p.WriteTo(s.conn)
			if err != nil {
				return err
			}
			s.writeSize += n
			s.lastSendAt = time.Now().Unix()
		}
	}
}

func (s *TcpConn) readWork() error {
	for {
		s.conn.SetReadDeadline(time.Now().Add(s.timeOut))

		pk := NewHVPacket()
		n, err := pk.ReadFrom(s.conn)
		if err != nil {
			return err
		}

		s.readSize += n
		s.lastRecvAt = time.Now().Unix()
		select {
		case <-s.chClosed:
			return nil
		case s.chRead <- pk:
		}
	}
}
