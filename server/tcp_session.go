package server

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/ajenpan/surf/msg"
	"github.com/ajenpan/surf/server/tcp"
	"google.golang.org/protobuf/proto"
)

var tcpSessionKey = &struct{}{}

func newTcpSession(socket *tcp.Socket) *TcpSession {
	ret := &TcpSession{
		Socket: socket,
	}
	socket.Meta.Store(tcpSessionKey, ret)
	return ret
}

func loadTcpSession(socket *tcp.Socket) *TcpSession {
	v, ok := socket.Meta.Load(tcpSessionKey)
	if !ok {
		return nil
	}
	return v.(*TcpSession)
}

type TcpSession struct {
	*tcp.Socket

	seqidx uint32
	cb     sync.Map
}

func (c *TcpSession) NextSeqID() uint32 {
	id := atomic.AddUint32(&c.seqidx, 1)
	if id == 0 {
		return c.NextSeqID()
	}
	return id
}

func (c *TcpSession) SetCallback(askid uint32, f FuncRespCallback) {
	c.cb.Store(askid, f)
}

func (c *TcpSession) RemoveCallback(askid uint32) {
	c.cb.Delete(askid)
}

func (c *TcpSession) GetCallback(askid uint32) FuncRespCallback {
	if v, has := c.cb.LoadAndDelete(askid); has {
		return v.(FuncRespCallback)
	}
	return nil
}

func (s *TcpSession) SessionType() string {
	return "tcp-session"
}

func (s *TcpSession) Send(p *MsgWraper) error {
	return s.Socket.Send(p)
}

func (s *TcpSession) Close() {
	if s.Socket != nil {
		s.Socket.Close()
	}
}

func (s *TcpSession) SendAsync(uid uint32, a proto.Message) error {
	wrap := NewMsgWraper()
	wrap.SetMsgtype(MsgTypeAsync)
	wrap.SetUid(uid)

	raw, err := proto.Marshal(a)
	if err != nil {
		return err
	}

	async := &AsyncMsg{
		Body: raw,
		Name: string(proto.MessageName(a).Name()),
	}

	body, err := proto.Marshal(async)
	if err != nil {
		return err
	}

	wrap.SetBody(body)
	return s.Send(wrap)
}

func (s *TcpSession) SendRequest(target uint32, req proto.Message, cb FuncRespCallback) error {
	var err error
	body, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	seqid := s.NextSeqID()
	wrap := &RequestMsg{
		Body:  body,
		Name:  string(proto.MessageName(req).Name()),
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

	s.SetCallback(seqid, cb)
	err = s.Send(msg)
	if err != nil {
		s.RemoveCallback(seqid)
	}
	return err
}

func (s *TcpSession) SendResponse(target uint32, req *RequestMsg, resp proto.Message, err error) error {
	respwrap := &ResponseMsg{
		Name:  string(proto.MessageName(resp).Name()),
		Seqid: req.Seqid,
	}
	var merr error
	respwrap.Body, merr = proto.Marshal(resp)
	if merr != nil {
		err = errors.Join(err, merr)
	}

	if err != nil {
		if serr, ok := err.(*ResponseError); ok {
			respwrap.Err = (*msg.Error)(serr)
		} else {
			respwrap.Err = (*msg.Error)(&ResponseError{
				Code:   -1,
				Detail: err.Error(),
			})
		}
	}

	body, err := proto.Marshal((proto.Message)(respwrap))
	if err != nil {
		return err
	}

	wrap := NewMsgWraper()
	wrap.SetMsgtype(MsgTypeResponse)
	wrap.SetUid(target)
	wrap.SetBody(body)
	return s.Send(wrap)
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
