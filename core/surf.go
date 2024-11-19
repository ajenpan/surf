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
	msgCore "github.com/ajenpan/surf/msg/core"
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

	Calltable *calltable.CallTable

	Marshaler         marshal.Marshaler
	PublicKeyFilePath string

	OnClientDisconnect func(uid uint32, gateNodeId uint32, reason int32)
}

func NewSurf(opt Options) (*Surf, error) {
	s := &Surf{}
	err := s.Init(opt)
	return s, err
}

type Surf struct {
	opts      Options
	PublicKey *rsa.PublicKey

	Reg *registry.EtcdRegistry

	tcpsvr  *network.TcpServer
	wssvr   *network.WSServer
	httpsvr *http.Server
	nodeid  uint32

	serverCaller *PacketRouteCaller
	innerCaller  *PacketRouteCaller
}

func (s *Surf) Init(opt Options) error {
	pubkey, err := utilRsa.LoadRsaPublicKeyFromUrl(opt.PublicKeyFilePath)
	if err != nil {
		return err
	}
	s.PublicKey = pubkey

	if opt.Marshaler == nil {
		opt.Marshaler = &marshal.ProtoMarshaler{}
	}

	s.opts = opt

	s.serverCaller = &PacketRouteCaller{
		calltable: opt.Calltable,
		Handler:   opt.Server,
	}

	s.innerCaller = &PacketRouteCaller{
		calltable: calltable.ExtractProtoFile(msgCore.File_core_proto, s),
		Handler:   s,
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
	s.opts.Calltable.RangeByID(func(key uint32, value *calltable.Method) bool {
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

	s.opts.Calltable.RangeByName(func(key string, method *calltable.Method) bool {

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
	syn := s.serverCaller.GetSYN()

	rpk := NewRoutePacket(body)
	rpk.SetMsgType(RoutePackMsgType_Request)
	rpk.SetMsgId(msgid)
	rpk.SetToUID(uid)
	rpk.SetToURole(ServerType_User)
	rpk.SetFromUID(s.GetNodeId())
	rpk.SetFromURole(s.GetServerType())
	rpk.SetSYN(syn)

	s.serverCaller.pushRespCallback(syn, cb)

	pk := rpk.ToHVPacket()

	err = conn.Send(pk)
	if err != nil {
		s.serverCaller.popRespCallback(syn)
	}
	return err
}

func (s *Surf) SendAsyncToClient(conn network.Conn, to_uid uint32, to_urole uint16, msgid uint32, msg any) error {
	body, err := s.opts.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}

	rpk := NewRoutePacket(body)
	rpk.SetMsgType(0)
	rpk.SetToUID(to_uid)
	rpk.SetToURole(to_urole)
	rpk.SetMsgId(msgid)
	rpk.SetFromUID(s.GetNodeId())
	rpk.SetFromURole(s.GetServerType())
	rpk.SetSYN(s.serverCaller.GetSYN())

	return conn.Send(rpk.ToHVPacket())
}

func (s *Surf) SendAsyncToClientByUId(uid uint32, msgid uint32, msg any) error {
	conn := s.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("not found route")
	}
	return s.SendAsyncToClient(conn, uid, ServerType_User, msgid, msg)
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

		var ctx Context = &HttpContext{
			W:     w,
			R:     r,
			UInfo: uinfo,
		}

		method.Call(s.opts.Server, ctx, req)
	}
}

func (h *Surf) onGatePacket(s network.Conn, pk *network.HVPacket) {
	switch pk.Meta.GetType() {
	case network.PacketType_Route:
		rpk := NewRoutePacket(nil).FromHVPacket(pk)
		h.onRoutePacket(s, rpk)
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

func (s *Surf) OnNotifyClientDisconnect(ctx Context, msg *msgCore.NotifyClientDisconnect) {
	log.Infof("recv notify client disconnect, uid:%v, gateNodeId:%v, reason:%v", msg.Uid, msg.GateNodeId, msg.Reason)

	if s.opts.OnClientDisconnect != nil {
		s.opts.OnClientDisconnect(msg.Uid, msg.GateNodeId, int32(msg.Reason))
	}
}

func (s *Surf) onRoutePacket(c network.Conn, rpk *RoutePacket) {
	marshaler := marshal.NewMarshalerById(rpk.GetMarshalType())
	if marshaler == nil {
		log.Error("invalid marshaler type:", rpk.GetMarshalType())
		return
	}

	ctx := &ConnContext{
		Conn:      c,
		Core:      s,
		ReqPacket: rpk,
		Marshal:   marshaler,
	}

	if rpk.GetToURole() == ServerType_Core {
		s.innerCaller.Call(ctx)
	} else {
		s.serverCaller.Call(ctx)
	}
}
