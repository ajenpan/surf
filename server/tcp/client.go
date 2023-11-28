package tcp

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ajenpan/surf/log"
)

type ClientOption func(*ClientOptions)

type ClientOptions struct {
	RemoteAddress        string
	Token                string
	Timeout              time.Duration
	ReconnectDelaySecond int32

	OnSocketMessage func(*Client, Packet)
	OnSocketConn    func(*Client)
	OnSocketDisconn func(*Client, error)
}

func NewClient(opts *ClientOptions) *Client {
	if opts.Timeout < time.Duration(DefaultMinTimeoutSec)*time.Second {
		opts.Timeout = time.Duration(DefaultTimeoutSec) * time.Second
	}

	ret := &Client{
		Opt: opts,
	}

	return ret
}

type Client struct {
	*Socket
	Opt   *ClientOptions
	mutex sync.Mutex
}

func doAckAction(c net.Conn, body []byte, timeout time.Duration) error {
	p := newHVPacket()
	p.SetType(PacketTypeDoAction)
	p.SetBody(body)
	return writePacket(c, timeout, p)
}

func (c *Client) doconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Socket != nil && c.Socket.IsValid() {
		c.Socket.Close()
	}

	conn, err := net.DialTimeout("tcp", c.Opt.RemoteAddress, c.Opt.Timeout)
	if err != nil {
		return err
	}
	socket, err := c.doHandShake(conn)
	if err != nil {
		conn.Close()
		return err
	}
	c.Socket = socket

	go func() {
		tk := time.NewTicker(c.Opt.Timeout / 3)
		defer tk.Stop()
		heartbeatPakcet := newHVPacket()
		heartbeatPakcet.SetType(PacketTypeHeartbeat)
		checkPos := int64(c.Opt.Timeout.Seconds() / 2)

		var writeErr error
		var readErr error

		go func() {
			writeErr = socket.writeWork()
		}()

		recvchan := make(chan Packet, 100)

		go func() {
			defer close(recvchan)
			readErr = socket.readWork(recvchan)
		}()

		defer func() {
			c.Socket.Close()
			if c.Opt.OnSocketDisconn != nil {
				c.Opt.OnSocketDisconn(c, errors.Join(writeErr, readErr))
			}
			if c.Opt.ReconnectDelaySecond > 0 {
				c.reconnect()
			}
		}()

		if c.Opt.OnSocketConn != nil {
			c.Opt.OnSocketConn(c)
		}

		for {
			select {
			case <-socket.chClosed:
				return
			case now, ok := <-tk.C:
				if !ok {
					return
				}
				nowUnix := now.Unix()
				lastSendAt := atomic.LoadInt64(&socket.lastSendAt)
				if nowUnix-lastSendAt >= checkPos {
					socket.Send(heartbeatPakcet)
				}
			case p, ok := <-recvchan:
				if !ok {
					return
				}
				switch p.PacketType() {
				case HVPacketType:
					hvPacket, ok := p.(*hvPacket)
					if !ok {
						continue
					}
					switch hvPacket.GetType() {
					case PacketTypeHeartbeat:
						log.Debug("client recv heartbeat,sid:", socket.SessionID())
					}
				default:
					if c.Opt.OnSocketMessage != nil {
						c.Opt.OnSocketMessage(c, p)
					}
				}
			}
		}
	}()
	return nil
}

func (c *Client) reconnect() {
	time.AfterFunc(time.Duration(c.Opt.ReconnectDelaySecond)*time.Second, func() {
		log.Info("start to reconnect to ", c.Opt.RemoteAddress)

		if c.IsValid() {
			log.Error("already connected")
			return
		}
		err := c.doconnect()
		if err != nil {
			log.Error("connect error:", err)
			// go on reconnect
			c.reconnect()
		}
	})
}

func (c *Client) doHandShake(conn net.Conn) (*Socket, error) {
	rwtimeout := c.Opt.Timeout

	p := newHVPacket()
	p.SetType(PacketTypeHandShake)
	if err := writePacket(conn, rwtimeout, p); err != nil {
		return nil, err
	}

	actions := map[string][]byte{
		"auth": []byte(c.Opt.Token),
	}

	socketid := ""

	var err error
	var pp *hvPacket

	for {
		pp, err = readPacketT[*hvPacket](conn, rwtimeout)
		if err != nil {
			break
		}
		if pp.GetType() == PacketTypeActionRequire {
			name := string(pp.GetBody())
			if data, has := actions[name]; !has {
				err = fmt.Errorf("action %s not found", name)
				break
			} else {
				if err = doAckAction(conn, data, rwtimeout); err != nil {
					break
				}
			}
		} else if pp.GetType() == PacketTypeAckResult {
			body := string(pp.GetBody())
			if len(body) == 0 {
				err = fmt.Errorf("ack result failed")
				break
			}
			socketid = body
			break
		} else {
			err = fmt.Errorf("invalid packet type: %d", pp.GetType())
			break
		}
	}
	if err != nil {
		return nil, err
	}
	socket := NewSocket(conn, SocketOptions{
		ID:      socketid,
		Timeout: rwtimeout,
	})

	socket.status = Connected
	return socket, nil
}

func (c *Client) Connect() error {
	if c.Opt.RemoteAddress == "" {
		return fmt.Errorf("remote address is empty")
	}
	err := c.doconnect()
	if err != nil && c.Opt.ReconnectDelaySecond > 0 {
		c.reconnect()
	}
	return err
}

func (c *Client) Close() {
	if c.Socket != nil {
		c.Socket.Close()
	}
}
