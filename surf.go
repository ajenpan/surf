package surf

import (
	"reflect"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/log"
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

	UnhandleFunc func(s *server.TcpClient, m *server.MsgWraper)
}

func New(opt *Options) *Surf {
	if opt == nil {
		opt = &Options{}
	}
	s := &Surf{
		Options:       opt,
		asyncHandle:   make(map[string]FuncAsyncHandle),
		requestHandle: make(map[string]FuncRequestHandle),
		clients:       make(map[string]*server.TcpClient),
	}

	for _, addr := range opt.RouteAddrs {
		opts := &server.TcpClientOptions{
			RemoteAddress: addr,
			AuthToken:     s.JWToken,
			OnMessage:     s.onMessage,
			OnStatus:      s.onStatus,
		}
		client := server.NewTcpClient(opts)
		s.clients[addr] = client
	}
	return s
}

type Surf struct {
	*Options

	asyncHLock  sync.RWMutex
	asyncHandle map[string]FuncAsyncHandle

	requestHLock  sync.RWMutex
	requestHandle map[string]FuncRequestHandle

	clients map[string]*server.TcpClient
	CT      *calltable.CallTable[string]
}

func (s *Surf) Start() error {
	for _, c := range s.clients {
		c.Connect()
	}
	return nil
}

func (s *Surf) Stop() {
	for _, c := range s.clients {
		c.Close()
	}
}

func (s *Surf) getAsyncCallbcak(name string) FuncAsyncHandle {
	s.asyncHLock.RLock()
	defer s.asyncHLock.RUnlock()
	return s.asyncHandle[name]
}

func (s *Surf) getRequestCallback(name string) FuncRequestHandle {
	s.requestHLock.RLock()
	defer s.requestHLock.RUnlock()
	return s.requestHandle[name]
}

func (h *Surf) onMessage(s *server.TcpClient, m *server.MsgWraper) {
	if m.GetMsgtype() == server.MsgTypeAsync {
		wrap := &server.AsyncMsg{}
		err := proto.Unmarshal(m.GetBody(), wrap)
		if err != nil {
			log.Error(err)
			return
		}
		cb := h.getAsyncCallbcak(wrap.GetName())
		if cb != nil {
			cb(&Context{Session: s, UId: m.GetUid()}, wrap)
			return
		}
	} else if m.GetMsgtype() == server.MsgTypeRequest {
		wrap := &server.RequestMsg{}
		err := proto.Unmarshal(m.GetBody(), wrap)
		if err != nil {
			log.Error(err)
			return
		}
		cb := h.getRequestCallback(wrap.GetName())
		if cb != nil {
			cb(&Context{Session: s, UId: m.GetUid()}, wrap)
			return
		}
	}

	if h.UnhandleFunc != nil {
		h.UnhandleFunc(s, m)
	}
}

func (h *Surf) onStatus(s *server.TcpClient, enable bool) {
	log.Infof("onstatus: %v, %v", s.SessionID(), enable)
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
	s.asyncHLock.Lock()
	defer s.asyncHLock.Unlock()
	s.asyncHandle[name] = func(ctx *Context, msg *server.AsyncMsg) {
		var impMsgType T
		impMsg := reflect.New(reflect.TypeOf(impMsgType).Elem()).Interface().(T)
		err := proto.Unmarshal(msg.GetBody(), impMsg)
		if err != nil {
			log.Error(err)
			return
		}
		cb(ctx, impMsg)
	}
}

func RegisterRequestHandle[TReq, TResp proto.Message](s *Surf, name string, cb func(ctx *Context, msg TReq) (TResp, error)) {
	if cb == nil {
		return
	}

	s.requestHLock.Lock()
	defer s.requestHLock.Unlock()

	s.requestHandle[name] = func(ctx *Context, msg *server.RequestMsg) {
		var reqTypeHold TReq
		req := reflect.New(reflect.TypeOf(reqTypeHold).Elem()).Interface().(TReq)
		proto.Unmarshal(msg.GetBody(), req)
		resp, err := cb(ctx, req)
		ctx.Session.SendResponse(ctx.UId, msg, resp, err)
	}
}
