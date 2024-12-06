package core

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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
	"github.com/google/uuid"
)

var log = slog.Default().With("module", "surf")

type ServerInfo struct {
	Svr       Server
	Marshaler marshal.Marshaler
	CallTable *calltable.CallTable

	OnClientDisconnect func(uid uint32, gateNodeId uint32, reason int32)
	OnClientConnect    func(uid uint32, gateNodeId uint32, ip string)
}

func NewSurf(ninfo *auth.NodeInfo, conf *NodeConf, svrinfo *ServerInfo) (*Surf, error) {
	if svrinfo.Svr == nil {
		return nil, fmt.Errorf("calltable is nil")
	}
	if svrinfo.Marshaler == nil {
		svrinfo.Marshaler = &marshal.ProtoMarshaler{}
	}
	if svrinfo.CallTable == nil {
		svrinfo.CallTable = calltable.NewCallTable()
	}

	surf := &Surf{
		serverInfo: svrinfo,
		ninfo:      ninfo,
		queue:      make(chan func(), 100),
		closed:     make(chan struct{}),
		conf:       conf.SurfConf,
		svrconf:    conf.ServerConf,
		regData: nodeRegistryData{
			Status: 1,
			Node:   ninfo,
			Meta:   registryMeta{},
		},
	}
	err := surf.init()
	if err != nil {
		return nil, err
	}
	return surf, nil
}

type Surf struct {
	conf       SurfConfig
	svrconf    []byte
	serverInfo *ServerInfo
	ninfo      *auth.NodeInfo

	rasPublicKey *rsa.PublicKey
	registry     *registry.EtcdRegistry
	watcher      *registry.EtcdWatch

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

func (s *Surf) init() error {
	pubkey, err := utilRsa.LoadRsaPublicKeyFromUrl(s.conf.PublicKeyFilePath)
	if err != nil {
		return err
	}
	s.rasPublicKey = pubkey

	s.innerCaller = &PacketRouteCaller{
		Calltable: calltable.ExtractMethodFromDesc(msgCore.File_core_proto.Messages(), s),
		Handler:   s,
	}

	s.serverInfo.CallTable.RangeByID(func(key uint32, value *calltable.Method) bool {
		log.Info("handle func", "msgid", key, "funcname", value.Name)
		return true
	})

	err = s.serverInfo.Svr.OnInit(s)
	if err != nil {
		return err
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

func (s *Surf) ServerConf() []byte {
	return s.svrconf
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

	if s.registry != nil {
		s.registry.Close()
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
	if s.serverInfo.Svr != nil {
		s.serverInfo.Svr.OnStop()
	}
	return nil
}

func (s *Surf) UpdateNodeData(status NodeState, data json.RawMessage) error {
	if s.registry == nil {
		return fmt.Errorf("registry not init")
	}

	s.mux.Lock()
	s.regData.Data = data
	s.regData.Status = status
	s.mux.Unlock()

	raw, err := json.Marshal(s.regData)
	if err != nil {
		return err
	}

	return s.registry.UpdateNodeData(string(raw))
}

func (s *Surf) Run() error {
	if len(s.conf.HttpListenAddr) > 1 {
		s.startHttpSvr()
	}

	if len(s.conf.WsListenAddr) > 1 {
		s.startWsSvr()
	}

	if len(s.conf.TcpListenAddr) > 1 {
		s.startTcpSvr()
	}

	defer s.Close()

	log.Info("start gate clients", "addrs", s.conf.GateAddrList)

	for _, addr := range s.conf.GateAddrList {
		log.Info("start gate client", "addr", addr)
		client := network.NewWSClient(network.WSClientOptions{
			RemoteAddress:  addr,
			OnConnPacket:   s.onGatePacket,
			OnConnEnable:   s.onGateStatus,
			AuthToken:      s.ninfo.Marshal(),
			UInfo:          s.ninfo,
			ReconnectDelay: 3 * time.Second,
		})
		client.Start()
	}

	if s.conf.EtcdConf != nil {
		regopts := registry.EtcdRegistryOpts{
			EtcdConf:   *s.conf.EtcdConf,
			NodeId:     fmt.Sprintf("%d", s.ninfo.NodeID()),
			NodeType:   s.ninfo.NodeName(),
			TimeoutSec: 5,
		}
		reg, err := registry.NewEtcdRegistry(regopts)
		if err != nil {
			return err
		}
		s.registry = reg
	}

	s.serverInfo.Svr.OnReady()

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
	log.Info("start http server", "addr", s.conf.HttpListenAddr)

	mux := http.NewServeMux()

	s.serverInfo.CallTable.RangeByName(func(key string, method *calltable.Method) bool {

		svrname := s.ninfo.NodeName()

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
		Addr:    s.conf.HttpListenAddr,
		Handler: mux,
	}
	ln, err := net.Listen("tcp", svr.Addr)
	if err != nil {
		panic(err)
	}

	s.regData.Meta.HttpListenAddr = ln.Addr().String()
	s.httpsvr = svr

	go svr.Serve(ln)
}

func (s *Surf) startWsSvr() {
	log.Info("start ws server", "addr", s.conf.WsListenAddr)

	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   s.conf.WsListenAddr,
		OnConnPacket: s.onGatePacket,
		OnConnEnable: s.onGateStatus,
		OnConnAuth:   s.onConnAuth,
	})
	if err != nil {
		panic(err)
	}

	s.regData.Meta.WsListenAddr = ws.Address()

	s.wssvr = ws
	s.wssvr.Start()
}

func (s *Surf) startTcpSvr() {
	log.Info("start tcp server", "addr", s.conf.TcpListenAddr)

	tcpsvr, err := network.NewTcpServer(network.TcpServerOptions{
		ListenAddr:       s.conf.TcpListenAddr,
		HeatbeatInterval: 30 * time.Second,
		OnConnPacket:     s.onGatePacket,
		OnConnStatus:     s.onGateStatus,
		OnConnAuth:       s.onConnAuth,
	})
	if err != nil {
		panic(err)
	}

	s.regData.Meta.TcpListenAddr = tcpsvr.Address().String()
	s.tcpsvr = tcpsvr
	s.tcpsvr.Start()
}

func (s *Surf) NodeID() uint32 {
	return s.ninfo.NodeID()
}

func (s *Surf) getServerType() uint16 {
	return s.ninfo.NodeType()
}

func (s *Surf) SendRequestToClientByUId(uid uint32, msgid uint32, msg any, cb RequestCallbackFunc) error {
	conn := s.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("not found route")
	}
	return s.SendRequestToClient(conn, uid, msgid, msg, cb)
}

func (s *Surf) SendRequestToClient(conn network.Conn, uid, msgid uint32, msg any, cb RequestCallbackFunc) error {
	body, err := s.serverInfo.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}

	syn := conn.NextSYN()

	rpk := NewRoutePacket(body)
	rpk.SetMsgType(RoutePackMsgType_Request)
	rpk.SetMsgId(msgid)
	rpk.SetToUID(uid)
	rpk.SetToURole(ServerType_Client)
	rpk.SetFromUID(s.NodeID())
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
	body, err := s.serverInfo.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}

	rpk := NewRoutePacket(body)
	rpk.SetMsgType(0)
	rpk.SetToUID(to_uid)
	rpk.SetToURole(to_urole)
	rpk.SetMsgId(msgid)
	rpk.SetFromUID(s.NodeID())
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

func (s *Surf) WrapMethod(url string, method *calltable.Method) http.HandlerFunc {
	// method.Func
	return func(w http.ResponseWriter, r *http.Request) {
		authdata := r.Header.Get("Authorization")
		if len(authdata) < 5 {
			http.Error(w, "Authorization failed", http.StatusUnauthorized)
			return
		}
		authdata = strings.TrimPrefix(authdata, "Bearer ")
		uinfo, err := s.onConnAuth([]byte(authdata))
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
			W:      w,
			R:      r,
			UInfo:  uinfo,
			ConnId: uuid.NewString(),
		}

		method.Call(s.serverInfo.Svr, ctx, req)
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
	log.Info("conn status", "id", c.ConnId(), "uid", c.UserID(), "utype", c.UserRole(), "status", enable)
	if c.UserRole() == ServerType_Client {
		if enable {
			h.notifyClientConnect(c.UserID(), h.ninfo.NodeID(), c.RemoteAddr())
		} else {
			h.notifyClientDisconnect(c.UserID(), h.ninfo.NodeID(), msgCore.NotifyClientDisconnect_Disconnect)
		}
	} else {
		// node:
	}
}

func (s *Surf) catch() {
	if err := recover(); err != nil {
		log.Error("catch panic", "err", err)
	}
}

func (s *Surf) notifyClientConnect(uid uint32, gateNodeId uint32, ip string) {
	log.Info("recv notify client connect", "uid", uid, "gateNodeId", gateNodeId, "ip", ip)
	if s.serverInfo.OnClientConnect != nil {
		s.Do(func() {
			s.serverInfo.OnClientConnect(uid, gateNodeId, ip)
		})
	}
}

func (s *Surf) notifyClientDisconnect(uid uint32, gateNodeId uint32, reason msgCore.NotifyClientDisconnect_Reason) {
	log.Info("recv notify client disconnect", "uid", uid, "gateNodeId", gateNodeId, "reason", reason)
	if s.serverInfo.OnClientDisconnect != nil {
		s.Do(func() {
			s.serverInfo.OnClientDisconnect(uid, gateNodeId, int32(reason))
		})
	}
}

func (s *Surf) OnNotifyClientDisconnect(ctx Context, msg *msgCore.NotifyClientDisconnect) {
	s.notifyClientDisconnect(msg.Uid, msg.GateNodeId, msg.Reason)
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
