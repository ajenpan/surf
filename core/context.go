package core

import (
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/core/utils/marshal"
)

type Context interface {
	Response(msg proto.Message, err error)
	Caller() uint32
}

type context struct {
	Conn    network.Conn
	Core    *Surf
	Pk      *network.HVPacket
	caller  uint32
	Marshal marshal.Marshaler
}

func (ctx *context) Response(msg proto.Message, herr error) {
	inHead := network.RoutePacketHead(ctx.Pk.GetHead())
	var body []byte
	var err error

	respmsgid := calltable.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	inHead.SetMsgId(respmsgid)

	if herr != nil {
		if err, ok := herr.(*errors.Error); ok {
			inHead.SetErrCode(int16(err.Code))
		} else {
			inHead.SetErrCode(-1)
		}
	}

	if msg != nil {
		body, err = ctx.Marshal.Marshal(msg)
		if err != nil {
			log.Error(err)
			return
		}
	}

	pk := network.NewRoutePacket(network.RoutePackType_SubFlag_Response, inHead, body)
	err = ctx.Conn.Send(pk)

	if err != nil {
		log.Error(err)
	}
}

func (ctx *context) SendAsync(msg proto.Message) error {
	msgid := calltable.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	return ctx.Core.SendAsyncToClient(ctx.Conn, ctx.caller, msgid, msg)
}

func (ctx *context) Caller() uint32 {
	return ctx.caller
}
