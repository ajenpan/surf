package network

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	ws "github.com/gorilla/websocket"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
)

type WSClientOptions struct {
	RemoteAddress    string
	HeatbeatInterval time.Duration

	OnConnPacket   FuncOnConnPacket
	OnConnStatus   FuncOnConnStatus
	AuthToken      []byte
	UInfo          auth.User
	ReconnectDelay time.Duration
}

type WSClientOption func(*WSClientOptions)

func NewWSClient(opts WSClientOptions) *WSClient {
	ret := &WSClient{
		opts:   opts,
		closed: make(chan struct{}),
	}
	if ret.opts.HeatbeatInterval < time.Duration(DefaultMinTimeoutSec/2)*time.Second {
		ret.opts.HeatbeatInterval = time.Duration(DefaultTimeoutSec/2) * time.Second
	}
	return ret
}

type WSClient struct {
	*WSConn
	opts   WSClientOptions
	mutex  sync.RWMutex
	closed chan struct{}

	timefix int64
}

func (c *WSClient) GetSvrTimeFix() int64 {
	return atomic.LoadInt64(&c.timefix)
}

func (c *WSClient) Start() error {
	return c.connect()
}

func (c *WSClient) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
		c.opts.ReconnectDelay = -1
		if c.WSConn != nil {
			return c.WSConn.Close()
		}
	}
	return nil
}

func (c *WSClient) onConnStatus(enable bool) {
	if c.opts.OnConnStatus != nil {
		c.opts.OnConnStatus(c, enable)
	}
	if !enable {
		c.reconnect()
	}
}

func (c *WSClient) reconnect() {
	if c.opts.ReconnectDelay > 0 {
		time.AfterFunc(c.opts.ReconnectDelay, func() {
			c.connect()
		})
	}
}

func (c *WSClient) connect() error {
	var err error
	dialer := &ws.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: c.opts.HeatbeatInterval,
	}

	connraw, _, err := dialer.Dial(c.opts.RemoteAddress, nil)
	if err != nil {
		c.reconnect()
		return err
	}

	conn, err := c.handshake(connraw)
	if err != nil {
		log.Errorf("connect handshake err:%v", err)
		return err
	}

	go c.work(conn)
	return nil
}

func (c *WSClient) work(conn *WSConn) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.onConnStatus(true)
	defer c.onConnStatus(false)

	tki := time.Duration(c.opts.HeatbeatInterval.Seconds()/3) * time.Second
	if tki < time.Second*2 {
		tki = time.Second * 2
	}

	tk := time.NewTicker(tki)
	defer tk.Stop()

	conn.status = Connected
	c.WSConn = conn
	defer conn.Close()

	go func() {
		defer conn.Close()
		conn.writeWork()
	}()

	go func() {
		defer conn.Close()
		err := conn.readWork()
		if err != nil {
			log.Errorf("readWork err: %v", err)
		}
	}()

	for {
		select {
		case <-c.closed:
			return nil
		case <-conn.chClosed:
			return nil
		case now := <-tk.C:
			lastSendAt := atomic.LoadInt64(&conn.lastSendAt)
			lastRecvAt := atomic.LoadInt64(&conn.lastRecvAt)
			unix := now.UnixMilli()

			ds := int64(0)
			if lastSendAt > lastRecvAt {
				ds = unix - lastRecvAt
			} else {
				ds = unix - lastSendAt
			}

			if ds >= int64(c.opts.HeatbeatInterval.Milliseconds())/2 {
				pk := NewHVPacket()
				pk.Meta.SetType(PacketType_Inner)
				pk.Meta.SetSubFlag(PacketInnerSubType_Heartbeat)

				head := make([]byte, 8)
				binary.LittleEndian.PutUint64(head, uint64(time.Now().UnixMilli()))
				pk.SetHead(head)

				conn.Send(pk)

				log.Infof("send Heartbeat")
			} else {
				log.Infof("pass heartbeat ds:%d, sendat:%d, recvAt:%d, now:%s", ds, lastSendAt, lastRecvAt, now.Format(time.StampMicro))
			}
		case packet, ok := <-conn.chRead:
			if !ok {
				return nil
			}
			if packet.Meta.GetType() == (PacketType_Inner) {
				switch packet.Meta.GetSubFlag() {
				case uint8(PacketInnerSubType_Heartbeat):
					now := time.Now().UnixMilli()
					sendat := int64(binary.LittleEndian.Uint64(packet.GetHead()))
					svrtime := (int64)(binary.LittleEndian.Uint64(packet.GetBody()))
					fix := now - (svrtime - (now-sendat)/2)
					atomic.StoreInt64(&c.timefix, fix)

					log.Infof("recv Heartbeat")
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

func (c *WSClient) doAckAction(conn *ws.Conn, body []byte) error {
	p := NewHVPacket()
	p.Meta.SetType(PacketType_Inner)
	p.Meta.SetSubFlag(PacketInnerSubType_CmdResult)
	p.SetBody(body)
	return wsconnWritePacket(conn, p)
}

func (c *WSClient) handshake(conn *ws.Conn) (*WSConn, error) {
	var err error
	timeout := c.opts.HeatbeatInterval

	deadline := time.Now().Add(timeout)
	conn.SetReadDeadline(deadline)
	conn.SetWriteDeadline(deadline)

	pk := NewHVPacket()
	pk.Meta.SetType(PacketType_Inner)
	pk.Meta.SetSubFlag(PacketInnerSubType_HandShakeStart)
	if err := wsconnWritePacket(conn, pk); err != nil {
		return nil, err
	}

	actions := map[string][]byte{
		"auth": []byte(c.opts.AuthToken),
	}

	socketid := ""

	for {
		pk, err = wsconnReadPacket(conn)
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
	return newWSConn(socketid, c.opts.UInfo, conn, c.opts.HeatbeatInterval*2), nil
}
