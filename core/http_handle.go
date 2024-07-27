package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ajenpan/surf/core/auth"
)

type HttpCallContext struct {
	w    http.ResponseWriter
	r    *http.Request
	core *Surf
}

func (ctx *HttpCallContext) Response(msg interface{}, err error) {
	type httpWrap struct {
		Error error       `json:"err"`
		Data  interface{} `json:"data"`
	}

	enc := json.NewEncoder(ctx.w)
	encerr := enc.Encode(&httpWrap{Data: msg, Error: err})

	if encerr != nil {
		ctx.w.WriteHeader(http.StatusInternalServerError)
		ctx.w.Write([]byte(encerr.Error()))
	} else {
		ctx.w.WriteHeader(http.StatusOK)
	}
}

func (ctx *HttpCallContext) SendAsync(msg interface{}) error {
	return fmt.Errorf("SendAsync is not impl")
}

func (ctx *HttpCallContext) Caller() auth.User {
	return nil
}
