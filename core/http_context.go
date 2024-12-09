package core

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/ajenpan/surf/core/auth"
	surferr "github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/marshal"
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

	responsed atomic.Bool
}

func (ctx *HttpContext) Response(msg proto.Message, herr error) {
	resped := ctx.responsed.Swap(true)
	if resped {
		log.Error("repeated response")
		return
	}

	ctx.respC <- func() {
		w := ctx.W
		marshaler := marshal.NewMarshaler(ctx.ReqPacket.GetMarshalType())
		if msg != nil && marshaler != nil {
			body, err := marshaler.Marshal(msg)
			if err == nil {
				w.Write(body)
			} else {
				log.Error("response marshal error", "err", err)
				herr = errors.Join(herr, err)
			}

			w.Header().Set("MsgId", fmt.Sprintf("%d", GetMsgId(msg)))
		}

		if herr != nil {
			var errcode int = -1
			if err, ok := herr.(*surferr.Error); ok {
				errcode = int(err.Code)
			}
			w.Header().Set("errcode", strconv.Itoa(errcode))
			w.Header().Set("errmsg", herr.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
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
	return ctx.ReqPacket
}
