package websocket

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func NewSocket(id string, c net.Conn) *Socket {
	ret := &Socket{
		id:       id,
		conn:     c,
		chSend:   make(chan *Packet, 10),
		chClosed: make(chan bool),
		state:    Connected,
	}
	return ret
}

var staticIdx uint64

func nextID() string {
	idx := atomic.AddUint64(&staticIdx, 1)
	if idx == 0 {
		idx = atomic.AddUint64(&staticIdx, 1)
	}
	return fmt.Sprintf("ws_%v_%v", idx, time.Now().Unix())
}

type SocketStat int32

const (
	Disconnected SocketStat = iota
	Connected    SocketStat = iota
)

type OnMessageFunc func(*Socket, *Packet)
type OnConnStatFunc func(*Socket, SocketStat)
type NewIDFunc func() string

type SocketOptions struct {
	ID string
}

type Socket struct {
	conn     net.Conn   // low-level conn fd
	state    SocketStat // current state
	id       string
	chSend   chan *Packet // push message queue
	chClosed chan bool

	lastSendAt uint64
	lastRecvAt uint64
}

func (s *Socket) ID() string {
	return s.id
}

func (s *Socket) SendPacket(p *Packet) error {
	if atomic.LoadInt32((*int32)(&s.state)) == int32(Disconnected) {
		return errors.New("send packet failed, the socket is disconnected")
	}
	s.chSend <- p
	return nil
}

func (s *Socket) Send(msgid uint32, body []byte) error {
	p := &Packet{}
	p.Body = body
	return s.SendPacket(p)
}

func (s *Socket) Close() {
	stat := atomic.SwapInt32((*int32)(&s.state), int32(Disconnected))
	if stat == int32(Disconnected) {
		return
	}

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	close(s.chSend)
	close(s.chClosed)
}

// returns the remote network address.
func (s *Socket) RemoteAddr() net.Addr {
	if s == nil {
		return nil
	}
	return s.conn.RemoteAddr()
}

func (s *Socket) LocalAddr() net.Addr {
	if s == nil {
		return nil
	}
	return s.conn.LocalAddr()
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

func (s *Socket) readPacket(p *Packet) error {
	if s.Status() == Disconnected {
		return errors.New("recv packet failed, the socket is disconnected")
	}

	var err error
	p.Body, _, err = wsutil.ReadClientData(s.conn)

	if err != nil {
		return err
	}

	atomic.StoreUint64(&s.lastRecvAt, uint64(time.Now().Unix()))
	return nil
}

func (s *Socket) writePacket(p *Packet) error {
	if s.Status() == Disconnected {
		return errors.New("recv packet failed, the socket is disconnected")
	}

	err := wsutil.WriteServerMessage(s.conn, ws.OpText, p.Body)
	if err != nil {
		return err
	}

	atomic.StoreUint64(&s.lastSendAt, uint64(time.Now().Unix()))
	return nil
}
