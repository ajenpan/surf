package surf

import (
	"reflect"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/log"
	"github.com/ajenpan/surf/msg"
	"github.com/ajenpan/surf/server"
	"github.com/ajenpan/surf/utils/calltable"
)

type Context struct {
	Session server.Session
	UId     uint32
}

type FuncAsyncHandle func(*Context, *server.AsyncMsg)
type FuncRequestHandle func(*Context, *server.RequestMsg)

type Options struct {
	JWToken    string
	RouteAddrs []string

	OnSessionConn    func(s server.Session)
	OnSessionDisconn func(s server.Session, e error)
	OnSessionMsg     func(s server.Session, m *server.MsgWraper)
}

func New(opt *Options) *Surf {
	if opt == nil {
		opt = &Options{}
	}
	s := &Surf{
		Options:     opt,
		routeClient: make(map[string]*server.TcpClient),
	}

	for _, addr := range opt.RouteAddrs {
		opts := &server.TcpClientOptions{
			RemoteAddress:     addr,
			AuthToken:         s.JWToken,
			OnMessage:         s.onMessage,
			OnStatus:          s.onStatus,
			ReconnectDelaySec: 10,
		}
		client := server.NewTcpClient(opts)
		s.routeClient[addr] = client
	}
	return s
}

type HandlerRegister struct {
	asyncHLock  sync.RWMutex
	asyncHandle map[string]FuncAsyncHandle

	requestHLock  sync.RWMutex
	requestHandle map[string]FuncRequestHandle
}

func (hr *HandlerRegister) getAsyncCallbcak(name string) FuncAsyncHandle {
	hr.asyncHLock.RLock()
	defer hr.asyncHLock.RUnlock()
	return hr.asyncHandle[name]
}
func (hr *HandlerRegister) getRequestCallback(name string) FuncRequestHandle {
	hr.requestHLock.RLock()
	defer hr.requestHLock.RUnlock()
	return hr.requestHandle[name]
}

func (hr *HandlerRegister) RegisterAysncHandle(name string, cb FuncAsyncHandle) {
	hr.asyncHLock.Lock()
	defer hr.asyncHLock.Unlock()
	hr.asyncHandle[name] = cb
}

func (hr *HandlerRegister) RegisterRequestHandle(name string, cb FuncRequestHandle) {
	hr.requestHLock.Lock()
	defer hr.requestHLock.Unlock()
	hr.requestHandle[name] = cb
}

func (hr *HandlerRegister) OnServerMsgWraper(ctx *Context, m *server.MsgWraper) bool {
	if m.GetMsgtype() == server.MsgTypeAsync {
		wrap := &server.AsyncMsg{}
		proto.Unmarshal(m.GetBody(), wrap)
		cb := hr.getAsyncCallbcak(wrap.GetName())
		if cb != nil {
			cb(ctx, wrap)
			return true
		}
	} else if m.GetMsgtype() == server.MsgTypeRequest {
		wrap := &server.RequestMsg{}
		proto.Unmarshal(m.GetBody(), wrap)
		cb := hr.getRequestCallback(wrap.GetName())
		if cb != nil {
			cb(ctx, wrap)
			return true
		}
	}
	return false
}

type Surf struct {
	*Options

	HandlerRegister

	routeClient map[string]*server.TcpClient
	CT          *calltable.CallTable[string]
}

func (s *Surf) Start() error {
	RegisterAysncHandle(s, "NotifyLinkStatus", s.onNotifyLinkStatus)

	for _, c := range s.routeClient {
		c.Connect()
	}
	return nil
}

func (s *Surf) Stop() {
	for _, c := range s.routeClient {
		c.Close()
	}
}

func (h *Surf) onMessage(s *server.TcpClient, m *server.MsgWraper) {
	ctx := &Context{Session: s, UId: m.GetUid()}
	h.OnServerMsgWraper(ctx, m)
}

func (h *Surf) onStatus(s *server.TcpClient, enable bool) {
	log.Infof("route onstatus: %v, %v", s.SessionID(), enable)
}

func (h *Surf) onReqLinkSession(ctx *Context, m *msg.ReqLinkSession) (resp *msg.RespLinkSession, err error) {
	resp = &msg.RespLinkSession{}
	return
}

func (h *Surf) onNotifyLinkStatus(ctx *Context, m *msg.NotifyLinkStatus) {

}

// func (h *Surf) OnAsync(s server.Session, uid uint32, m *server.AsyncMsg) {
// 	var err error
// 	method := h.CT.Get(m.Name)
// 	if method == nil {
// 		log.Warnf("method not found: %s", m.Name)
// 		return
// 	}
// 	pbmarshal := &marshal.ProtoMarshaler{}
// 	req := method.NewRequest()
// 	err = pbmarshal.Unmarshal(m.Body, req)
// 	if err != nil {
// 		log.Errorf("unmarshal error: %v", err)
// 		return
// 	}
// 	result := method.Call(h, &Context{Session: s, UId: uid}, req)
// 	if len(result) == 1 {
// 		err = result[0].Interface().(error)
// 	}
// 	if err != nil {
// 		log.Warnf("method call error: %v", err)
// 	}
// }
// func (h *Surf) OnRequest(s server.Session, uid uint32, m *server.RequestMsg) {
// 	var err error
// 	method := h.CT.Get(m.Name)
// 	if method == nil {
// 		log.Warnf("method not found: %s", m.Name)
// 		return
// 	}
// 	pbmarshal := &marshal.ProtoMarshaler{}
// 	req := method.NewRequest()
// 	err = pbmarshal.Unmarshal(m.Body, req)
// 	if err != nil {
// 		log.Errorf("unmarshal error: %v", err)
// 		return
// 	}
// 	result := method.Call(h, &Context{Session: s, UId: uid}, req)
// 	var resp proto.Message
// 	if len(result) != 2 {
// 		err = fmt.Errorf("method resp param error")
// 	} else {
// 		err, _ = result[0].Interface().(error)
// 		resp, _ = result[1].Interface().(proto.Message)
// 	}
// 	s.SendResponse(uid, m, resp, err)
// }

func RegisterAysncHandle[T proto.Message](s *Surf, name string, cb func(ctx *Context, msg T)) {
	if cb == nil {
		return
	}
	s.RegisterAysncHandle(name, func(ctx *Context, msg *server.AsyncMsg) {
		var impMsgType T
		impMsg := reflect.New(reflect.TypeOf(impMsgType).Elem()).Interface().(T)
		err := proto.Unmarshal(msg.GetBody(), impMsg)
		if err != nil {
			log.Error(err)
			return
		}
		cb(ctx, impMsg)
	})
}

func RegisterRequestHandle[TReq, TResp proto.Message](s *Surf, name string, cb func(ctx *Context, msg TReq) (TResp, error)) {
	if cb == nil {
		return
	}
	s.RegisterRequestHandle(name, func(ctx *Context, msg *server.RequestMsg) {
		var reqTypeHold TReq
		req := reflect.New(reflect.TypeOf(reqTypeHold).Elem()).Interface().(TReq)
		proto.Unmarshal(msg.GetBody(), req)
		resp, err := cb(ctx, req)
		ctx.Session.SendResponse(ctx.UId, msg, resp, err)
	})
}
