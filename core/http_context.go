package core

import (
	"fmt"
	"net/http"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/network"
	"google.golang.org/protobuf/proto"
)

type HttpContext struct {
	W         http.ResponseWriter
	R         *http.Request
	UInfo     auth.User
	Core      *Surf
	ConnId    string
	ReqPacket *RoutePacket
	respC     chan func()
}

type httpResponeWrap struct {
	ErrCode int         `json:"errcode"`
	ErrMsg  string      `json:"errmsg"`
	Data    interface{} `json:"data"`
}

func (ctx *HttpContext) Response(msg proto.Message, herr error) {

	// var body []byte
	// var err error

	// rpk := NewRoutePacket(nil)

	// respmsgid := GetMsgId(msg)
	// rpk.SetMsgId(respmsgid)
	// rpk.SetSYN(ctx.ReqPacket.GetSYN())
	// rpk.SetToUID(ctx.FromUserID())
	// rpk.SetToURole(ctx.FromUserRole())
	// rpk.SetFromUID(ctx.Core.NodeID())
	// rpk.SetFromURole(ctx.Core.getServerType())
	// rpk.SetMsgType(RoutePackMsgType_Response)
	// rpk.SetMarshalType(ctx.ReqPacket.GetMarshalType())

	// if herr != nil {
	// 	if err, ok := herr.(*errors.Error); ok {
	// 		rpk.SetErrCode(int16(err.Code))
	// 	} else {
	// 		rpk.SetErrCode(-1)
	// 	}
	// }

	// marshal := marshal.NewMarshaler(ctx.ReqPacket.GetMarshalType())

	// if msg != nil && marshal == nil {
	// 	body, err = marshal.Marshal(msg)
	// 	if err != nil {
	// 		log.Error("response marshal error", "err", err)
	// 		return
	// 	}
	// 	rpk.Body = body
	// }

	// err = ctx.Conn.Send(rpk.ToHVPacket())
	// if err != nil {
	// 	log.Error("response send error", "err", err)
	// }

	// ctx.respC <- func() {
	// 	enc := json.NewEncoder(ctx.W)
	// 	wrap := &httpResponeWrap{Data: msg}

	// 	if err != nil {
	// 		wrap.ErrMsg = err.Error()
	// 		if errs, ok := err.(*errors.Error); ok {
	// 			wrap.ErrCode = int(errs.Code)
	// 		} else {
	// 			wrap.ErrCode = -1
	// 		}
	// 	}

	// 	encerr := enc.Encode(wrap)

	// 	if encerr != nil {
	// 		ctx.W.WriteHeader(http.StatusInternalServerError)
	// 		ctx.W.Write([]byte(encerr.Error()))
	// 	}
	// }
}

func (ctx *HttpContext) Request(msg proto.Message, cb func(pk *network.HVPacket, err error)) {
	// do nothing
}

func (ctx *HttpContext) SendAsync(msg proto.Message) error {
	return fmt.Errorf("SendAsync is not impl")
}

func (ctx *HttpContext) FromUserID() uint32 {
	return ctx.UInfo.UserID()
}

func (ctx *HttpContext) FromUserRole() uint16 {
	return ctx.UInfo.UserRole()
}

func (ctx *HttpContext) ConnID() string {
	return ctx.ConnId
}

func (ctx *HttpContext) Packet() *RoutePacket {
	return ctx.ReqPacket
}
