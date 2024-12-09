package core

import (
	"sync/atomic"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
)

type Context interface {
	SendAsync(msg proto.Message) error
	Response(msg proto.Message, err error)
	FromUserID() uint32
	FromUserRole() uint16
	ConnID() string
	Packet() *RoutePacket
}

type ConnContext struct {
	Conn      network.Conn
	Core      *Surf
	ReqPacket *RoutePacket
	responsed atomic.Bool
}

func (ctx *ConnContext) Response(msg proto.Message, herr error) {
	resped := ctx.responsed.Swap(true)
	if resped {
		log.Error("repeated response")
		return
	}

	var body []byte
	var err error

	rpk := NewRoutePacket(nil)

	respmsgid := GetMsgId(msg)
	rpk.SetMsgId(respmsgid)
	rpk.SetSYN(ctx.ReqPacket.GetSYN())
	rpk.SetToUID(ctx.FromUserID())
	rpk.SetToURole(ctx.FromUserRole())
	rpk.SetFromUID(ctx.Core.NodeID())
	rpk.SetFromURole(ctx.Core.getServerType())
	rpk.SetMsgType(RoutePackMsgType_Response)
	rpk.SetMarshalType(ctx.ReqPacket.GetMarshalType())

	if herr != nil {
		if err, ok := herr.(*errors.Error); ok {
			rpk.SetErrCode(int16(err.Code))
		} else {
			rpk.SetErrCode(-1)
		}
	}

	marshal := marshal.NewMarshaler(ctx.ReqPacket.GetMarshalType())

	if msg != nil && marshal != nil {
		body, err = marshal.Marshal(msg)
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
	msgid := GetMsgId(msg)
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
func (ctx *ConnContext) Packet() *RoutePacket {
	return ctx.ReqPacket
}
