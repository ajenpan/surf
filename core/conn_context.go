package core

import (
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
)

type Context interface {
	Response(msg proto.Message, err error)
	Caller() uint32
}

type connContext struct {
	Conn      network.Conn
	Core      *Surf
	ReqPacket *network.RoutePacket
	caller    uint32
	Marshal   marshal.Marshaler
}

func (ctx *connContext) Response(msg proto.Message, herr error) {
	var body []byte
	var err error

	rpk := network.NewRoutePacket(nil)
	rpk.Head.CopyFrom(ctx.ReqPacket.Head)

	respmsgid := calltable.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	rpk.SetMsgId(respmsgid)

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
			log.Error(err)
			return
		}
		rpk.Body = body
	}

	pk := rpk.ToHVPacket()
	err = ctx.Conn.Send(pk)

	if err != nil {
		log.Error(err)
	}
}

func (ctx *connContext) SendAsync(msg proto.Message) error {
	msgid := calltable.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	return ctx.Core.SendAsyncToClient(ctx.Conn, ctx.caller, msgid, msg)
}

func (ctx *connContext) Caller() uint32 {
	return ctx.caller
}
