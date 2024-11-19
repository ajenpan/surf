package core

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/registry"
	"github.com/ajenpan/surf/core/utils/calltable"
	utilRsa "github.com/ajenpan/surf/core/utils/rsagen"
	utilSignal "github.com/ajenpan/surf/core/utils/signal"
)

type Server interface {
	ServerType() uint16
	ServerName() string
}

type Options struct {
	Server Server

	GateToken    []byte
	UInfo        auth.User
	GateAddrList []string

	HttpListenAddr string
	WsListenAddr   string
	TcpListenAddr  string

	RouteCallTable *calltable.CallTable

	Marshaler         marshal.Marshaler
	PublicKeyFilePath string
}

func NewSurf(opt Options) (*Surf, error) {
	s := &Surf{}
	err := s.Init(opt)
	return s, err
}

type Surf struct {
	opts      Options
	PublicKey *rsa.PublicKey

	Reg *registry.Registry

	tcpsvr  *network.TcpServer
	wssvr   *network.WSServer
	httpsvr *http.Server
	nodeid  uint32

	routeCaller *PacketRouteCaller
}

func (s *Surf) Init(opt Options) error {
	pubkey, err := utilRsa.LoadRsaPublicKeyFromUrl(opt.PublicKeyFilePath)
	if err != nil {
		return err
	}
	s.PublicKey = pubkey

	if opt.RouteCallTable == nil {
		opt.RouteCallTable = calltable.NewCallTable()
	}

	if opt.Marshaler == nil {
		opt.Marshaler = &marshal.ProtoMarshaler{}
	}

	s.opts = opt

	s.routeCaller = &PacketRouteCaller{
		calltable: opt.RouteCallTable,
		Handler:   opt.Server,
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
	if len(s.opts.HttpListenAddr) > 1 {
		s.startHttpSvr()
	}

	if len(s.opts.WsListenAddr) > 1 {
		s.startWsSvr()
	}

	if len(s.opts.TcpListenAddr) > 1 {
		s.startTcpSvr()
	}

	return nil
}

func (s *Surf) Run() error {
	s.opts.RouteCallTable.RangeByID(func(key uint32, value *calltable.Method) bool {
		log.Infof("handle func,msgid:%d, funcname:%s", key, value.Name)
		return true
	})

	if err := s.Start(); err != nil {
		return err
	}
	defer s.Close()

	log.Infof("start gate clients:%v", s.opts.GateAddrList)

	for _, addr := range s.opts.GateAddrList {
		log.Infof("start gate client, addr: %s", addr)
		client := network.NewWSClient(network.WSClientOptions{
			RemoteAddress:  addr,
			OnConnPacket:   s.onGatePacket,
			OnConnEnable:   s.onGateStatus,
			AuthToken:      []byte(s.opts.GateToken),
			UInfo:          s.opts.UInfo,
			ReconnectDelay: 3 * time.Second,
		})
		client.Start()
	}

	sig := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", sig.String())
	return nil
}

func (s *Surf) startHttpSvr() {
	log.Info("start http server, listenaddr ", s.opts.HttpListenAddr)

	mux := http.NewServeMux()

	s.opts.RouteCallTable.RangeByName(func(key string, method *calltable.Method) bool {

		svrname := s.opts.Server.ServerName()

		if len(svrname) > 0 {
			key = "/" + svrname + "/" + key
		} else {
			key = "/" + key
		}

		cb := s.WrapMethod(key, method)
		mux.HandleFunc(key, cb)
		log.Info("http handle func: ", key)
		return true
	})

	svr := &http.Server{
		Addr:    s.opts.HttpListenAddr,
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
	log.Info("start ws server, listenaddr ", s.opts.WsListenAddr)

	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   s.opts.WsListenAddr,
		OnConnPacket: s.onGatePacket,
		OnConnEnable: s.onGateStatus,
		OnConnAuth:   s.onConnAuth,
	})
	if err != nil {
		panic(err)
	}

	s.wssvr = ws
	s.wssvr.Start()
}

func (s *Surf) startTcpSvr() {
	log.Info("start tcp server, listenaddr ", s.opts.TcpListenAddr)

	tcpsvr, err := network.NewTcpServer(network.TcpServerOptions{
		ListenAddr:       s.opts.TcpListenAddr,
		HeatbeatInterval: 30 * time.Second,
		OnConnPacket:     s.onGatePacket,
		OnConnStatus:     s.onGateStatus,
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
	return uint16(s.opts.Server.ServerType())
}

func (s *Surf) SendRequestToClientByUId(uid uint32, msgid uint32, msg any, cb RequestCallbackFunc) error {
	conn := s.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("not found route")
	}
	return s.SendRequestToClient(conn, uid, msgid, msg, cb)
}

func (s *Surf) SendRequestToClient(conn network.Conn, uid, msgid uint32, msg any, cb RequestCallbackFunc) error {
	body, err := s.opts.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}
	syn := s.routeCaller.GetSYN()

	rpk := NewRoutePacket(body)
	rpk.SetMsgType(RoutePackMsgType_Request)
	rpk.SetMsgId(msgid)
	rpk.SetToUID(uid)
	rpk.SetToURole(ServerType_Client)
	rpk.SetFromUID(s.GetNodeId())
	rpk.SetFromURole(s.GetServerType())
	rpk.SetSYN(syn)

	s.routeCaller.pushRespCallback(syn, cb)

	pk := rpk.ToHVPacket()

	err = conn.Send(pk)
	if err != nil {
		s.routeCaller.popRespCallback(syn)
	}
	return err
}

func (s *Surf) SendAsyncToClient(conn network.Conn, uid uint32, msgid uint32, msg any) error {
	body, err := s.opts.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}

	rpk := NewRoutePacket(body)
	rpk.SetMsgType(0)
	rpk.SetToUID(uid)
	rpk.SetToURole(ServerType_Client)
	rpk.SetMsgId(msgid)
	rpk.SetFromUID(s.GetNodeId())
	rpk.SetFromURole(s.GetServerType())
	rpk.SetSYN(s.routeCaller.GetSYN())

	return conn.Send(rpk.ToHVPacket())
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

func (s *Surf) WrapMethod(url string, method *calltable.Method) http.HandlerFunc {
	// method.Func
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

		if err = json.Unmarshal(raw, req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// contenttype := r.Header.Get("Content-Type")
		// w.Header().Set("Content-Type", contenttype)

		var ctx Context = &httpCallContext{
			w:     w,
			r:     r,
			core:  s,
			uinfo: uinfo,
		}

		method.Call(s.opts.Server, ctx, req)
	}
}

func (h *Surf) onGatePacket(s network.Conn, pk *network.HVPacket) {
	switch pk.Meta.GetType() {
	case network.PacketType_Route:
		h.onRoutePacket(s, pk)
	case network.PacketType_Node:
		h.onNodePacket(s, pk)
	default:
	}
}

func (h *Surf) onConnAuth(data []byte) (network.User, error) {
	return auth.VerifyToken(h.PublicKey, data)
}

func (h *Surf) onGateStatus(c network.Conn, enable bool) {
	// TOOD:
	log.Infof("connid:%v, uid:%v,utype:%v, status:%v", c.ConnID(), c.UserID(), c.UserRole(), enable)
}

func (s *Surf) catch() {
	if err := recover(); err != nil {
		log.Error(err)
	}
}

func (s *Surf) onNodePacket(c network.Conn, pk *network.HVPacket) {
	npk := NewNodePacket(nil).FromHVPacket(pk)
	switch npk.GetMsgType() {
	case NodePackMsgType_Notify:

	case NodePackMsgType_Async:
	case NodePackMsgType_Request:
		if npk.GetMsgId() == 0 {
			log.Error("invalid async msgid:", npk.GetMsgId())
			return
		}
		marshaler := marshal.NewMarshalerById(npk.GetMarshalType())
		if marshaler == nil {
			log.Error("invalid marshaler type:", npk.GetMarshalType())
			return
		}
	}
}

func (s *Surf) onRoutePacket(c network.Conn, pk *network.HVPacket) {
	if len(pk.Head) != RoutePackHeadLen {
		log.Error("invalid packet head length:", len(pk.Head))
		return
	}
	rpk := NewRoutePacket(nil).FromHVPacket(pk)
	marshaler := marshal.NewMarshalerById(rpk.GetMarshalType())
	if marshaler == nil {
		log.Error("invalid marshaler type:", rpk.GetMarshalType())
		return
	}
	ctx := &connContext{
		Conn:      c,
		Core:      s,
		ReqPacket: rpk,
		uid:       rpk.GetFromUID(),
		urole:     uint32(rpk.GetFromURole()),
		Marshal:   marshaler,
	}
	s.routeCaller.Call(ctx)
}
