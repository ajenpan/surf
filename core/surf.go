package core

import (
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/registry"
	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/core/utils/marshal"
	utilSignal "github.com/ajenpan/surf/core/utils/signal"
)

type Options struct {
	ServerType uint16
	RouteToken string

	HttpListenAddr string
	WsListenAddr   string
	TcpListenAddr  string

	CTByName *calltable.CallTable[string]
	CTById   *calltable.CallTable[uint32]

	Marshaler marshal.Marshaler
	PublicKey *rsa.PublicKey
}

func NewSurf(opt *Options) *Surf {
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

type RequestCallbackFunc func(timeout bool, pk *network.HVPacket)

type RequestCallbackCache struct {
	cbfun   RequestCallbackFunc
	timeout *time.Timer
}

type Surf struct {
	*Options
	Reg *registry.Registry

	tcpsvr  *network.TcpServer
	wssvr   *network.WSServer
	httpsvr *http.Server
	nodeid  uint32

	respWatier sync.Map
	synIdx     uint32
}

func (s *Surf) init() error {
	return nil
}

func (s *Surf) Close() error {
	if s.tcpsvr != nil {
		s.tcpsvr.Stop()
	}
	if s.wssvr != nil {
		s.wssvr.Stop()
	}
	if s.httpsvr != nil {
		s.httpsvr.Close()
	}
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

func (s *Surf) Run() error {
	s.CTById.Range(func(key uint32, value *calltable.Method) bool {
		log.Infof("handle func,msgid:%d, funcname:%s", key, value.FuncName)
		return true
	})

	if err := s.Start(); err != nil {
		return err
	}

	sig := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", sig.String())
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

	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   s.WsListenAddr,
		OnConnPacket: s.onConnPacket,
		OnConnEnable: s.onConnStatus,
		OnConnAuth:   s.onConnAuth,
	})
	if err != nil {
		panic(err)
	}

	s.wssvr = ws
	s.wssvr.Start()
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
	s.tcpsvr.Start()
}

func (h *Surf) GetNodeId() uint32 {
	return h.nodeid
}

func (h *Surf) GetServerType() uint16 {
	return uint16(h.ServerType)
}

func (h *Surf) GetSYN() uint32 {
	ret := atomic.AddUint32(&h.synIdx, 1)
	if ret == 0 {
		return atomic.AddUint32(&h.synIdx, 1)
	}
	return ret
}

func (h *Surf) pushRespCallback(syn uint32, cb RequestCallbackFunc) error {
	timeout := time.AfterFunc(3*time.Second, func() {
		info := h.popRespCallback(syn)
		if info != nil && info.cbfun != nil {
			info.cbfun(true, nil)
		}

		if info.timeout != nil {
			info.timeout.Stop()
			info.timeout = nil
		}
	})

	cache := &RequestCallbackCache{
		cbfun:   cb,
		timeout: timeout,
	}

	h.respWatier.Store(syn, cache)
	return nil
}

func (h *Surf) popRespCallback(syn uint32) *RequestCallbackCache {
	cache, ok := h.respWatier.Load(syn)
	if !ok {
		return nil
	}
	return cache.(*RequestCallbackCache)
}

func (h *Surf) SendResponeToClient(uid uint32, syn uint32, errcode uint16, msgid uint32, msg any) error {
	conn := h.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("not found route")
	}
	body, err := h.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}
	head := network.RoutePacketHead(make([]byte, network.RoutePackHeadLen))
	head.SetClientId(uid)
	head.SetMsgId(msgid)
	head.SetNodeId(h.GetNodeId())
	head.SetSYN(syn)
	pk := network.NewRoutePacket(network.RoutePackType_SubFlag_Response, head, body)
	return conn.Send(pk)
}

func (h *Surf) SendRequestToClient(uid uint32, msgid uint32, msg any, cb RequestCallbackFunc) error {
	conn := h.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("not found route")
	}

	body, err := h.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}
	syn := h.GetSYN()
	head := network.RoutePacketHead(make([]byte, network.RoutePackHeadLen))
	head.SetClientId(uid)
	head.SetMsgId(msgid)
	head.SetNodeId(h.GetNodeId())
	head.SetSYN(syn)

	h.pushRespCallback(syn, cb)

	pk := network.NewRoutePacket(network.RoutePackType_SubFlag_Request, head, body)

	err = conn.Send(pk)
	if err != nil {
		h.popRespCallback(syn)
	}
	return err
}

func (h *Surf) SendAsyncToClient(uid uint32, msgid uint32, msg any) error {
	conn := h.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("not found route")
	}

	body, err := h.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}

	head := network.RoutePacketHead(make([]byte, network.RoutePackHeadLen))
	head.SetClientId(uid)
	head.SetMsgId(msgid)
	head.SetNodeId(h.GetNodeId())
	head.SetSYN(h.GetSYN())

	pk := network.NewRoutePacket(network.RoutePackType_SubFlag_Async, head, body)
	return conn.Send(pk)
}

func (h *Surf) SendToNode(nodeid uint32, pk *network.HVPacket) error {
	return nil
}

func (h *Surf) GetClientConn(id uint32) network.Conn {
	return nil
}

func (h *Surf) GetNodeConn(id uint32) network.Conn {
	return nil
}

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

		method.Call(ctx, req)
	}
}

func (h *Surf) onConnPacket(s network.Conn, pk *network.HVPacket) {
	switch pk.Meta.GetType() {
	case network.PacketType_Route:
		h.onRoutePacket(s, pk)
	case network.PacketType_NodeInner:
	default:
	}
}

func (h *Surf) onConnAuth(data []byte) (auth.User, error) {
	return auth.VerifyToken(h.PublicKey, data)
}

func (h *Surf) onConnStatus(s network.Conn, enable bool) {
	log.Infof("route onstatus: %v, %v", s.ConnID(), enable)
}

func (h *Surf) getClientInfo(uid uint32) *auth.UserInfo {
	return nil
}

func (h *Surf) catch() {
	if err := recover(); err != nil {
		log.Error(err)
	}
}

func (h *Surf) onRoutePacket(s network.Conn, pk *network.HVPacket) {
	defer h.catch()

	head := network.RoutePacketHead(pk.Head)

	clientinfo := h.getClientInfo(head.GetClientId())
	if clientinfo == nil {
		return
	}

	switch pk.Meta.GetSubFlag() {
	case network.RoutePackType_SubFlag_Async:
		fallthrough
	case network.RoutePackType_SubFlag_Request:
		method := h.CTById.Get(head.GetMsgId())
		if method == nil {
			// todo:
			return
		}

		// decode msg
		req := method.NewRequest()
		err := h.Marshaler.Unmarshal(pk.GetBody(), req)
		if err != nil {
			// todo:
			return
		}

		ctx := &context{
			Conn:   s,
			Core:   h,
			Pk:     pk,
			caller: head.GetClientId(),
		}

		method.Call(ctx, req)
	case network.RoutePackType_SubFlag_Response:
		cbinfo := h.popRespCallback(head.GetSYN())
		if cbinfo == nil {
			return
		}

	case network.RoutePackType_SubFlag_RouteErr:
	default:
	}
}
