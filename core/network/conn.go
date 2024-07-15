package network

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

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

type FuncOnConnPacket func(*Conn, *HVPacket)
type FuncOnConnEnable func(*Conn, bool)
type FuncOnAccpect func(net.Conn) bool
type FuncOnAuth func(data []byte) (auth.User, error)

var sid uint64 = 0

func GenConnID() string {
	return fmt.Sprintf("%d_%d", atomic.AddUint64(&sid, 1), time.Now().Unix())
}

type Conn struct {
	auth.User

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

func (s *Conn) ConnID() string {
	return s.id
}

func (s *Conn) Send(p *HVPacket) error {
	if !s.IsValid() {
		return ErrDisconn
	}
	select {
	case <-s.chClosed:
		return ErrDisconn
	case s.chWrite <- p:
		return nil
	}
}

func (s *Conn) Close() error {
	old := atomic.SwapInt32((*int32)(&s.status), int32(Disconnected))
	if old != Connected {
		return nil
	}

	select {
	case <-s.chClosed:
		return nil
	default:
		close(s.chClosed)
		return nil
	}
}

func (s *Conn) RemoteAddr() net.Addr {
	if !s.IsValid() {
		return nil
	}
	return s.conn.RemoteAddr()
}

func (s *Conn) LocalAddr() net.Addr {
	if !s.IsValid() {
		return nil
	}
	return s.conn.LocalAddr()
}

func (s *Conn) IsValid() bool {
	return s.Status() == Connected
}

func (s *Conn) Status() ConnStatus {
	return ConnStatus(atomic.LoadInt32((*int32)(&s.status)))
}

func (s *Conn) writeWork() error {
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

func (s *Conn) readWork() error {
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
