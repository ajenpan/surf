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

type httpCallContext struct {
	w     http.ResponseWriter
	r     *http.Request
	core  *Surf
	uinfo *auth.UserInfo
	// marshaler marshal.Marshaler
}

type httpResponeWrap struct {
	ErrCode int         `json:"errcode"`
	ErrMsg  string      `json:"errmsg"`
	Data    interface{} `json:"data"`
}

func (ctx *httpCallContext) Response(msg proto.Message, err error) {
	enc := json.NewEncoder(ctx.w)
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
		ctx.w.WriteHeader(http.StatusInternalServerError)
		ctx.w.Write([]byte(encerr.Error()))
	}
}

func (ctx *httpCallContext) Request(msg proto.Message, cb func(pk *network.HVPacket, err error)) {
	// do nothing
}

func (ctx *httpCallContext) Async(msg proto.Message) error {
	return fmt.Errorf("SendAsync is not impl")
}

func (ctx *httpCallContext) Caller() uint32 {
	return ctx.uinfo.UId
}
