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

type HttpCallContext struct {
	w     http.ResponseWriter
	r     *http.Request
	core  *Surf
	uinfo *auth.UserInfo
}

func (ctx *HttpCallContext) Response(msg proto.Message, err error) {
	type httpWrap struct {
		ErrCode int         `json:"errcode"`
		ErrMsg  string      `json:"errmsg"`
		Data    interface{} `json:"data"`
	}

	enc := json.NewEncoder(ctx.w)
	wrap := &httpWrap{Data: msg}

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
		ctx.w.WriteHeader(http.StatusInternalServerError)
		ctx.w.Write([]byte(encerr.Error()))
	}
}

func (ctx *HttpCallContext) Request(msg proto.Message, cb func(pk *network.HVPacket, err error)) {
	// do nothing
}

func (ctx *HttpCallContext) Async(msg interface{}) error {
	return fmt.Errorf("SendAsync is not impl")
}

func (ctx *HttpCallContext) Caller() uint32 {
	return ctx.uinfo.UId
}
