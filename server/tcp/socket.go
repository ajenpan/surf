package tcp

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type SocketStat int32

const (
	Disconnected SocketStat = iota
	Connected    SocketStat = iota
)

var DefaultTimeoutSec = 30
var DefaultMinTimeoutSec = 10

type OnMessageFunc func(*Socket, *THVPacket)
type OnConnStatFunc func(*Socket, bool)
type NewIDFunc func() string

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
		chSend:   make(chan Packet, 100),
		chClosed: make(chan bool),
		state:    Connected,
	}
	return ret
}

type UserInfo struct {
	UId   uint64
	UName string
	Role  string
}

func (u *UserInfo) UID() uint64 {
	return u.UId
}

func (u *UserInfo) UserName() string {
	return u.UName
}

func (u *UserInfo) UserRole() string {
	return u.Role
}

type Socket struct {
	*UserInfo

	Meta sync.Map

	conn  net.Conn   // low-level conn fd
	state SocketStat // current state
	id    string

	chSend   chan Packet // push message queue
	chClosed chan bool

	timeOut time.Duration

	lastSendAt int64
	lastRecvAt int64
}

func (s *Socket) ID() string {
	return s.id
}

func (s *Socket) SendPacket(p Packet) error {
	if atomic.LoadInt32((*int32)(&s.state)) == int32(Disconnected) {
		return ErrDisconn
	}
	s.chSend <- p
	return nil
}

func (s *Socket) Close() error {
	if s == nil {
		return nil
	}
	stat := atomic.SwapInt32((*int32)(&s.state), int32(Disconnected))
	if stat == int32(Disconnected) {
		return nil
	}

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	close(s.chSend)
	close(s.chClosed)
	return nil
}

// returns the remote network address.
func (s *Socket) RemoteAddr() string {
	if s == nil {
		return ""
	}
	if s.conn == nil {
		return ""
	}
	return s.conn.RemoteAddr().String()
}

func (s *Socket) LocalAddr() string {
	if s == nil {
		return ""
	}
	if s.conn == nil {
		return ""
	}
	return s.conn.LocalAddr().String()
}

func (s *Socket) Valid() bool {
	return s.Status() == Connected
}

// retrun socket work status
func (s *Socket) Status() SocketStat {
	if s == nil {
		return Disconnected
	}
	return SocketStat(atomic.LoadInt32((*int32)(&s.state)))
}

func (s *Socket) writeWork() {
	for p := range s.chSend {
		s.writePacket(p)
	}
}

func (s *Socket) readPacket(p Packet) error {
	if s.Status() == Disconnected {
		return ErrDisconn
	}
	err := readPacket(s.conn, p, s.timeOut)
	s.lastRecvAt = time.Now().Unix()
	return err
}

func (s *Socket) writePacket(p Packet) error {
	if s.Status() == Disconnected {
		return ErrDisconn
	}
	err := writePacket(s.conn, p, s.timeOut)
	s.lastSendAt = time.Now().Unix()
	return err
}

func readPacket(conn net.Conn, p Packet, timeout time.Duration) error {
	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout))
	}
	_, err := p.ReadFrom(conn)
	return err
}

func writePacket(conn net.Conn, p Packet, timeout time.Duration) error {
	if timeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(timeout))
	}
	_, err := p.WriteTo(conn)
	return err
}
