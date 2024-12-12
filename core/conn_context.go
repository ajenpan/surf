package core

import (
	"context"
	"sync/atomic"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
)

type Context interface {
	Async(msg proto.Message) error
	Response(msg proto.Message, err error)

	FromUId() uint32
	FromURole() uint16
	ConnId() string
	Packet() *RoutePacket
}

type ConnContext struct {
	ReqConn   network.Conn
	Core      *Surf
	ReqPacket *RoutePacket
	resped    atomic.Bool
}

func (ctx *ConnContext) Response(msg proto.Message, herr error) {
	context.Background()
	resped := ctx.resped.Swap(true)
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
	rpk.SetToUId(ctx.FromUId())
	rpk.SetToURole(ctx.FromURole())
	rpk.SetFromUId(ctx.Core.NodeID())
	rpk.SetFromURole(ctx.Core.NodeType())
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
		rpk.SetBody(body)
	}

	err = ctx.ReqConn.Send(rpk.ToHVPacket())
	if err != nil {
		log.Error("response send error", "err", err)
	}
}

func (ctx *ConnContext) Async(msg proto.Message) error {
	return ctx.Core.SendAsync(ctx.ReqConn, ctx.FromURole(), ctx.FromUId(), msg)
}

func (ctx *ConnContext) FromUId() uint32 {
	return ctx.ReqPacket.GetFromUId()
}

func (ctx *ConnContext) FromURole() uint16 {
	return ctx.ReqPacket.GetFromURole()
}

func (ctx *ConnContext) Conn() network.Conn {
	return ctx.ReqConn
}
func (ctx *ConnContext) ConnId() string {
	return ctx.ReqConn.ConnId()
}
func (ctx *ConnContext) Packet() *RoutePacket {
	return ctx.ReqPacket
}
