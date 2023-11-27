package surf

import (
	"reflect"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/server"
	"github.com/ajenpan/surf/server/tcp"

	"github.com/ajenpan/surf/utils/rsagen"
)

func New(opt *Options) *Surf {
	if opt == nil {
		opt = &Options{}
	}

	ret := &Surf{
		Options: opt,
		msgCB:   make(map[string]func(server.Session, *server.Message)),
	}
	return ret
}

type Options struct {
}

type Surf struct {
	*Options
	tcpsvr *tcp.Server
	msgCB  map[string]func(server.Session, *server.Message)
}

func (s *Surf) Start() error {
	pk, err := rsagen.LoadRsaPublicKeyFromFile("public.pem")
	if err != nil {
		return err
	}
	opts := &server.TcpServerOptions{
		ListenAddr:       ":12002",
		AuthPublicKey:    pk,
		OnSessionMessage: s.onSessionMessage,
		OnSessionStatus:  s.onSessionStatus,
	}
	svr, err := server.NewTcpServer(opts)
	if err != nil {
		return err
	}
	return svr.Start()
}

func (h *Surf) onSessionMessage(s server.Session, m *server.Message) {

}

func (h *Surf) onSessionStatus(s server.Session, enable bool) {

}

func (s *Surf) Stop() {

}

func RegisterMsgHandle[T proto.Message](s *Surf, name string, cb func(s server.Session, msg T)) {
	s.msgCB[name] = func(sess server.Session, msg *server.Message) {
		var impMsgType T
		impMsg := reflect.New(reflect.TypeOf(impMsgType).Elem()).Interface().(T)
		proto.Unmarshal(msg.GetBody(), impMsg)
		cb(sess, impMsg)
	}
}

func RegisterRequestHandle[TReq, TResp proto.Message](s *Surf, name string, cb func(s server.Session, msg TReq) (TResp, error)) {
	s.msgCB[name] = func(sess server.Session, msg *server.Message) {
		var reqTypeHold TReq
		req := reflect.New(reflect.TypeOf(reqTypeHold).Elem()).Interface().(TReq)

		// var respTypeHold TResp
		// resp := reflect.New(reflect.TypeOf(respTypeHold).Elem()).Interface().(TResp)

		proto.Unmarshal(msg.GetBody(), req)
		resp, err := cb(sess, req)
		if err != nil {
			// TODO:
		}
		raw, _ := proto.Marshal(resp)
		msg.SetBody(raw)
		sess.Send(msg)
	}
}
