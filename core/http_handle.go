package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/network"
	"google.golang.org/protobuf/proto"
)

type HttpCallContext struct {
	w    http.ResponseWriter
	r    *http.Request
	core *Surf
}

// Request(msg proto.Message, cb func(pk *network.HVPacket, err error))
// Response(msg proto.Message, err error)
// Async(msg proto.Message) error
// Caller() auth.User

func (ctx *HttpCallContext) Response(msg proto.Message, err error) {
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

func (ctx *HttpCallContext) Request(msg proto.Message, cb func(pk *network.HVPacket, err error)) {

}

func (ctx *HttpCallContext) Async(msg interface{}) error {
	return fmt.Errorf("SendAsync is not impl")
}

func (ctx *HttpCallContext) Caller() auth.User {
	return nil
}
