package tcp

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ClientOption func(*ClientOptions)

type ClientOptions struct {
	RemoteAddress string
	OnMessage     OnMessageFunc
	OnConnStat    OnConnStatFunc
	Token         string
	Timeout       time.Duration
	SessionID     string
}

func NewClient(opts *ClientOptions) *Client {
	if opts.Timeout < time.Duration(DefaultMinTimeoutSec)*time.Second {
		opts.Timeout = time.Duration(DefaultTimeoutSec) * time.Second
	}

	socket := NewSocket(nil, SocketOptions{
		ID:      opts.SessionID,
		Timeout: opts.Timeout / 2,
	})

	ret := &Client{
		Opt:    opts,
		Socket: socket,
	}

	return ret
}

type Client struct {
	*Socket
	Opt   *ClientOptions
	mutex sync.Mutex
}

func doAckAction(c net.Conn, name string, body []byte, timeout time.Duration) error {
	p := newEmptyTHVPacket()
	p.SetType(PacketTypeDoAction)
	p.SetHead([]byte(name))
	p.SetBody(body)
	return writePacket(c, p, timeout)
}

func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Socket != nil && c.Socket.conn != nil {
		c.Socket.conn.Close()
		c.Socket.conn = nil
	}

	if c.Opt.RemoteAddress == "" {
		return fmt.Errorf("remote address is empty")
	}

	rwtimeout := c.Opt.Timeout

	conn, err := net.DialTimeout("tcp", c.Opt.RemoteAddress, c.Opt.Timeout)
	if err != nil {
		return err
	}

	// socket := NewSocket(conn, SocketOptions{
	// 	Timeout: c.Opt.Timeout,
	// })
	socketid := ""
	p := newEmptyTHVPacket()
	p.SetType(PacketTypeAck)
	if err := writePacket(conn, p, rwtimeout); err != nil {
		return err
	}

	actions := map[string][]byte{
		"auth": []byte(c.Opt.Token),
	}

	for {
		p.Reset()
		if err = readPacket(conn, p, rwtimeout); err != nil {
			break
		}
		if p.GetType() == PacketTypeActionRequire {
			name := string(p.GetHead())
			if data, has := actions[name]; !has {
				err = fmt.Errorf("action %s not found", name)
				break
			} else {
				if err = doAckAction(conn, name, data, rwtimeout); err != nil {
					break
				}
			}
		} else if p.GetType() == PacketTypeAckResult {
			head := string(p.GetHead())
			body := string(p.GetBody())
			if head != "ok" {
				err = fmt.Errorf("ack result failed, head: %s, body: %s", head, body)
				break
			}
			if len(body) > 0 {
				socketid = body
			}
			break
		} else {
			err = fmt.Errorf("invalid packet type: %d", p.GetType())
			break
		}
	}

	if err != nil {
		return err
	}

	//here is connect finished
	c.Socket.conn = conn
	c.Socket.id = socketid

	socket := c.Socket

	go func() {
		defer socket.Close()

		go socket.writeWork()

		if c.Opt.OnConnStat != nil {
			c.Opt.OnConnStat(c.Socket, true)

			defer func() {
				c.Opt.OnConnStat(c.Socket, false)
			}()
		}

		go func() {
			tk := time.NewTicker(c.Opt.Timeout / 3)
			defer tk.Stop()

			heartbeatPakcet := newEmptyTHVPacket()
			heartbeatPakcet.SetType(PacketTypeHeartbeat)

			checkPos := int64(c.Opt.Timeout.Seconds() / 2)

			for {
				select {
				case <-tk.C:
					nowUnix := time.Now().Unix()
					lastSendAt := atomic.LoadInt64(&socket.lastSendAt)
					if nowUnix-lastSendAt >= checkPos {
						socket.chSend <- heartbeatPakcet
					}
				case <-socket.chClosed:
					return
				}
			}
		}()

		var socketErr error = nil
		for {
			p := newEmptyTHVPacket()
			if socketErr = socket.readPacket(p); socketErr != nil {
				fmt.Println(err)
				break
			}

			typ := p.GetType()
			if typ > PacketTypeInnerEndAt_ {
				if c.Opt.OnMessage != nil {
					c.Opt.OnMessage(socket, p)
				}
			}
		}
		fmt.Println("socket read error: ", socketErr)
	}()
	return nil
}

func (c *Client) Close() {
	if c.Socket != nil {
		c.Socket.Close()
	}
}
