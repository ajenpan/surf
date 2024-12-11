package core

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/registry"
	utilSignal "github.com/ajenpan/surf/core/utils/signal"
)

var log = slog.Default().With("module", "surf")

var DefaultRequestTimeoutSec uint32 = 3

type ServerInfo struct {
	Svr       Server
	Marshaler marshal.Marshaler

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
		clientGateMap:  make(map[uint32]*clientGateItem),
		gateConnMap:    make(map[string]network.Conn),
		gateHoldUserid: make(map[string]map[uint32]*clientGateItem),
	}
	err := surf.init()
	if err != nil {
		return nil, err
	}
	return surf, nil
}

type clientGateItem struct {
	conn     network.Conn
	clientIp string
	connAt   time.Time
}

type Surf struct {
	conf       SurfConfig
	svrconf    []byte
	serverInfo *ServerInfo
	ninfo      *auth.NodeInfo

	rasPublicKey *rsa.PublicKey
	registry     *registry.EtcdRegistry
	nodeWatcher  *registry.EtcdWatcher

	tcpsvr  *network.TcpServer
	wssvr   *network.WSServer
	httpsvr *http.Server

	caller *PacketRouteCaller

	queue  chan func()
	closed chan struct{}

	mux sync.Mutex

	regData nodeRegistryData

	clientGateMap  map[uint32]*clientGateItem
	gateConnMap    map[string]network.Conn
	gateHoldUserid map[string]map[uint32]*clientGateItem
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

func (s *Surf) PublicKey() *rsa.PublicKey {
	return s.rasPublicKey
}

func (s *Surf) NodeID() uint32 {
	return s.ninfo.NodeID()
}

func (s *Surf) NodeType() uint16 {
	return s.ninfo.NodeType()
}

func (s *Surf) NodeName() string {
	return s.ninfo.NName
}

func (s *Surf) NodeInfo() auth.NodeInfo {
	return *s.ninfo
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

	if err := s.startNodeRegistry(); err != nil {
		return err
	}

	if err := s.startNodeWatcher(); err != nil {
		return err
	}

	s.connectGates()

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

func (s *Surf) SendRequestToClient(uid uint32, msg proto.Message, cb RequestCallbackFunc) error {
	conn := s.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("conn not found")
	}
	return s.SendRequest(conn, NodeType_Client, uid, msg, cb)
}

func (s *Surf) SendAsyncToClient(uid uint32, msg proto.Message) error {
	conn := s.GetClientConn(uid)
	if conn == nil {
		return fmt.Errorf("conn not found")
	}
	return s.SendAsync(conn, NodeType_Client, uid, msg)
}

func (s *Surf) SendRequestToNode(ntype uint16, nid uint32, msg proto.Message, cb RequestCallbackFunc) error {
	if ntype == 0 {
		return fmt.Errorf("err node type")
	}
	if nid == 0 {
		nid = s.NextNodeId(ntype)
	}
	conn := s.GetClientConn(nid)
	if conn == nil {
		return fmt.Errorf("conn not found")
	}
	return s.SendRequest(conn, ntype, nid, msg, cb)
}

func (s *Surf) SendAsyncToNode(ntype uint16, nid uint32, msg proto.Message) error {
	conn := s.GetClientConn(nid)
	if conn == nil {
		return fmt.Errorf("conn not found")
	}
	return s.SendAsync(conn, ntype, nid, msg)
}

func (s *Surf) SendAsync(conn network.Conn, urole uint16, uid uint32, msg proto.Message) error {
	body, err := s.serverInfo.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}
	rpk := NewRoutePacket(body)
	rpk.SetToUId(uid)
	rpk.SetToURole(urole)
	rpk.SetFromUId(s.NodeID())
	rpk.SetFromURole(s.NodeType())
	rpk.SetMsgType(RoutePackMsgType_Async)
	rpk.SetMarshalType(s.serverInfo.Marshaler.Id())
	err = conn.Send(rpk.ToHVPacket())
	log.Debug("SendAsync", "from", rpk.GetFromUId(), "fromrole", rpk.GetFromURole(), "to", rpk.GetToUId(),
		"torole", rpk.GetToURole(), "msgid", rpk.GetMsgId(), "msgtype", rpk.GetMsgType(), "err", err)
	return err
}

func (s *Surf) SendRequest(conn network.Conn, urole uint16, uid uint32, msg proto.Message, cb RequestCallbackFunc) error {
	syn := conn.NextSYN()
	body, err := s.serverInfo.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}
	rpk := NewRoutePacket(body)
	rpk.SetToUId(uid)
	rpk.SetToURole(urole)
	rpk.SetFromUId(s.NodeID())
	rpk.SetFromURole(s.NodeType())
	rpk.SetMsgType(RoutePackMsgType_Request)
	rpk.SetMarshalType(s.serverInfo.Marshaler.Id())
	rpk.SetSYN(syn)
	cbkey := ResponseRouteKey{conn.UserID(), syn}
	s.caller.PushRespCallback(cbkey, DefaultRequestTimeoutSec, cb)
	err = conn.Send(rpk.ToHVPacket())
	if err != nil {
		s.caller.PopRespCallback(cbkey)
	}
	log.Debug("SendRequest", "from", rpk.GetFromUId(), "fromrole", rpk.GetFromURole(), "to", rpk.GetToUId(),
		"torole", rpk.GetToURole(), "msgid", rpk.GetMsgId(), "msgtype", rpk.GetMsgType(), "err", err)
	return err
}

func (s *Surf) GetClientConn(uid uint32) network.Conn {
	item, has := s.clientGateMap[uid]
	if !has {
		return nil
	}
	return item.conn
}

func (s *Surf) HandleFuncs(msgid uint32, chain ...HandlerFunc) {
	if len(chain) == 0 {
		return
	}
	s.caller.handlers.Add(msgid, chain)
}

func (s *Surf) NextNodeId(ntype uint16) uint32 {
	return 0
}
