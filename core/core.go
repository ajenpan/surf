package core

import (
	"crypto/rsa"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/registry"
)

// type FuncAsyncHandle func(*Context, *network.AsyncMsg)
// type FuncRequestHandle func(*Context, *network.RequestMsg)

type Options struct {
	JWToken    string
	pk         *rsa.PublicKey
	RouteAddrs []string
}

func New(opt Options) *Surf {
	s := &Surf{
		Options: opt,
		// routeClient: make(map[string]*network.TcpClient),
	}

	// for _, addr := range opt.RouteAddrs {
	// 	opts := &network.TcpClientOptions{
	// 		RemoteAddress:     addr,
	// 		AuthToken:         s.JWToken,
	// 		OnMessage:         s.onMessage,
	// 		OnStatus:          s.onStatus,
	// 		ReconnectDelaySec: 10,
	// 	}
	// 	client := network.NewTcpClient(opts)
	// 	s.routeClient[addr] = client
	// }
	return s
}

// type HandlerRegister struct {
// 	asyncHLock  sync.RWMutex
// 	asyncHandle map[string]FuncAsyncHandle

// 	requestHLock  sync.RWMutex
// 	requestHandle map[string]FuncRequestHandle
// }

// func (hr *HandlerRegister) getAsyncCallbcak(name string) FuncAsyncHandle {
// 	hr.asyncHLock.RLock()
// 	defer hr.asyncHLock.RUnlock()
// 	return hr.asyncHandle[name]
// }
// func (hr *HandlerRegister) getRequestCallback(name string) FuncRequestHandle {
// 	hr.requestHLock.RLock()
// 	defer hr.requestHLock.RUnlock()
// 	return hr.requestHandle[name]
// }

// func (hr *HandlerRegister) RegisterAysncHandle(name string, cb FuncAsyncHandle) {
// 	hr.asyncHLock.Lock()
// 	defer hr.asyncHLock.Unlock()
// 	hr.asyncHandle[name] = cb
// }

// func (hr *HandlerRegister) RegisterRequestHandle(name string, cb FuncRequestHandle) {
// 	hr.requestHLock.Lock()
// 	defer hr.requestHLock.Unlock()
// 	hr.requestHandle[name] = cb
// }

// func (hr *HandlerRegister) OnServerMsgWraper(ctx *Context, m *network.MsgWraper) bool {
// if m.GetMsgtype() == network.MsgTypeAsync {
// 	wrap := &network.AsyncMsg{}
// 	proto.Unmarshal(m.GetBody(), wrap)
// 	cb := hr.getAsyncCallbcak(wrap.GetName())
// 	if cb != nil {
// 		cb(ctx, wrap)
// 		return true
// 	}
// } else if m.GetMsgtype() == network.MsgTypeRequest {
// 	wrap := &network.RequestMsg{}
// 	proto.Unmarshal(m.GetBody(), wrap)
// 	cb := hr.getRequestCallback(wrap.GetName())
// 	if cb != nil {
// 		cb(ctx, wrap)
// 		return true
// 	}
// }
// return false
// }

type Surf struct {
	Options
	Reg *registry.Registry

	tcpsvr *network.TcpServer
}

// func (s *Surf) Start() error {
// 	RegisterAysncHandle(s, "NotifyLinkStatus", s.onNotifyLinkStatus)
// 	for _, c := range s.routeClient {
// 		c.Connect()
// 	}
// 	return nil
// }
//	func (s *Surf) Stop() {
//		for _, c := range s.routeClient {
//			c.Close()
//		}
//	}

func (s *Surf) init() error {
	var err error
	tcpsvr, err := network.NewTcpServer(network.TcpServerOptions{
		ListenAddr:       ":9999",
		HeatbeatInterval: 30 * time.Second,
		OnPacket:         s.onPacket,
		OnStatus:         s.onStatus,
		OnAuth: func(data []byte) (auth.User, error) {
			return auth.VerifyToken(s.pk, data)
		},
	})
	if err != nil {
		return err
	}
	s.tcpsvr = tcpsvr
	return nil
}

func (s *Surf) start() error {

	s.tcpsvr.Start()

	return nil
}

func (s *Surf) stop() {

}

func (s *Surf) Run() error {

	return nil
}

func (h *Surf) onPacket(s *network.Conn, pk *network.HVPacket) {
	// ctx := &Context{Session: s, UId: m.GetUid()}
	// h.OnServerMsgWraper(ctx, m)
	sf := pk.GetSubFlag()
	switch sf {
	case 1:
		pk.GetBody()

	default:

	}

}

func (h *Surf) onStatus(s *network.Conn, enable bool) {
	// log.Infof("route onstatus: %v, %v", s.SessionID(), enable)
}

// func (h *Surf) OnAsync(s network.Session, uid uint32, m *network.AsyncMsg) {
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
// func (h *Surf) OnRequest(s network.Session, uid uint32, m *network.RequestMsg) {
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
	// if cb == nil {
	// 	return
	// }
	// s.RegisterAysncHandle(name, func(ctx *Context, msg *network.AsyncMsg) {
	// 	var impMsgType T
	// 	impMsg := reflect.New(reflect.TypeOf(impMsgType).Elem()).Interface().(T)
	// 	err := proto.Unmarshal(msg.GetBody(), impMsg)
	// 	if err != nil {
	// 		log.Error(err)
	// 		return
	// 	}
	// 	cb(ctx, impMsg)
	// })
}

func RegisterRequestHandle[TReq, TResp proto.Message](s *Surf, name string, cb func(ctx *Context, msg TReq) (TResp, error)) {
	// if cb == nil {
	// 	return
	// }
	// s.RegisterRequestHandle(name, func(ctx *Context, msg *network.RequestMsg) {
	// 	var reqTypeHold TReq
	// 	req := reflect.New(reflect.TypeOf(reqTypeHold).Elem()).Interface().(TReq)
	// 	proto.Unmarshal(msg.GetBody(), req)
	// 	resp, err := cb(ctx, req)
	// 	ctx.Session.SendResponse(ctx.UId, msg, resp, err)
	// })
}
