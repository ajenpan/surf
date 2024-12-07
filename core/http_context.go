package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/network"
	"google.golang.org/protobuf/proto"
)

type HttpContext struct {
	W      http.ResponseWriter
	R      *http.Request
	UInfo  auth.User
	ConnId string
}

type httpResponeWrap struct {
	ErrCode int         `json:"errcode"`
	ErrMsg  string      `json:"errmsg"`
	Data    interface{} `json:"data"`
}

func (ctx *HttpContext) Response(msg proto.Message, err error) {
	enc := json.NewEncoder(ctx.W)
	wrap := &httpResponeWrap{Data: msg}

	if err != nil {
		wrap.ErrMsg = err.Error()
		if errs, ok := err.(*errors.Error); ok {
			wrap.ErrCode = int(errs.Code)
		} else {
			wrap.ErrCode = -1
		}
	}

	encerr := enc.Encode(wrap)

	if encerr != nil {
		ctx.W.WriteHeader(http.StatusInternalServerError)
		ctx.W.Write([]byte(encerr.Error()))
	}
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
	return nil
}
