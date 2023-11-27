package server

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/log"
	"github.com/ajenpan/surf/msg"
	"github.com/ajenpan/surf/server/tcp"
)

type TcpClientOptions struct {
	RemoteAddress        string
	AuthToken            string
	ReconnectDelaySecond int32
	OnMessage            func(*TcpClient, *Message)
	OnStatus             func(*TcpClient, bool)
}

func NewTcpClient(opts *TcpClientOptions) *TcpClient {
	ret := &TcpClient{
		opts: opts,
	}
	imp := tcp.NewClient(&tcp.ClientOptions{
		RemoteAddress:        opts.RemoteAddress,
		Token:                opts.AuthToken,
		OnSocketMessage:      ret.OnMessage,
		OnSocketConn:         ret.OnConn,
		OnSocketDisconn:      ret.OnDisconn,
		ReconnectDelaySecond: opts.ReconnectDelaySecond,
	})
	ret.Client = imp
	return ret
}

var ErrTimeout = errors.New("timeout")

type FuncRespCallback = func(*TcpClient, *msg.ResponseMsgWrap)

type TcpClient struct {
	*tcp.Client
	opts   *TcpClientOptions
	seqidx uint32
	cb     sync.Map
}

func (s *TcpClient) Send(p *Message) error {
	return s.Client.Send(p)
}

func (s *TcpClient) SessionType() string {
	return "tcp-session"
}

func (c *TcpClient) SetCallback(askid uint32, f FuncRespCallback) {
	c.cb.Store(askid, f)
}

func (c *TcpClient) RemoveCallback(askid uint32) {
	c.cb.Delete(askid)
}

func (c *TcpClient) GetCallback(askid uint32) FuncRespCallback {
	if v, has := c.cb.LoadAndDelete(askid); has {
		return v.(FuncRespCallback)
	}
	return nil
}

func (c *TcpClient) OnMessage(socket *tcp.Client, p tcp.Packet) {
	if p.PacketType() != PacketBinaryRouteType {
		log.Error("unknow packet type:", p.PacketType())
		return
	}

	m, ok := p.(*Message)
	if !ok {
		log.Error("unknow packet type:", p.PacketType())
	}

	if m.GetMsgtype() == MsgTypeResponse {
		resp := &msg.ResponseMsgWrap{}
		err := proto.Unmarshal(m.GetBody(), resp)
		if err != nil {
			log.Error("unknow packet type:", p.PacketType())
			return
		}
		seqid := resp.GetSeqid()
		if seqid == 0 {
			log.Error("seqid is 0")
			return
		}
		if cb := c.GetCallback(seqid); cb != nil {
			cb(c, resp)
		}
	}

	if c.opts.OnMessage != nil {
		c.opts.OnMessage(c, m)
	}
}

func (c *TcpClient) OnConn(socket *tcp.Client) {
	if c.opts.OnStatus != nil {
		c.opts.OnStatus(c, true)
	}
}

func (c *TcpClient) OnDisconn(socket *tcp.Client, err error) {
	if c.opts.OnStatus != nil {
		c.opts.OnStatus(c, false)
	}
}

func (c *TcpClient) SendAsyncMsg(target uint32, req proto.Message) error {
	var err error
	body, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	wrap := &msg.AsyncMsgWrap{
		Name: string(req.ProtoReflect().Descriptor().FullName().Name()),
		Body: body,
	}
	raw, err := proto.Marshal(wrap)
	if err != nil {
		return err
	}
	msg := NewMessage()
	msg.SetBody(raw)
	msg.SetMsgtype(1)
	msg.SetUid(target)

	return c.Send(msg)
}

func (c *TcpClient) NextSeqID() uint32 {
	id := atomic.AddUint32(&c.seqidx, 1)
	if id == 0 {
		return c.NextSeqID()
	}
	return id
}

func NewTcpRespCallbackFunc[T proto.Message](f func(*TcpClient, T, error)) FuncRespCallback {
	return func(c *TcpClient, resp *msg.ResponseMsgWrap) {
		var tresp T
		rsep := reflect.New(reflect.TypeOf(tresp).Elem()).Interface().(T)
		err1 := proto.Unmarshal(resp.Body, rsep)
		if err1 != nil {
			f(c, rsep, err1)
			return
		}
		f(c, rsep, resp.Err)
	}
}

func (c *TcpClient) SendReqMsg(target uint32, req proto.Message, cb FuncRespCallback) error {
	var err error
	body, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	seqid := c.NextSeqID()
	wrap := &msg.RequestMsgWrap{
		Body:  body,
		Name:  string(req.ProtoReflect().Descriptor().FullName().Name()),
		Seqid: seqid,
	}
	raw, err := proto.Marshal(wrap)
	if err != nil {
		return err
	}

	msg := NewMessage()
	msg.SetBody(raw)
	msg.SetMsgtype(1)
	msg.SetUid(target)

	c.SetCallback(seqid, cb)
	err = c.Send(msg)
	if err != nil {
		c.RemoveCallback(seqid)
	}
	return err
}

func (c *TcpClient) SendRespMsg(target uint32, seqid uint32, resp proto.Message) error {
	var err error
	body, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	wrap := &msg.RequestMsgWrap{
		Body:  body,
		Name:  string(resp.ProtoReflect().Descriptor().FullName().Name()),
		Seqid: seqid,
	}
	raw, err := proto.Marshal(wrap)
	if err != nil {
		return err
	}
	msg := NewMessage()
	msg.SetBody(raw)
	msg.SetMsgtype(1)
	msg.SetUid(target)
	return c.Send(msg)
}

// func (c *TcpClient) SyncCall(target uint64, ctx context.Context, req proto.Message, resp proto.Message) error {
// var err error
// seqid := c.NextSeqID()

// res := make(chan error, 1)
// var askid = 0
// c.SetCallback(uint32(askid), func(c *TcpClient, p *tcp.THVPacket) {
// 	var err error
// 	defer func() {
// 		res <- err
// 	}()
// 	head, err := tcp.CastRoutHead(p.GetHead())
// 	if err != nil {
// 		return
// 	}
// 	msgtype := head.GetMsgTyp()
// 	if msgtype == tcp.RouteTypRespErr {
// 		resperr := &msg.Error{Code: -1}
// 		err := proto.Unmarshal(p.GetBody(), resperr)
// 		if err != nil {
// 			return
// 		}
// 		err = resperr
// 		return
// 	} else if head.GetMsgTyp() == tcp.RouteTypResponse {
// 		gotmsgid := head.GetMsgID()
// 		expectmsgid := uint32(calltable.GetMessageMsgID(resp))
// 		if gotmsgid == expectmsgid {
// 			err = proto.Unmarshal(p.GetBody(), resp)
// 		} else {
// 			err = fmt.Errorf("msgid not match, expect:%v, got:%v", expectmsgid, gotmsgid)
// 		}
// 	} else {
// 		err = fmt.Errorf("unknow msgtype:%v", msgtype)
// 	}
// })

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
// return nil
// }
