package server

import (
	"context"
	"crypto/rsa"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/server/tcp"
	"github.com/ajenpan/surf/utils/marshal"
)

type TcpClientOptions struct {
	RemoteAddress    string
	AuthToken        string
	AuthPublicKey    *rsa.PublicKey
	AutoRecconect    bool
	OnSessionMessage FuncOnSessionMessage
	OnSessionStatus  FuncOnSessionStatus
}

func NewTcpClient(opts *TcpClientOptions) *TcpClient {
	auth := func(b []byte) (*tcp.UserInfo, error) {
		uid, uname, role, err := VerifyToken(opts.AuthPublicKey, string(b))
		if err != nil {
			return nil, err
		}
		return &tcp.UserInfo{
			UId:   uid,
			UName: uname,
			Role:  role,
		}, nil
	}

	uinfo, err := auth([]byte(opts.AuthToken))
	if err != nil {
		return nil
	}

	ret := &TcpClient{
		opts:               opts,
		reconnectTimeDelay: 15 * time.Second,
	}
	c := tcp.NewClient(&tcp.ClientOptions{
		RemoteAddress: opts.RemoteAddress,
		Token:         opts.AuthToken,
		OnMessage:     ret.OnTcpMessage,
		OnConnStat:    ret.OnTcpConn,
	})
	c.UserInfo = uinfo
	ret.imp = c
	return ret
}

type OnRespCBFunc func(*TcpClient, *tcp.THVPacket)

type TcpClient struct {
	imp                *tcp.Client
	opts               *TcpClientOptions
	reconnectTimeDelay time.Duration
	seqIndex           uint32
	cb                 sync.Map
}

func (c *TcpClient) Reconnect() {
	err := c.imp.Connect()
	if err != nil {
		fmt.Println("connect error:", err)
		if c.opts.AutoRecconect {
			fmt.Println("start to reconnect")
			time.AfterFunc(c.reconnectTimeDelay, func() {
				c.Reconnect()
			})
		}
	}
}

func (s *TcpClient) OnTcpMessage(socket *tcp.Socket, p *tcp.THVPacket) {
	sess := loadTcpSession(socket)
	if sess == nil {
		return
	}
	msg, err := sess.pkg2msg(p)
	if err != nil {
		return
	}

	if s.opts.OnSessionMessage != nil {
		s.opts.OnSessionMessage(sess, msg)
	}
}

func (s *TcpClient) Start() error {
	return s.imp.Connect()
}

func (s *TcpClient) Stop() error {
	s.imp.Close()
	return nil
}

func (s *TcpClient) OnTcpConn(socket *tcp.Socket, enable bool) {
	var sess *TcpSession
	if enable {
		sess = &TcpSession{
			Socket: socket,
		}
		socket.Meta.Store(tcpSessionKey, sess)
	} else {
		sess = loadTcpSession(socket)
		socket.Meta.Delete(tcpSessionKey)
	}

	if sess != nil && s.opts.OnSessionStatus != nil {
		s.opts.OnSessionStatus(sess, enable)
	}

	if !enable {
		if s.opts.AutoRecconect {
			s.Reconnect()
		}
	}
}

func (c *TcpClient) MakeRequestPacket(target uint64, req proto.Message) (*Message, error) {
	m := &marshal.ProtoMarshaler{}
	body, err := m.Marshal(req)
	if err != nil {
		return nil, err
	}

	msg := NewMessage()
	msg.Head.MsgName = string(proto.MessageName(req).Name())
	msg.Body = body
	msg.Head.Seq = c.GetAskID()
	msg.Head.SourceUid = c.imp.UID()
	msg.Head.TargetUid = target
	msg.Head.MsgType = 1
	return msg, nil
}

func SendRequestWithCB[T proto.Message](c *TcpClient, target uint64, ctx context.Context, req proto.Message, cb func(error, *TcpClient, T)) {
	go func() {
		var tresp T
		rsep := reflect.New(reflect.TypeOf(tresp).Elem()).Interface().(T)
		err := c.SyncCall(target, ctx, req, rsep)
		cb(err, c, rsep)
	}()
}

func (c *TcpClient) SyncCall(target uint64, ctx context.Context, req proto.Message, resp proto.Message) error {
	// var err error

	// msg, err := c.MakeRequestPacket(target, req)
	// if err != nil {
	// 	return err
	// }
	// askid := msg.Head.Seq

	// res := make(chan error, 1)
	var askid = 0
	c.SetCallback(uint32(askid), func(c *TcpClient, p *tcp.THVPacket) {
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

	// err = c.imp.SendPacket(packet)

	// if err != nil {
	// 	c.RemoveCallback(askid)
	// 	return err
	// }

	// select {
	// case err = <-res:
	// 	return err
	// case <-ctx.Done():
	// 	// dismiss callback
	// 	c.SetCallback(askid, func(c *TcpClient, packet *tcp.THVPacket) {})
	// 	return ctx.Err()
	// }
	return nil
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
