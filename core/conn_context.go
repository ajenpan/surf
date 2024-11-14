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

	respmsgid := calltable.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	rpk.SetMsgId(respmsgid)
	rpk.SetNodeId(ctx.Core.GetNodeId())
	rpk.SetSvrType(ctx.Core.GetServerType())
	rpk.SetSYN(ctx.ReqPacket.GetSYN())
	rpk.SetClientId(ctx.ReqPacket.GetClientId())
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
	return ctx.Core.SendAsyncToClient(ctx.Conn, ctx.caller, msgid, msg)
}

func (ctx *connContext) Caller() uint32 {
	return ctx.caller
}
