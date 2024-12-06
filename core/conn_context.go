package core

import (
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
)

type HandlerFunc func(Context)

type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. ie. the last handler is the main one.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

type Context interface {
	SendAsync(msg proto.Message) error
	Response(msg proto.Message, err error)
	FromUserID() uint32
	FromUserRole() uint16
	ConnID() string
}

type ConnContext struct {
	Conn      network.Conn
	Core      *Surf
	ReqPacket *RoutePacket
	Marshal   marshal.Marshaler
}

func (ctx *ConnContext) Response(msg proto.Message, herr error) {
	var body []byte
	var err error

	rpk := NewRoutePacket(nil)

	respmsgid := calltable.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	rpk.SetMsgId(respmsgid)
	rpk.SetSYN(ctx.ReqPacket.GetSYN())
	rpk.SetToUID(ctx.FromUserID())
	rpk.SetToURole(ctx.FromUserRole())
	rpk.SetFromUID(ctx.Core.NodeID())
	rpk.SetFromURole(ctx.Core.getServerType())
	rpk.SetMsgType(RoutePackMsgType_Response)

	if herr != nil {
		if err, ok := herr.(*errors.Error); ok {
			rpk.SetErrCode(int16(err.Code))
		} else {
			rpk.SetErrCode(-1)
		}
	}

	if msg != nil {
		body, err = ctx.Marshal.Marshal(msg)
		if err != nil {
			log.Error("response marshal error", "err", err)
			return
		}
		rpk.Body = body
	}

	err = ctx.Conn.Send(rpk.ToHVPacket())
	if err != nil {
		log.Error("response send error", "err", err)
	}
}

func (ctx *ConnContext) SendAsync(msg proto.Message) error {
	msgid := calltable.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	return ctx.Core.SendAsyncToClient(ctx.Conn, ctx.FromUserID(), ctx.FromUserRole(), msgid, msg)
}

func (ctx *ConnContext) FromUserID() uint32 {
	return ctx.ReqPacket.GetFromUID()
}

func (ctx *ConnContext) FromUserRole() uint16 {
	return ctx.ReqPacket.GetFromURole()
}

func (ctx *ConnContext) ConnID() string {
	return ctx.Conn.ConnId()
}
