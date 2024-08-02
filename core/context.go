package core

import (
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/network"
)

type Context interface {
	Response(msg interface{}, err error)
	SendAsync(msg interface{}) error
	Caller() auth.User
}

type context struct {
	Conn    network.Conn
	Core    *Surf
	Raw     *network.HVPacket
	RoutePk network.RoutePacketRaw

	Client *auth.UserInfo
}

func (ctx *context) Response(msg proto.Message, err error) {
	if err != nil {
		if err, ok := err.(*errors.Error); ok {
			ctx.RoutePk.SetErrCode(int16(err.Code))
		} else {
			ctx.RoutePk.SetErrCode(-1)
		}
	}

	if msg != nil {
		mar := &proto.MarshalOptions{}
		mar.MarshalAppend(ctx.RoutePk.GetHead(), msg)
	}

}

func (ctx *context) SendAsync(msg proto.Message) error {
	return nil
}

func (ctx *context) Caller() auth.User {
	return ctx.Client
}
