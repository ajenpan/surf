package core

import (
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/network"
)

type Context interface {
	Response(msg proto.Message, err error)
	Caller() auth.User
}

type context struct {
	Conn   network.Conn
	Core   *Surf
	Pk     *network.HVPacket
	Client *auth.UserInfo
}

func (ctx *context) Response(msg proto.Message, err error) {
	if err != nil {
		// if err, ok := err.(*errors.Error); ok {
		// 	ctx.RoutePk.SetErrCode(int16(err.Code))
		// } else {
		// 	ctx.RoutePk.SetErrCode(-1)
		// }
	}

	if msg != nil {

	}

}

func (ctx *context) SendAsync(msg proto.Message) error {
	return nil
}

func (ctx *context) Caller() auth.User {
	return ctx.Client
}
