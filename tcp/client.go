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

	p := NewEmptyTHVPacket()
	funcs := []func() error{
		func() error { //send ack
			p.SetType(PacketTypeAck)
			return socket.writePacket(p)
		}, func() error {
			p.Reset()
			err = socket.readPacket(p)
			if err != nil {
				return err
			}
			if p.GetType() != PacketTypeAck {
				return fmt.Errorf("read ack failed, typ: %d", p.GetType())
			}
			//set socket id
			if len(p.Body) > 0 {
				socket.id = string(p.Body)
			}
			return nil
		}, func() error { // auth
			p.Reset()
			p.SetType(PacketTypeAuth)
			p.SetBody([]byte(c.Opt.Token))
			return socket.writePacket(p)
		}, func() error {
			p.Reset()
			err = socket.readPacket(p)
			if err != nil {
				return err
			}
			if p.GetType() != PacketTypeAuth {
				return fmt.Errorf("read auth failed, typ: %d", p.GetType())
			}
			body := string(p.GetBody())
			if body != "ok" {
				return fmt.Errorf("auth failed, body: %s", body)
			}
			return nil
		},
	}

	for _, f := range funcs {
		if err := f(); err != nil {
			socket.Close()
			return err
		}
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

			heartbeatPakcet := NewEmptyTHVPacket()
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
			p := NewEmptyTHVPacket()
			if socketErr = socket.readPacket(p); socketErr != nil {
				//todo: print out error
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
