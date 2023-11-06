package server

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/server/tcp"
)

type TcpClientOptions struct {
	RemoteAddress string
	AuthToken     string
	Agent         Agent
}

func NewTcpClient(opts *TcpClientOptions) *TcpClient {
	ret := &TcpClient{
		AutoRecconect:      true,
		reconnectTimeDelay: 15 * time.Second,
	}

	c := tcp.NewClient(&tcp.ClientOptions{
		RemoteAddress: opts.RemoteAddress,
		Token:         opts.AuthToken,
		OnMessage: func(s *tcp.Socket, p *tcp.THVPacket) {
			ptype := p.GetType()
			if ptype == PacketTypRoute {
				var head RouteHead = p.GetHead()
				if head.GetMsgTyp() == RouteTypResponse || head.GetMsgTyp() == RouteTypRespErr {
					if cb := ret.GetCallback(head.GetAskID()); cb != nil {
						cb(ret, p)
					}
				}
			}
			if ret.agent != nil {
				// ret.agent.OnSessionMessage(ret, p)
			}
		},
		OnConnStat: func(s *tcp.Socket, enable bool) {
			// if ret.OnConnectFunc != nil {
			// 	ret.OnConnectFunc(ret, enable)
			// }
			if !enable {
				if ret.AutoRecconect {
					ret.Reconnect()
				}
			}
		},
	})
	ret.Client = c
	return ret
}

type OnRespCBFunc func(*TcpClient, *tcp.THVPacket)

type TcpClient struct {
	*tcp.Client
	opts               *TcpClientOptions
	reconnectTimeDelay time.Duration
	AutoRecconect      bool
	cb                 sync.Map
	seqIndex           uint32
	agent              Agent
}

func (c *TcpClient) Reconnect() {
	err := c.Connect()
	if err != nil {
		fmt.Println("connect error:", err)
		if c.AutoRecconect {
			fmt.Println("start to reconnect")
			time.AfterFunc(c.reconnectTimeDelay, func() {
				c.Reconnect()
			})
		}
	}
}

func (c *TcpClient) MakeRequestPacket(target uint32, req proto.Message) (*tcp.THVPacket, uint32, error) {
	// msgid := calltable.GetMessageMsgID(req)
	// if msgid == 0 {
	// 	return nil, 0, fmt.Errorf("not found msgid:%v", msgid)
	// }

	// raw, err := proto.Marshal(req)
	// if err != nil {
	// 	return nil, 0, err
	// }
	// askid := c.GetAskID()
	// head := tcp.NewRoutHead()
	// head.SetAskID(askid)
	// head.SetMsgID(uint32(msgid))
	// head.SetTargetUID(target)
	// head.SetMsgTyp(tcp.RouteTypRequest)

	// ret := tcp.NewPackFrame(tcp.PacketTypRoute, head, raw)

	// return ret, askid, nil

	return nil, 0, nil
}

func SendRequestWithCB[T proto.Message](c *TcpClient, target uint32, ctx context.Context, req proto.Message, cb func(error, *TcpClient, T)) {
	go func() {
		var tresp T
		rsep := reflect.New(reflect.TypeOf(tresp).Elem()).Interface().(T)
		err := c.SyncCall(target, ctx, req, rsep)
		cb(err, c, rsep)
	}()
}

func (c *TcpClient) GroupBroadcast(group string, m proto.Message) error {
	// raw, err := proto.Marshal(m)
	// if err != nil {
	// 	return fmt.Errorf("marshal failed:%v", err)
	// }
	// msgid := calltable.GetMessageMsgID(m)
	// if msgid == 0 {
	// 	return fmt.Errorf("not found msgid:%v", msgid)
	// }
	// req := &msg.GroupBroadcastRequest{
	// 	Group:   group,
	// 	Msgid:   uint32(msgid),
	// 	Msgdata: raw,
	// }
	// resp := &msg.GroupBroadcastResponse{}
	// err = c.SyncCall(0, context.Background(), req, resp)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (c *TcpClient) SyncCall(target uint32, ctx context.Context, req proto.Message, resp proto.Message) error {
	var err error

	packet, askid, err := c.MakeRequestPacket(target, req)
	if err != nil {
		return err
	}

	res := make(chan error, 1)

	c.SetCallback(askid, func(c *TcpClient, p *tcp.THVPacket) {
		// var err error
		// defer func() {
		// 	res <- err
		// }()
		// head, err := tcp.CastRoutHead(p.GetHead())
		// if err != nil {
		// 	return
		// }
		// msgtype := head.GetMsgTyp()
		// if msgtype == tcp.RouteTypRespErr {
		// 	resperr := &msg.Error{Code: -1}
		// 	err := proto.Unmarshal(p.GetBody(), resperr)
		// 	if err != nil {
		// 		return
		// 	}
		// 	err = resperr
		// 	return
		// } else if head.GetMsgTyp() == tcp.RouteTypResponse {
		// 	gotmsgid := head.GetMsgID()
		// 	expectmsgid := uint32(calltable.GetMessageMsgID(resp))
		// 	if gotmsgid == expectmsgid {
		// 		err = proto.Unmarshal(p.GetBody(), resp)
		// 	} else {
		// 		err = fmt.Errorf("msgid not match, expect:%v, got:%v", expectmsgid, gotmsgid)
		// 	}
		// } else {
		// 	err = fmt.Errorf("unknow msgtype:%v", msgtype)
		// }
	})

	err = c.SendPacket(packet)

	if err != nil {
		c.RemoveCallback(askid)
		return err
	}

	select {
	case err = <-res:
		return err
	case <-ctx.Done():
		// dismiss callback
		c.SetCallback(askid, func(c *TcpClient, packet *tcp.THVPacket) {})
		return ctx.Err()
	}
}

func (s *TcpClient) GetAskID() uint32 {
	ret := atomic.AddUint32(&s.seqIndex, 1)
	if ret == 0 {
		ret = atomic.AddUint32(&s.seqIndex, 1)
	}
	return ret
}

func (c *TcpClient) SetCallback(askid uint32, f OnRespCBFunc) {
	c.cb.Store(askid, f)
}

func (c *TcpClient) RemoveCallback(askid uint32) {
	c.cb.Delete(askid)
}

func (c *TcpClient) GetCallback(askid uint32) OnRespCBFunc {
	if v, has := c.cb.LoadAndDelete(askid); has {
		return v.(OnRespCBFunc)
	}
	return nil
}

func (r *TcpClient) AsyncCall(target uint32, m proto.Message) error {
	return nil
	// raw, err := proto.Marshal(m)
	// if err != nil {
	// 	return fmt.Errorf("marshal %v failed:%v", proto.MessageName(m), err)
	// }

	// msgid := calltable.GetMessageMsgID(m)
	// if msgid == 0 {
	// 	return fmt.Errorf("not found msgid:%v in msg %v", msgid, proto.MessageName(m))
	// }

	// head := tcp.NewRoutHead()
	// head.SetMsgID(uint32(msgid))
	// head.SetTargetUID(target)
	// head.SetMsgTyp(tcp.RouteTypAsync)
	// return r.SendPacket(tcp.NewPackFrame(tcp.PacketTypRoute, head, raw))
}

// func (r *TcpClient) TargetEcho(target uint32, raw []byte, cb func(error, []byte)) {
// 	SendRequestWithCB(r, target, context.Background(), &msg.Echo{Body: raw}, func(err error, c *TcpClient, resp *msg.Echo) {
// 		cb(err, resp.Body)
// 	})
// }
