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

func doAckAction(s *Socket, name string, body []byte) error {
	p := newEmptyTHVPacket()
	p.SetType(PacketTypeDoAction)
	p.SetHead([]byte(name))
	p.SetBody(body)
	return s.writePacket(p)
}

func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Socket != nil {
		c.Socket.Close()
		c.Socket = nil
	}

	if c.Opt.RemoteAddress == "" {
		return fmt.Errorf("remote address is empty")
	}

	conn, err := net.DialTimeout("tcp", c.Opt.RemoteAddress, c.Opt.Timeout)
	if err != nil {
		return err
	}

	socket := NewSocket(conn, SocketOptions{
		Timeout: c.Opt.Timeout,
	})

	p := newEmptyTHVPacket()
	p.SetType(PacketTypeAck)
	if err := socket.writePacket(p); err != nil {
		return err
	}

	actions := map[string][]byte{
		"auth": []byte(c.Opt.Token),
	}

	for {
		p.Reset()
		if err = socket.readPacket(p); err != nil {
			break
		}
		if p.GetType() == PacketTypeActionRequire {
			name := string(p.GetHead())
			if data, has := actions[name]; !has {
				err = fmt.Errorf("action %s not found", name)
				break
			} else {
				if err = doAckAction(socket, name, data); err != nil {
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
				socket.id = body
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
	c.Socket = socket

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
