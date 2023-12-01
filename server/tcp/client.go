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
	RemoteAddress     string
	Token             string
	Timeout           time.Duration
	ReconnectDelaySec int32

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
		Socket: Socket{
			chWrite:  make(chan Packet, 100),
			chClosed: make(chan struct{}),
			timeOut:  opts.Timeout,
			status:   Disconnected,
		},
	}
	return ret
}

type Client struct {
	Socket
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

	if c.IsValid() {
		c.Close()
	}

	conn, err := net.DialTimeout("tcp", c.Opt.RemoteAddress, c.Opt.Timeout)
	if err != nil {
		return err
	}
	err = c.doHandShake(conn)
	if err != nil {
		conn.Close()
		return err
	}

	go func() {
		socket := &c.Socket

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
			if c.Opt.ReconnectDelaySec > 0 {
				c.reconnect()
			}
		}()

		if c.Opt.OnSocketConn != nil {
			c.Opt.OnSocketConn(c)
		}

		// start heartbeat
		tkcheck := c.Opt.Timeout / 4
		tk := time.NewTicker(tkcheck)
		defer tk.Stop()
		heartbeatPakcet := newHVPacket()
		heartbeatPakcet.SetType(PacketTypeHeartbeat)
		checkPos := (int64)(tkcheck.Seconds())

		for {
			select {
			case <-socket.chClosed:
				return
			case now, ok := <-tk.C:
				if !ok {
					return
				}
				nowUnix := now.Unix()
				lastRecvAt := atomic.LoadInt64(&socket.lastRecvAt)
				lastSendAt := atomic.LoadInt64(&socket.lastSendAt)
				idletime := nowUnix - min(lastRecvAt, lastSendAt)
				if idletime >= checkPos {
					//log.Debugf("client send heartbeat,sid:%v %v ,now:%v %v %v", socket.SessionID(), int(time.Duration(idletime).Seconds()), nowUnix, socket.lastSendAt, socket.lastRecvAt)
					if err := socket.Send(heartbeatPakcet); err != nil {
						log.Error("send heartbeat:", err)
					}
				} else {
					//log.Debugf("client miss heartbeat,sid:%v %v ,now:%v %v %v", socket.SessionID(), int(time.Duration(idletime).Seconds()), nowUnix, socket.lastSendAt, socket.lastRecvAt)
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
	time.AfterFunc(time.Duration(c.Opt.ReconnectDelaySec)*time.Second, func() {
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

func (c *Client) doHandShake(conn net.Conn) error {
	rwtimeout := c.Opt.Timeout

	p := newHVPacket()
	p.SetType(PacketTypeHandShake)
	if err := writePacket(conn, rwtimeout, p); err != nil {
		return err
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
		} else if pp.GetType() == PacketTypeAckSuccess {
			socketid = string(pp.GetBody())
			break
		} else if pp.GetType() == PacketTypeAckFailure {
			err = fmt.Errorf("ack failure: %v", string(pp.GetBody()))
			break
		} else {
			err = fmt.Errorf("invalid packet type: %d", pp.GetType())
			break
		}
	}
	if err != nil {
		return err
	}

	c.id = socketid
	c.status = Connected
	c.setconn(conn)
	return nil
}

func (c *Client) Connect() error {
	if c.Opt.RemoteAddress == "" {
		return fmt.Errorf("remote address is empty")
	}
	err := c.doconnect()
	if err != nil && c.Opt.ReconnectDelaySec > 0 {
		c.reconnect()
	}
	return err
}
