package core

import (
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/registry"
	"github.com/ajenpan/surf/core/utils/calltable"
	utilSignal "github.com/ajenpan/surf/core/utils/signal"
)

type Server interface {
	ServerType() uint16
	ServerName() string
}

type Options struct {
	Server    Server
	GateToken []byte
	UInfo     auth.User

	ControlListenAddr string
	HttpListenAddr    string
	WsListenAddr      string
	TcpListenAddr     string
	GateAddrList      []string

	CTByName *calltable.CallTable[string]
	CTById   *calltable.CallTable[uint32]

	Marshaler marshal.Marshaler
	PublicKey *rsa.PublicKey
}

func converCalltable(source *calltable.CallTable[uint32]) *calltable.CallTable[string] {
	result := calltable.NewCallTable[string]()
	if source == nil {
		return result
	}

	source.Range(func(key uint32, value *calltable.Method) bool {
		result.Add(value.HandleName, value)
		return true
	})
	return result
}

func NewSurf(opt Options) *Surf {
	s := &Surf{}
	err := s.init(opt)
	if err != nil {
		panic(err)
	}
	return s
}

type RequestCallbackFunc func(timeout bool, pk *network.HVPacket)

type RequestCallbackCache struct {
	cbfun   RequestCallbackFunc
	timeout *time.Timer
}

type Surf struct {
	Options
	Reg *registry.Registry

	tcpsvr  *network.TcpServer
	wssvr   *network.WSServer
	httpsvr *http.Server
	nodeid  uint32

	respWatier sync.Map
	synIdx     uint32
}

func (s *Surf) init(opt Options) error {
	s.Options = opt
	if s.CTById == nil {
		s.CTById = calltable.NewCallTable[uint32]()
	}
	if s.CTByName == nil {
		s.CTByName = converCalltable(s.CTById)
	} else {
		s.CTByName.Merge(converCalltable(s.CTById), false)
	}
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

	if len(s.ControlListenAddr) > 1 {
		// todo:
	}
	// quit := make(chan struct{})
	// s.httpsvr = &network.HttpSvr{
	// 	Addr:    s.HttpListenAddr,
	// 	Marshal: &marshal.JSONPb{},
	// 	Mux:     http.NewServeMux(),
	// }
	// s.httpsvr.ServerCallTable(&s.CTByName)

	return nil
}

func (s *Surf) Run() error {
	s.CTById.Range(func(key uint32, value *calltable.Method) bool {
		log.Infof("handle func,msgid:%d, funcname:%s", key, value.HandleName)
		return true
	})

	if err := s.Start(); err != nil {
		return err
	}
	defer s.Close()

	log.Infof("start gate clients:%v", s.GateAddrList)

	for _, addr := range s.GateAddrList {
		log.Infof("start gate client, addr: %s", addr)
		client := network.NewWSClient(network.WSClientOptions{
			RemoteAddress:  addr,
			OnConnPacket:   s.onConnPacket,
			OnConnStatus:   s.onConnStatus,
			AuthToken:      []byte(s.Options.GateToken),
			UInfo:          s.Options.UInfo,
			ReconnectDelay: 3 * time.Second,
		})
		client.Start()
	}

	sig := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", sig.String())
	return nil
}

func (s *Surf) startHttpSvr() {
	log.Info("start http server, listenaddr ", s.HttpListenAddr)

	mux := http.NewServeMux()

	s.CTByName.Range(func(key string, method *calltable.Method) bool {

		svrname := s.Server.ServerName()

		if len(svrname) > 0 {
			key = "/" + svrname + "/" + key
		} else {
			key = "/" + key
		}

		cb := s.WrapMethod(method)
		mux.HandleFunc(key, cb)
		log.Info("http handle func: ", key)
		return true
	})

	svr := &http.Server{
		Addr:    s.HttpListenAddr,
		Handler: mux,
	}
	ln, err := net.Listen("tcp", svr.Addr)
	if err != nil {
		panic(err)
	}

	s.httpsvr = svr
	go svr.Serve(ln)
}

func (s *Surf) startWsSvr() {
	log.Info("start ws server, listenaddr ", s.WsListenAddr)

	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   s.WsListenAddr,
		OnConnPacket: s.onConnPacket,
		OnConnStatus: s.onConnStatus,
		OnConnAuth:   s.onConnAuth,
	})
	if err != nil {
		panic(err)
	}

	s.wssvr = ws
	s.wssvr.Start()
}

func (s *Surf) startTcpSvr() {
	log.Info("start tcp server, listenaddr ", s.TcpListenAddr)

	tcpsvr, err := network.NewTcpServer(network.TcpServerOptions{
		ListenAddr:       s.TcpListenAddr,
		HeatbeatInterval: 30 * time.Second,
		OnConnPacket:     s.onConnPacket,
		OnConnStatus:     s.onConnStatus,
		OnConnAuth:       s.onConnAuth,
	})
	if err != nil {
		panic(err)
	}

	s.tcpsvr = tcpsvr
	s.tcpsvr.Start()
}

func (s *Surf) GetNodeId() uint32 {
	return s.nodeid
}

func (s *Surf) GetServerType() uint16 {
	return uint16(s.Server.ServerType())
}

func (s *Surf) GetSYN() uint32 {
	ret := atomic.AddUint32(&s.synIdx, 1)
	if ret == 0 {
		return atomic.AddUint32(&s.synIdx, 1)
	}
	return ret
}

func (s *Surf) pushRespCallback(syn uint32, cb RequestCallbackFunc) error {
	timeout := time.AfterFunc(3*time.Second, func() {
		info := s.popRespCallback(syn)
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

	s.respWatier.Store(syn, cache)
	return nil
}

func (s *Surf) popRespCallback(syn uint32) *RequestCallbackCache {
	cache, ok := s.respWatier.Load(syn)
	if !ok {
		return nil
	}
	return cache.(*RequestCallbackCache)
}

// func (s *Surf) SendResponeToClient(uid uint32, syn uint32, errcode uint16, msgid uint32, msg any) error {
// 	conn := s.GetClientConn(uid)
// 	if conn == nil {
// 		return fmt.Errorf("not found route")
// 	}
// 	body, err := s.Marshaler.Marshal(msg)
// 	if err != nil {
// 		return err
// 	}
// 	head := network.RoutePacketHead(make([]byte, network.RoutePackHeadLen))
// 	head.SetClientId(uid)
// 	head.SetMsgId(msgid)
// 	head.SetNodeId(s.GetNodeId())
// 	head.SetSYN(syn)
// 	pk := network.NewRoutePacket(network.RoutePackType_SubFlag_Response, head, body)
// 	return conn.Send(pk)
// }

func (s *Surf) SendRequestToClientByUId(uid uint32, msgid uint32, msg any, cb RequestCallbackFunc) error {
	conn := s.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("not found route")
	}
	return s.SendRequestToClient(conn, uid, msgid, msg, cb)
}

func (s *Surf) SendRequestToClient(conn network.Conn, uid, msgid uint32, msg any, cb RequestCallbackFunc) error {
	body, err := s.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}
	syn := s.GetSYN()
	head := network.RoutePacketHead(make([]byte, network.RoutePackHeadLen))
	head.SetClientId(uid)
	head.SetMsgId(msgid)
	head.SetNodeId(s.GetNodeId())
	head.SetSYN(syn)

	s.pushRespCallback(syn, cb)

	pk := network.NewRoutePacket(network.RoutePackType_SubFlag_Request, head, body)

	err = conn.Send(pk)
	if err != nil {
		s.popRespCallback(syn)
	}
	return err
}

func (s *Surf) SendAsyncToClient(conn network.Conn, uid uint32, msgid uint32, msg any) error {
	body, err := s.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}

	head := network.RoutePacketHead(make([]byte, network.RoutePackHeadLen))
	head.SetClientId(uid)
	head.SetMsgId(msgid)
	head.SetNodeId(s.GetNodeId())
	head.SetSYN(s.GetSYN())

	pk := network.NewRoutePacket(network.RoutePackType_SubFlag_Async, head, body)
	return conn.Send(pk)
}

func (s *Surf) SendAsyncToClientByUId(uid uint32, msgid uint32, msg any) error {
	conn := s.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("not found route")
	}
	return s.SendAsyncToClient(conn, uid, msgid, msg)
}

func (s *Surf) SendToNode(nodeid uint32, pk *network.HVPacket) error {
	// todo
	return nil
}

func (s *Surf) GetClientConn(id uint32) network.Conn {
	// todo
	return nil
}

func (s *Surf) GetNodeConn(id uint32) network.Conn {
	// todo
	return nil
}

func (s *Surf) WrapMethod(method *calltable.Method) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authdata := r.Header.Get("Authorization")
		if len(authdata) < 5 {
			http.Error(w, "Authorization failed", http.StatusUnauthorized)
			return
		}
		authdata = strings.TrimPrefix(authdata, "Bearer ")

		uinfo, err := auth.VerifyToken(s.PublicKey, []byte(authdata))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		req := method.NewRequest()
		marshaler := marshal.NewMarshalerByName(r.Header.Get("Content-Type"))
		if marshaler == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = marshaler.Unmarshal(raw, req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		contenttype := r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", contenttype)

		var ctx Context = &HttpCallContext{
			w:         w,
			r:         r,
			core:      s,
			uinfo:     uinfo,
			marshaler: marshaler,
		}

		method.Call(s.Server, ctx, req)
	}
}

func (h *Surf) onConnPacket(s network.Conn, pk *network.HVPacket) {
	switch pk.Meta.GetType() {
	case network.PacketType_Route:
		h.onRoutePacket(s, pk)
	case network.PacketType_RouteNode:
		h.onNodeInnerPacket(s, pk)
	default:
	}
}

func (h *Surf) onConnAuth(data []byte) (auth.User, error) {
	return auth.VerifyToken(h.PublicKey, data)
}
func (h *Surf) onConnStatus(c network.Conn, enable bool) {
	log.Infof("connid:%v, uid:%v status:%v", c.ConnID(), c.UserID(), enable)
}

func (s *Surf) catch() {
	if err := recover(); err != nil {
		log.Error(err)
	}
}

func (s *Surf) onNodeInnerPacket(c network.Conn, pk *network.HVPacket) {
	//todo:
}

func (s *Surf) onRoutePacket(c network.Conn, pk *network.HVPacket) {
	defer s.catch()

	if len(pk.Head) != network.RoutePackHeadLen {
		log.Error("invalid packet head length:", len(pk.Head))
		return
	}

	head := network.RoutePacketHead(pk.Head)

	switch pk.Meta.GetSubFlag() {
	case network.RoutePackType_SubFlag_Async:
		fallthrough
	case network.RoutePackType_SubFlag_Request:
		method := s.CTById.Get(head.GetMsgId())
		if method == nil {
			// todo:
			return
		}
		marshaler := marshal.NewMarshalerById(head.GetMarshalType())
		req := method.NewRequest()
		err := marshaler.Unmarshal(pk.GetBody(), req)
		if err != nil {
			// todo:
			return
		}

		ctx := &context{
			Conn:    c,
			Core:    s,
			Pk:      pk,
			caller:  head.GetClientId(),
			Marshal: marshaler,
		}

		method.Call(s.Server, ctx, req)
	case network.RoutePackType_SubFlag_Response:
		cbinfo := s.popRespCallback(head.GetSYN())
		if cbinfo == nil {
			return
		}
		if cbinfo.timeout != nil {
			cbinfo.timeout.Stop()
		}
		if cbinfo.cbfun != nil {
			cbinfo.cbfun(false, pk)
		}
	case network.RoutePackType_SubFlag_RouteErr:
	default:
	}
}
