package network

import (
	"sync/atomic"
	"time"

	ws "github.com/gorilla/websocket"
)

func newWSConn(id string, uinfo User, conn *ws.Conn, rwtimeout time.Duration) *WSConn {
	if rwtimeout.Seconds() < float64(DefaultHeartbeatSec) {
		rwtimeout = time.Duration(DefaultHeartbeatSec*2) * time.Second
	}
	ret := &WSConn{
		id:         id,
		imp:        conn,
		rwtimeout:  rwtimeout,
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

func wsconnWritePacket(conn *ws.Conn, p *HVPacket) error {
	writer, err := conn.NextWriter(ws.BinaryMessage)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = p.WriteTo(writer)
	return err
}

func wsconnReadPacket(conn *ws.Conn) (*HVPacket, error) {
	_, reader, err := conn.NextReader()
	if err != nil {
		return nil, err
	}
	pk := NewHVPacket()
	_, err = pk.ReadFrom(reader)
	return pk, err
}

type WSConn struct {
	userInfo
	SYNGenerator

	imp       *ws.Conn
	status    ConnStatus
	chClosed  chan struct{}
	rwtimeout time.Duration

	chWrite chan *HVPacket
	chRead  chan *HVPacket

	id         string
	userdata   any
	lastSendAt int64
	lastRecvAt int64
}

func (c *WSConn) SetUserData(d any) {
	c.userdata = d
}

func (c *WSConn) GetUserData() any {
	return c.userdata
}

func (c *WSConn) Send(p *HVPacket) error {
	if !c.Enable() {
		return ErrDisconn
	}
	select {
	case c.chWrite <- p:
	case <-c.chClosed:
		return ErrDisconn
	}
	return nil
}

func (c *WSConn) ConnID() string {
	return c.id
}

func (s *WSConn) RemoteAddr() string {
	if !s.Enable() {
		return ""
	}
	return s.imp.RemoteAddr().String()
}

func (c *WSConn) Close() error {
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

func (c *WSConn) Enable() bool {
	return c.Status() == Connected
}

func (c *WSConn) Status() ConnStatus {
	return ConnStatus(atomic.LoadInt32((*int32)(&c.status)))
}

func (c *WSConn) writePacket(p *HVPacket) error {
	return wsconnWritePacket(c.imp, p)
}

func (c *WSConn) readPacket() (*HVPacket, error) {
	return wsconnReadPacket(c.imp)
}

func (c *WSConn) writeWork() error {
	for {
		select {
		case <-c.chClosed:
			return nil
		case p, ok := <-c.chWrite:
			if !ok {
				return nil
			}
			c.imp.SetWriteDeadline(time.Now().Add(c.rwtimeout))
			err := c.writePacket(p)
			if err != nil {
				return err
			}
			atomic.SwapInt64(&c.lastSendAt, time.Now().UnixMilli())
		}
	}
}

func (c *WSConn) readWork() error {
	for {
		rdl := time.Now().Add(c.rwtimeout)
		c.imp.SetReadDeadline(rdl)
		pk, err := c.readPacket()
		if err != nil {
			return err
		}
		atomic.SwapInt64(&c.lastRecvAt, time.Now().UnixMilli())
		select {
		case <-c.chClosed:
			return nil
		case c.chRead <- pk:
		}
	}
}
