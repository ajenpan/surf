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

type NodeType = uint16

const (
	NodeType_Client NodeType = 0
	NodeType_Core   NodeType = 100
	NodeType_Gate   NodeType = 101

	NodeName_Gate string = "gate"
)

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
			Node:   *ninfo,
			Meta:   registryMeta{},
		},
		clientGateMap:  make(map[uint32]*clientGateItem),
		gateConnMap:    make(map[string]network.Conn),
		gateHoldUserid: make(map[string]map[uint32]*clientGateItem),
		httpMux:        http.NewServeMux(),
	}
	err := surf.init()
	if err != nil {
		return nil, err
	}
	return surf, nil
}

type clientGateItem struct {
	connId   string
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
	httpMux *http.ServeMux

	caller *PacketRouteCaller

	queue  chan func()
	closed chan struct{}

	mux sync.Mutex

	regData nodeRegistryData

	clientGateMap  map[uint32]*clientGateItem
	gateConnMap    map[string]network.Conn
	gateHoldUserid map[string]map[uint32]*clientGateItem

	nodeGroup *NodeGroup
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

func (s *Surf) UpdateNodeData(state NodeState, data json.RawMessage) error {
	if s.registry == nil {
		return fmt.Errorf("registry not init")
	}

	s.regData.Data = data
	s.regData.Status = state

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
	conn := s.ClientGateConn(uid)
	if conn == nil {
		return fmt.Errorf("conn not found")
	}
	return s.SendRequest(conn, NodeType_Client, uid, msg, cb)
}

func (s *Surf) SendAsyncToClient(uid uint32, msg proto.Message) error {
	conn := s.ClientGateConn(uid)
	if conn == nil {
		return fmt.Errorf("conn not found")
	}
	return s.SendAsync(conn, NodeType_Client, uid, msg)
}

func (s *Surf) SendRequestToNode(ntype uint16, nid uint32, msg proto.Message, cb RequestCallbackFunc) error {
	if ntype == 0 && nid == 0 {
		return fmt.Errorf("err target")
	}

	var err error
	if ntype == 0 {
		info := s.nodeGroup.Get(nid)
		if info != nil {
			ntype = info.Node.NType
		}
		if ntype == 0 {
			return fmt.Errorf("err node type")
		}
	}

	if nid == 0 {
		nid, err = s.NextNodeId(ntype)
		if err != nil {
			return err
		}
	}

	conn := s.NodeGateConn()
	if conn == nil {
		return fmt.Errorf("conn not found")
	}
	return s.SendRequest(conn, ntype, nid, msg, cb)
}

func (s *Surf) SendAsyncToNode(ntype uint16, nid uint32, msg proto.Message) error {
	conn := s.ClientGateConn(nid)
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
	msgid := GetMsgId(msg)

	rpk := NewRoutePacket(body)
	rpk.SetMsgId(msgid)
	rpk.SetToUId(uid)
	rpk.SetToURole(urole)
	rpk.SetFromUId(s.NodeID())
	rpk.SetFromURole(s.NodeType())
	rpk.SetMsgType(RoutePackMsgType_Async)
	rpk.SetMarshalType(s.serverInfo.Marshaler.Id())
	err = conn.Send(rpk.ToHVPacket())
	log.Debug("SendAsync", "from", rpk.FromUId(), "fromrole", rpk.FromURole(), "to", rpk.ToUId(),
		"torole", rpk.ToURole(), "msgid", rpk.MsgId(), "msgtype", rpk.MsgType(), "err", err)
	return err
}

func (s *Surf) SendRequest(conn network.Conn, urole uint16, uid uint32, msg proto.Message, cb RequestCallbackFunc) error {
	syn := conn.NextSYN()
	body, err := s.serverInfo.Marshaler.Marshal(msg)
	if err != nil {
		return err
	}
	msgid := GetMsgId(msg)

	rpk := NewRoutePacket(body)
	rpk.SetMsgId(msgid)

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
	log.Debug("SendRequest", "from", rpk.FromUId(), "fromrole", rpk.FromURole(), "to", rpk.ToUId(),
		"torole", rpk.ToURole(), "msgid", rpk.MsgId(), "msgtype", rpk.MsgType(), "err", err)
	return err
}

func (s *Surf) ClientGateConn(uid uint32) network.Conn {
	item, has := s.clientGateMap[uid]
	if !has {
		return nil
	}
	return s.gateConnMap[item.connId]
}

func (s *Surf) NodeGateConn() network.Conn {
	for _, v := range s.gateConnMap {
		return v
	}
	return nil
}

func (s *Surf) HandleFuncs(msgid uint32, chain ConnHandler) {
	if msgid == 0 {
		return
	}
	s.caller.handlers.Add(msgid, chain)
}

func (s *Surf) HttpMux() *http.ServeMux {
	return s.httpMux
}

func (s *Surf) NextNodeId(ntype uint16) (uint32, error) {
	node := s.nodeGroup.Choice(ntype)
	if node == nil {
		return 0, fmt.Errorf("there's no enable node online")
	}
	return node.Node.NId, nil
}
