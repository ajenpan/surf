package network

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ajenpan/surf/core/auth"
	ws "github.com/gorilla/websocket"
)

type WSConn struct {
	auth.User
	sync.Map

	imp      *ws.Conn
	status   ConnStatus
	chClosed chan struct{}
	timeOut  time.Duration

	chWrite chan *HVPacket
	chRead  chan *HVPacket

	id string
}

func (c *WSConn) Send(p *HVPacket) error {
	return c.writePacket(p)
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

func (c *WSConn) writePacket(h *HVPacket) error {
	writer, err := c.imp.NextWriter(ws.BinaryMessage)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = h.WriteTo(writer)
	return err

}

func (c *WSConn) readPacket() (*HVPacket, error) {
	_, reader, err := c.imp.NextReader()
	if err != nil {
		return nil, err
	}
	pk := NewHVPacket()
	_, err = pk.ReadFrom(reader)
	return pk, err
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
			c.imp.SetWriteDeadline(time.Now().Add(c.timeOut))
			err := c.writePacket(p)
			if err != nil {
				return err
			}
		}
	}
}

func (c *WSConn) readWork() error {
	for {
		c.imp.SetReadDeadline(time.Now().Add(c.timeOut))
		pk, err := c.readPacket()
		if err != nil {
			return nil
		}
		select {
		case <-c.chClosed:
			return nil
		case c.chRead <- pk:

		}
	}
}
