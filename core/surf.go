package core

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
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

type Options struct {
	Server Server

	GateToken    []byte
	UInfo        auth.User
	GateAddrList []string

	HttpListenAddr string
	WsListenAddr   string
	TcpListenAddr  string
	CmdListenAddr  string

	Calltable *calltable.CallTable

	Marshaler         marshal.Marshaler
	PublicKeyFilePath string

	OnClientDisconnect func(uid uint32, gateNodeId uint32, reason int32)

	EtcdConf *registry.EtcdConfig
}

type nodeRegistryData struct {
	NodeStatus int             `json:"node_status"`
	SurfData   map[string]any  `json:"surf_data"`
	ServerData json.RawMessage `json:"server_data"`
}

func NewSurf(opt Options) (*Surf, error) {
	s := &Surf{
		queue:  make(chan func(), 100),
		closed: make(chan struct{}),
	}
	err := s.Init(opt)
	return s, err
}

type Surf struct {
	opts         Options
	rasPublicKey *rsa.PublicKey

	reg *registry.EtcdRegistry

	tcpsvr  *network.TcpServer
	wssvr   *network.WSServer
	httpsvr *http.Server

	serverCaller *PacketRouteCaller
	innerCaller  *PacketRouteCaller

	queue  chan func()
	closed chan struct{}

	mux sync.Mutex

	regData nodeRegistryData
}

func (s *Surf) Init(opt Options) error {
	pubkey, err := utilRsa.LoadRsaPublicKeyFromUrl(opt.PublicKeyFilePath)
	if err != nil {
		return err
	}
	s.rasPublicKey = pubkey

	if opt.Marshaler == nil {
		opt.Marshaler = &marshal.ProtoMarshaler{}
	}

	s.opts = opt

	s.serverCaller = &PacketRouteCaller{
		Calltable: opt.Calltable,
		Handler:   opt.Server,
	}

	s.innerCaller = &PacketRouteCaller{
		Calltable: calltable.ExtractMethodFromDesc(msgCore.File_core_proto.Messages(), s),
		Handler:   s,
	}

	if opt.EtcdConf != nil {
		regopts := registry.EtcdRegistryOpts{
			EtcdConf:   *opt.EtcdConf,
			NodeId:     fmt.Sprintf("%d", opt.UInfo.UserID()),
			ServerType: uint16(opt.Server.ServerType()),
			ServerName: opt.Server.ServerName(),
			TimeoutSec: 5,
		}
		reg, err := registry.NewEtcdRegistry(regopts)
		if err != nil {
			return err
		}
		s.reg = reg

		s.regData.NodeStatus = 1
	}

	return nil
}

func (s *Surf) Do(fn func()) {
	select {
	case <-s.closed:
		return
	default:
		select {
		case s.queue <- fn:
		default:
			log.Error("queue full, drop fn")
		}
	}
}

func (s *Surf) Close() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	select {
	case <-s.closed:
		return nil
	default:
		close(s.closed)

		close(s.queue)
	}

	if s.reg != nil {
		s.reg.Close()
	}

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

func (s *Surf) UpdateNodeData(status int, newdata json.RawMessage) error {
	s.mux.Lock()
	s.regData.ServerData = newdata
	s.regData.NodeStatus = status
	s.mux.Unlock()

	raw, err := json.Marshal(s.regData)
	if err != nil {
		return err
	}

	return s.reg.UpdateNodeData(string(raw))
}

func (s *Surf) Run() error {
	s.opts.Calltable.RangeByID(func(key uint32, value *calltable.Method) bool {
		log.Info("handle func", "msgid", key, "funcname", value.Name)
		return true
	})

	if len(s.opts.HttpListenAddr) > 1 {
		s.startHttpSvr()
	}

	if len(s.opts.WsListenAddr) > 1 {
		s.startWsSvr()
	}

	if len(s.opts.TcpListenAddr) > 1 {
		s.startTcpSvr()
	}

	defer s.Close()

	log.Info("start gate clients", "addrs", s.opts.GateAddrList)

	for _, addr := range s.opts.GateAddrList {
		log.Info("start gate client", "addr", addr)
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

	if s.reg != nil {
		s.UpdateNodeData(1, nil)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, utilSignal.ShutdownSignals()...)

	for {
		select {
		case s := <-signals:
			log.Info("recv close signal", "signal", s)
			return nil
		case <-s.closed:
			log.Info("surf closed")
			return nil
		case fn, ok := <-s.queue:
			if !ok {
				log.Error("queue closed")
				return nil
			}
			fn()
		}
	}
}

func (s *Surf) startHttpSvr() {
	log.Info("start http server", "addr", s.opts.HttpListenAddr)

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
		log.Info("http handle func", "path", key)
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
	log.Info("start ws server", "addr", s.opts.WsListenAddr)

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
	log.Info("start tcp server", "addr", s.opts.TcpListenAddr)

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

func (s *Surf) getNodeId() uint32 {
	return s.opts.UInfo.UserID()
}

func (s *Surf) getServerType() uint16 {
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

	syn := conn.NextSYN()

	rpk := NewRoutePacket(body)
	rpk.SetMsgType(RoutePackMsgType_Request)
	rpk.SetMsgId(msgid)
	rpk.SetToUID(uid)
	rpk.SetToURole(ServerType_Client)
	rpk.SetFromUID(s.getNodeId())
	rpk.SetFromURole(s.getServerType())
	rpk.SetSYN(syn)

	s.serverCaller.PushRespCallback(conn.UserID(), syn, 3, cb)

	pk := rpk.ToHVPacket()

	err = conn.Send(pk)
	if err != nil {
		s.serverCaller.PopRespCallback(conn.UserID(), syn)
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
	rpk.SetFromUID(s.getNodeId())
	rpk.SetFromURole(s.getServerType())

	log.Info("SendAsyncToClient", "from", rpk.GetFromUID(), "fromrole", rpk.GetFromURole(), "to", rpk.GetToUID(),
		"torole", rpk.GetToURole(), "msgid", rpk.GetMsgId(), "msgtype", rpk.GetMsgType())

	return conn.Send(rpk.ToHVPacket())
}

func (s *Surf) SendAsyncToClientByUId(uid uint32, msgid uint32, msg any) error {
	conn := s.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("not found route")
	}
	return s.SendAsyncToClient(conn, uid, ServerType_Client, msgid, msg)
}

func (s *Surf) SendToNode(nodeid uint32, svrtype uint16, pk *network.HVPacket) error {
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
		uinfo, err := auth.VerifyToken(s.rasPublicKey, []byte(authdata))
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
		log.Error("invalid packet type", "type", pk.Meta.GetType())
	}
}

func (h *Surf) onConnAuth(data []byte) (network.User, error) {
	return auth.VerifyToken(h.rasPublicKey, data)
}

func (h *Surf) onGateStatus(c network.Conn, enable bool) {
	// TOOD:
	log.Info("conn status", "id", c.ConnID(), "uid", c.UserID(), "utype", c.UserRole(), "status", enable)
}

func (s *Surf) catch() {
	if err := recover(); err != nil {
		log.Error("catch panic", "err", err)
	}
}

func (s *Surf) OnNotifyClientDisconnect(ctx Context, msg *msgCore.NotifyClientDisconnect) {
	log.Info("recv notify client disconnect", "uid", msg.Uid, "gateNodeId", msg.GateNodeId, "reason", msg.Reason)
	if s.opts.OnClientDisconnect != nil {
		s.Do(func() {
			s.opts.OnClientDisconnect(msg.Uid, msg.GateNodeId, int32(msg.Reason))
		})
	}
}

func (s *Surf) onRoutePacket(c network.Conn, rpk *RoutePacket) {
	marshaler := marshal.NewMarshalerById(rpk.GetMarshalType())
	if marshaler == nil {
		log.Error("invalid marshaler type", "type", rpk.GetMarshalType())
		return
	}

	ctx := &ConnContext{
		Conn:      c,
		Core:      s,
		ReqPacket: rpk,
		Marshal:   marshaler,
	}

	s.Do(func() {
		if rpk.GetToURole() == ServerType_Core {
			s.innerCaller.Call(ctx)
		} else {
			s.serverCaller.Call(ctx)
		}
	})
}
