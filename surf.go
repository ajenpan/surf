package surf

import (
	"reflect"

	"github.com/ajenpan/surf/server"
	"github.com/ajenpan/surf/server/tcp"
	"google.golang.org/protobuf/proto"
)

func New(opt *Options) *Surf {
	if opt == nil {
		opt = &Options{}
	}

	ret := &Surf{
		Options: opt,
		MsgCB:   make(map[string]func(server.Session, *server.Message)),
	}
	return ret
}

type Options struct {
	TcpListenAddr string
}

type Surf struct {
	*Options

	tcpsvr *tcp.Server

	MsgCB map[string]func(server.Session, *server.Message)
	// reqCB map[string]func(server.Session, *server.Message)
}

func (s *Surf) Start() {

}

func RegisterMsgHandle[T proto.Message](s *Surf, name string, cb func(s server.Session, msg T)) {
	s.MsgCB[name] = func(sess server.Session, msg *server.Message) {
		var impMsgType T
		impMsg := reflect.New(reflect.TypeOf(impMsgType).Elem()).Interface().(T)
		proto.Unmarshal(msg.Body, impMsg)
		cb(sess, impMsg)
	}
}

func RegisterRequestHandle[TReq, TResp proto.Message](s *Surf, name string, cb func(s server.Session, msg TReq) (TResp, error)) {
	s.MsgCB[name] = func(sess server.Session, msg *server.Message) {
		var reqTypeHold TReq
		req := reflect.New(reflect.TypeOf(reqTypeHold).Elem()).Interface().(TReq)

		// var respTypeHold TResp
		// resp := reflect.New(reflect.TypeOf(respTypeHold).Elem()).Interface().(TResp)

		proto.Unmarshal(msg.Body, req)
		resp, err := cb(sess, req)
		if err != nil {
			// TODO:
		}
		msg.Body, _ = proto.Marshal(resp)
		sess.Send(msg)
	}
}
