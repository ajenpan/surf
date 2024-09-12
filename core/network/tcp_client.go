package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ajenpan/surf/core/auth"
)

type TcpClientOptions struct {
	RemoteAddress    string
	HeatbeatInterval time.Duration

	OnConnPacket   FuncOnConnPacket
	OnConnEnable   FuncOnConnEnable
	AuthToken      []byte
	UInfo          auth.User
	ReconnectDelay time.Duration
}

func NewTcpClient(opts TcpClientOptions) *TcpClient {
	ret := &TcpClient{
		opts:   opts,
		closed: make(chan struct{}),
	}
	if ret.opts.HeatbeatInterval < time.Duration(DefaultMinTimeoutSec)*time.Second {
		ret.opts.HeatbeatInterval = time.Duration(DefaultTimeoutSec) * time.Second
	}
	return ret
}

type TcpClient struct {
	*TcpConn
	opts   TcpClientOptions
	mutex  sync.RWMutex
	closed chan struct{}
}

func (c *TcpClient) onConnEnable(enable bool) {
	if c.opts.OnConnEnable != nil {
		c.opts.OnConnEnable(c, enable)
	}
	if c.opts.ReconnectDelay > 0 {
		time.AfterFunc(time.Second*c.opts.ReconnectDelay, func() {

		})
	}
}

func (c *TcpClient) reconnect() {
	if c.opts.ReconnectDelay > 0 {
		time.AfterFunc(c.opts.ReconnectDelay*time.Second, func() {
			c.connect()
		})
	}
}

func (c *TcpClient) connect() error {
	var err error
	connraw, err := net.DialTimeout("tcp", c.opts.RemoteAddress, c.opts.HeatbeatInterval/2)
	if err != nil {
		c.reconnect()
		return err
	}
	conn, err := c.handshake(connraw)
	if err != nil {
		return err
	}
	go c.work(conn)
	return nil
}

func (c *TcpClient) work(conn *TcpConn) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	defer conn.imp.Close()

	tk := time.NewTicker(c.opts.HeatbeatInterval / 2)
	defer tk.Stop()
	defer c.onConnEnable(false)

	c.TcpConn = conn
	defer conn.Close()

	go func() {
		defer conn.Close()
		conn.writeWork()
	}()

	go func() {
		defer conn.Close()
		conn.readWork()
	}()

	c.onConnEnable(true)

	for {
		select {
		case <-c.closed:
			return nil
		case <-conn.chClosed:
			return nil
		case now := <-tk.C:
			lastSendAt := atomic.LoadInt64(&conn.lastSendAt)
			unix := now.UnixMilli()
			if unix-lastSendAt >= int64(c.opts.HeatbeatInterval.Milliseconds()) {
				pk := NewHVPacket()
				pk.Meta.SetType(PacketType_Inner)
				pk.Meta.SetSubFlag(PacketInnerSubType_Heartbeat)

				head := make([]byte, 8)
				binary.LittleEndian.PutUint64(head, uint64(time.Now().UnixMilli()))
				pk.SetHead(head)

				conn.Send(pk)
			}
		case packet, ok := <-conn.chRead:
			if !ok {
				return nil
			}
			if packet.Meta.GetType() == (PacketType_Inner) {
				switch packet.Meta.GetSubFlag() {
				case uint8(PacketInnerSubType_Heartbeat):
					// TODO:
				default:
					return nil
				}
				continue
			}

			if c.opts.OnConnPacket != nil {
				c.opts.OnConnPacket(conn, packet)
			}
		}
	}
}

func (c *TcpClient) doAckAction(conn net.Conn, body []byte) error {
	p := NewHVPacket()
	p.Meta.SetType(PacketType_Inner)
	p.Meta.SetSubFlag(PacketInnerSubType_CmdResult)
	p.SetBody(body)
	_, err := p.WriteTo(conn)
	return err
}

func (c *TcpClient) handshake(conn net.Conn) (*TcpConn, error) {
	var err error
	timeout := c.opts.HeatbeatInterval

	deadline := time.Now().Add(timeout)
	conn.SetReadDeadline(deadline)
	conn.SetWriteDeadline(deadline)

	pk := NewHVPacket()
	pk.Meta.SetType(PacketType_Inner)
	pk.Meta.SetSubFlag(PacketInnerSubType_HandShakeStart)

	if _, err := pk.WriteTo(conn); err != nil {
		return nil, err
	}

	actions := map[string][]byte{
		"auth": []byte(c.opts.AuthToken),
	}

	socketid := ""

	for {
		pk.Reset()

		_, err = pk.ReadFrom(conn)
		if err != nil {
			break
		}
		if pk.Meta.GetType() != PacketType_Inner {
			err = fmt.Errorf("packet type error %d", pk.Meta.GetType())
			break
		}
		if pk.Meta.GetSubFlag() == PacketInnerSubType_Cmd {
			name := string(pk.GetHead())
			if data, has := actions[name]; !has {
				err = fmt.Errorf("action %s not found", name)
				break
			} else {
				if err = c.doAckAction(conn, data); err != nil {
					break
				}
			}
		} else if pk.Meta.GetSubFlag() == PacketInnerSubType_HandShakeFinish {
			body := string(pk.GetBody())
			if len(body) == 0 {
				err = fmt.Errorf("ack result failed")
				break
			}
			socketid = body
			break
		} else {
			err = fmt.Errorf("invalid packet type: %d", pk.Meta.GetSubFlag())
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return newTcpConn(socketid, c.opts.UInfo, conn, c.opts.HeatbeatInterval), nil
}

func (c *TcpClient) Start() error {
	return c.connect()
}

func (c *TcpClient) Stop() error {
	return c.Close()
}

func (c *TcpClient) Close() error {
	select {
	case <-c.closed:
	default:

		close(c.closed)

		if c.TcpConn != nil {
			c.TcpConn.Close()
		}

		c.opts.ReconnectDelay = -1
	}
	return nil
}
