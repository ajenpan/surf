package server

import (
	"errors"
	"reflect"
	"sync/atomic"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/log"
	"github.com/ajenpan/surf/msg"
	"github.com/ajenpan/surf/server/tcp"
)

type TcpClientOptions struct {
	RemoteAddress     string
	AuthToken         string
	ReconnectDelaySec int32
	OnMessage         func(*TcpClient, *MsgWraper)
	OnStatus          func(*TcpClient, bool)
}

func NewTcpClient(opts *TcpClientOptions) *TcpClient {
	ret := &TcpClient{
		opts: opts,
	}
	imp := tcp.NewClient(&tcp.ClientOptions{
		RemoteAddress:     opts.RemoteAddress,
		Token:             opts.AuthToken,
		OnSocketMessage:   ret.OnMessage,
		OnSocketConn:      ret.OnConn,
		OnSocketDisconn:   ret.OnDisconn,
		ReconnectDelaySec: opts.ReconnectDelaySec,
	})
	ret.imp = imp
	return ret
}

var ErrTimeout = errors.New("timeout")

type TcpClient struct {
	TcpSession

	imp  *tcp.Client
	opts *TcpClientOptions
}

// func (s *TcpClient) Send(p *MsgWraper) error {
// 	return s.Client.Send(p)
// }

func (c *TcpClient) Connect() error {
	return c.imp.Connect()
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
	if p.PacketType() != PacketTypeRouteMsgWraper {
		log.Error("unknow packet type:", p.PacketType())
		return
	}

	m, ok := p.(*MsgWraper)
	if !ok {
		log.Error("unknow packet type:", p.PacketType())
		return
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

func (c *TcpClient) OnConn(cc *tcp.Client) {
	c.Socket = &cc.Socket
	if c.opts.OnStatus != nil {
		c.opts.OnStatus(c, true)
	}
}

func (c *TcpClient) OnDisconn(socket *tcp.Client, err error) {
	if c.opts.OnStatus != nil {
		c.opts.OnStatus(c, false)
	}
}

func (c *TcpClient) NextSeqID() uint32 {
	id := atomic.AddUint32(&c.seqidx, 1)
	if id == 0 {
		return c.NextSeqID()
	}
	return id
}

func NewTcpRespCallbackFunc[T proto.Message](f func(Session, T, error)) FuncRespCallback {
	return func(c Session, resp *msg.ResponseMsgWrap) {
		var tresp T
		rsep := reflect.New(reflect.TypeOf(tresp).Elem()).Interface().(T)
		err1 := proto.Unmarshal(resp.Body, rsep)
		if err1 != nil {
			f(c, rsep, err1)
			return
		}

		var err error
		if !(resp.Err == nil || resp.Err.Code == 0) {
			err = resp.Err
		}

		f(c, rsep, err)
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

	msg := NewMsgWraper()
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
	msg := NewMsgWraper()
	msg.SetBody(raw)
	msg.SetMsgtype(1)
	msg.SetUid(target)
	return c.Send(msg)
}
