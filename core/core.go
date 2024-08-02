package core

import (
	"crypto/rsa"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ajenpan/surf/core/log"
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/registry"
	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/core/utils/marshal"
)

type Options struct {
	ServerId     uint32
	ConnectRoute bool
	RouteToken   string

	HttpListenAddr string
	WsListenAddr   string
	TcpListenAddr  string

	CTByName *calltable.CallTable[string]
	CTById   *calltable.CallTable[int32]
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
	pk  *rsa.PublicKey
	Reg *registry.Registry

	tcpsvr  *network.TcpServer
	wssvr   *network.WSServer
	httpsvr *http.Server
}

func (s *Surf) init() error {
	var err error
	tcpsvr, err := network.NewTcpServer(network.TcpServerOptions{
		ListenAddr:       ":9999",
		HeatbeatInterval: 30 * time.Second,
		OnConnPacket:     s.onConnPacket,
		OnConnEnable:     s.onConnStatus,
		OnConnAuth:       s.onConnAuth,
	})
	if err != nil {
		return err
	}
	s.tcpsvr = tcpsvr
	return nil
}

func (s *Surf) Close() error {

	return nil
}

func (s *Surf) Start() error {
	if len(s.HttpListenAddr) > 1 {
		s.startHttpSvr()
	}

	if (len(s.WsListenAddr)) > 1 {
		s.startWsSvr()
	}

	if len(s.TcpListenAddr) > 1 {
		s.startTcpSvr()
	}
	// quit := make(chan struct{})
	// s.httpsvr = &network.HttpSvr{
	// 	Addr:    s.HttpListenAddr,
	// 	Marshal: &marshal.JSONPb{},
	// 	Mux:     http.NewServeMux(),
	// }

	// s.httpsvr.ServerCallTable(&s.CTByName)

	// go func() {
	// 	err := s.httpsvr.Run()
	// 	select {
	// 	case <-quit:
	// 	case errchan <- err:
	// 	}
	// }()

	return nil
}

func (s *Surf) startHttpSvr() {
	log.Infof("startHttpSvr")

	mux := http.NewServeMux()
	s.CTByName.Range(func(key string, method *calltable.Method) bool {
		if !strings.HasPrefix(key, "/") {
			key = "/" + key
		}
		cb := s.WrapMethod(method)
		mux.HandleFunc(key, cb)
		return true
	})

	svr := &http.Server{
		Addr:    s.HttpListenAddr,
		Handler: mux,
	}
	s.httpsvr = svr
	go svr.ListenAndServe()
}

func (s *Surf) startWsSvr() {
	log.Infof("startWsSvr")

	ws := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   s.WsListenAddr,
		OnConnPacket: s.onConnPacket,
		OnConnEnable: s.onConnStatus,
		OnConnAuth:   s.onConnAuth,
	})

	ws.Start()
}

func (s *Surf) startTcpSvr() {
	log.Infof("startTcpSvr")

	tcpsvr, err := network.NewTcpServer(network.TcpServerOptions{
		ListenAddr:       ":9999",
		HeatbeatInterval: 30 * time.Second,
		OnConnPacket:     s.onConnPacket,
		OnConnEnable:     s.onConnStatus,
		OnConnAuth:       s.onConnAuth,
	})
	if err != nil {
		panic(err)
	}
	s.tcpsvr = tcpsvr
	tcpsvr.Start()
}

func (h *Surf) onConnPacket(s network.Conn, pk *network.HVPacket) {
	// ctx := &Context{Session: s, UId: m.GetUid()}
	// h.OnServerMsgWraper(ctx, m)
	// sf := pk.GetSubFlag()
	// switch sf {
	// case 1:
	// 	pk.GetBody()
	// default:
	// }
}

func (h *Surf) onConnAuth(data []byte) (auth.User, error) {
	return auth.VerifyToken(h.pk, data)
}

func (h *Surf) onConnStatus(s network.Conn, enable bool) {
	// log.Infof("route onstatus: %v, %v", s.SessionID(), enable)
}

func OnRouteAsync(ct *calltable.CallTable[uint32], conn network.Conn, rpk network.RoutePacketRaw) {
	// var err error
	// rpk := network.RoutePacketRaw(pk.GetBody())

	method := ct.Get(rpk.GetMsgId())
	if method == nil {
		rpk.SetMsgType(network.RouteMsgType_HandleErr)
		rpk.SetErrCode(network.RouteHandleErrCode_MethodNotFound)
		pk := network.NewHVPacket()
		pk.SetFlag(1)
		pk.SetBody(rpk.GetHead())
		conn.Send(pk)
		return
	}

	msg := method.NewRequest()
	mar := &proto.UnmarshalOptions{}
	err := mar.Unmarshal(rpk.GetBody(), msg.(proto.Message))
	if err != nil {
		rpk.SetMsgType(network.RouteMsgType_HandleErr)
		rpk.SetErrCode(network.RouteHandleErrCode_MethodParseErr)
		pk := network.NewHVPacket()
		pk.SetFlag(1)
		pk.SetBody(rpk.GetHead())
		conn.Send(pk)
		return
	}

	// var ctx Context = &context{
	// 	Conn: conn,
	// 	Core: nil,
	// 	Raw:  pk,
	// }

	// method.Call(ctx, msg)
}

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

func (s *Surf) WrapMethod(method *calltable.Method) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		req := method.NewRequest()

		if err = (&marshal.JSONPb{}).Unmarshal(raw, req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		contenttype := r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", contenttype)

		var ctx Context = &HttpCallContext{
			w:    w,
			r:    r,
			core: s,
		}

		// here call method
		method.Call(ctx, req)
	}
}
