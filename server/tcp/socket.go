package tcp

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ajenpan/surf/auth"
)

var ErrDisconn = errors.New("socket disconnected")
var ErrInvalidPacket = errors.New("invalid packet")

var DefaultTimeoutSec = 30
var DefaultMinTimeoutSec = 10

type SocketStatus int32

const (
	Disconnected SocketStatus = iota
	// Connectting  SocketStatus = iota
	// Handshake    SocketStatus = iota
	Connected SocketStatus = iota
)

type SocketOptions struct {
	ID      string
	Timeout time.Duration
}

type SocketOption func(*SocketOptions)

func NewSocket(conn net.Conn, opts SocketOptions) *Socket {
	if opts.Timeout < time.Duration(DefaultMinTimeoutSec)*time.Second {
		opts.Timeout = time.Duration(DefaultTimeoutSec) * time.Second
	}

	ret := &Socket{
		id:       opts.ID,
		conn:     conn,
		timeOut:  opts.Timeout,
		chWrite:  make(chan Packet, 100),
		chClosed: make(chan struct{}),
		status:   Disconnected,
	}
	return ret
}

type Socket struct {
	auth.UserInfo
	Meta sync.Map

	conn net.Conn
	id   string

	chWrite  chan Packet
	chClosed chan struct{}

	timeOut time.Duration

	lastSendAt int64
	lastRecvAt int64

	status SocketStatus
}

func (s *Socket) SessionID() string {
	return s.id
}

func (s *Socket) Send(p Packet) error {
	if !s.IsValid() {
		return ErrDisconn
	}
	select {
	case <-s.chClosed:
		return ErrDisconn
	case s.chWrite <- p:
	}
	return nil
}

func (s *Socket) Close() {
	stat := atomic.SwapInt32((*int32)(&s.status), int32(Disconnected))
	if stat == int32(Disconnected) {
		return
	}

	select {
	case <-s.chClosed:
		return
	default:
		close(s.chClosed)
	}

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	close(s.chWrite)
}

// returns the remote network address.
func (s *Socket) RemoteAddr() net.Addr {
	if s == nil || s.conn == nil {
		return nil
	}
	return s.conn.RemoteAddr()
}

func (s *Socket) LocalAddr() net.Addr {
	if s == nil || s.conn == nil {
		return nil
	}
	return s.conn.LocalAddr()
}

func (s *Socket) IsValid() bool {
	return s.Status() == Connected
}

func (s *Socket) Status() SocketStatus {
	return SocketStatus(atomic.LoadInt32((*int32)(&s.status)))
}

func (s *Socket) writeWork() error {
	defer func() {
		s.Close()
	}()
	for {
		select {
		case <-s.chClosed:
			return nil
		case p, ok := <-s.chWrite:
			if !ok {
				return nil
			}
			if err := writePacket(s.conn, s.timeOut, p); err != nil {
				return err
			}
			s.lastSendAt = time.Now().Unix()
		}
	}
}

func (s *Socket) readWork(recv chan<- Packet) error {
	defer func() {
		s.Close()
	}()
	for {
		select {
		case <-s.chClosed:
			return nil
		default:
			p, err := readPacket(s.conn, s.timeOut)
			if err != nil {
				return err
			}
			s.lastRecvAt = time.Now().Unix()
			recv <- p
		}
	}
}

func readPacket(conn net.Conn, timeout time.Duration) (Packet, error) {
	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout))
	}
	var err error
	pktype := make([]byte, 1)
	_, err = conn.Read(pktype)
	if err != nil {
		return nil, err
	}

	pk := NewPacket(pktype[0])
	if pk == nil {
		return nil, ErrInvalidPacket
	}
	_, err = pk.ReadFrom(conn)
	return pk, err
}

func readPacketT[PacketTypeT Packet](conn net.Conn, timeout time.Duration) (PacketTypeT, error) {
	pk, err := readPacket(conn, timeout)
	var pkk PacketTypeT
	if err != nil {
		return pkk, err
	}
	pkk, ok := pk.(PacketTypeT)
	if !ok {
		return pkk, ErrInvalidPacket
	}
	return pkk, nil
}

func writePacket(conn net.Conn, timeout time.Duration, p Packet) error {
	if timeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(timeout))
	}
	var err error
	_, err = conn.Write([]byte{p.PacketType()})
	if err != nil {
		return err
	}
	_, err = p.WriteTo(conn)
	return err
}

// func readHVPacket(conn net.Conn, p *hvPacket, timeout time.Duration) error {
// 	if timeout > 0 {
// 		conn.SetReadDeadline(time.Now().Add(timeout))
// 	}
// 	if _, err := io.ReadFull(conn, p.head); err != nil {
// 		return err
// 	}
// 	if err := p.head.check(); err != nil {
// 		return err
// 	}
// 	bodylen := p.head.getBodyLen()
// 	if bodylen > 0 {
// 		if int(bodylen) > cap(p.body) {
// 			p.body = make([]byte, bodylen)
// 		} else {
// 			p.body = p.body[:bodylen]
// 		}
// 		if _, err := io.ReadFull(conn, p.body); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
