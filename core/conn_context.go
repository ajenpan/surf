package core

import (
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
)

type Context interface {
	SendAsync(msg proto.Message) error
	Response(msg proto.Message, err error)

	UserID() uint32
	UserRole() uint32
}

type connContext struct {
	Conn      network.Conn
	Core      *Surf
	ReqPacket *network.RoutePacket
	uid       uint32
	urole     uint32
	Marshal   marshal.Marshaler
}

func (ctx *connContext) Response(msg proto.Message, herr error) {
	var body []byte
	var err error

	rpk := network.NewRoutePacket(nil)

	respmsgid := calltable.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	rpk.SetMsgId(respmsgid)
	rpk.SetToUID(ctx.ReqPacket.GetFromUID())
	rpk.SetToURole(ctx.ReqPacket.GetFromURole())
	rpk.SetSYN(ctx.ReqPacket.GetSYN())
	rpk.SetFromUID(ctx.Core.GetNodeId())
	rpk.SetFromURole(ctx.Core.GetServerType())
	rpk.SetMsgType(network.RoutePackMsgType_Response)

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

	err = ctx.Conn.Send(rpk.ToHVPacket())
	if err != nil {
		log.Error(err)
	}
}

func (ctx *connContext) SendAsync(msg proto.Message) error {
	msgid := calltable.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	return ctx.Core.SendAsyncToClient(ctx.Conn, ctx.uid, msgid, msg)
}

func (ctx *connContext) UserID() uint32 {
	return ctx.uid
}

func (ctx *connContext) UserRole() uint32 {
	return ctx.urole
}
