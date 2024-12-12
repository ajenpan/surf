package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	etcclientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/registry"
	utilRsa "github.com/ajenpan/surf/core/utils/rsagen"
	msgCore "github.com/ajenpan/surf/msg/core"
)

func (s *Surf) init() error {
	pubkey, err := utilRsa.LoadRsaPublicKeyFromUrl(s.conf.PublicKeyFilePath)
	if err != nil {
		return err
	}
	s.rasPublicKey = pubkey

	s.caller = NewPacketRouteCaller()

	if s.NodeType() != NodeType_Gate {
		HandleFunc(s, s.onNotifyClientConnect)
		HandleFunc(s, s.onNotifyClientDisconnect)
	}

	err = s.serverInfo.Svr.OnInit(s)
	if err != nil {
		return err
	}

	return nil
}
func (s *Surf) getMsgIdFromPath(path string) uint32 {
	// TODO: name to id map
	idx := strings.LastIndexByte(path, '/') + 1
	if idx <= 0 || idx >= len(path) {
		return 0
	}
	msgIdStr := path[idx:]
	msgid, _ := strconv.Atoi(msgIdStr)
	return uint32(msgid)
}

func (s *Surf) startHttpSvr() {
	log.Info("start http server", "addr", s.conf.HttpListenAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Print("onrecv ", r.URL.Path)
		msgid := s.getMsgIdFromPath(r.URL.Path)
		if msgid <= 0 {
			http.Error(w, "handle not found", http.StatusNotFound)
			return
		}
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

		msgType := RoutePackMsgType_Request
		msgTypeStr := r.Header.Get("MsgType")
		if len(msgTypeStr) > 0 {
			msgType, _ = strconv.Atoi(msgTypeStr)
		}

		rpk := NewRoutePacket(raw)
		rpk.SetFromUId(uinfo.UserID())
		rpk.SetFromURole(uinfo.UserRole())
		rpk.SetMarshalType(marshal.NameToId(r.Header.Get("Content-Type")))
		rpk.SetToUId(s.NodeID())
		rpk.SetToURole(s.NodeType())
		rpk.SetMsgId(msgid)
		rpk.SetMsgType(uint8(msgType))

		var ctx = &HttpContext{
			W:         w,
			R:         r,
			UInfo:     uinfo,
			ReqPacket: rpk,
			RespChan:  make(chan func()),
		}

		if msgType == RoutePackMsgType_Request {
			s.Do(func() {
				s.caller.Call(ctx)
			})
		} else {
			s.Do(func() {
				s.caller.Call(ctx)
				ctx.RespChan <- func() {
					ctx.W.WriteHeader(http.StatusOK)
				}
			})
		}
		if f := <-ctx.RespChan; f != nil {
			f()
		}
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

func (h *Surf) onGatePacket(conn network.Conn, pk *network.HVPacket) {
	switch pk.Meta.GetType() {
	case network.PacketType_Route:
		rpk := NewRoutePacket(nil).FromHVPacket(pk)
		if rpk == nil {
			log.Error("parse route pakcet error")
			return
		}
		ctx := &ConnContext{
			ReqConn:   conn,
			Core:      h,
			ReqPacket: rpk,
		}

		h.Do(func() {
			h.caller.Call(ctx)
		})
	default:
		log.Error("invalid packet type", "type", pk.Meta.GetType())
	}
}

func (s *Surf) catch() {
	if err := recover(); err != nil {
		log.Error("catch panic", "err", err)
	}
}

func (s *Surf) onConnAuth(data []byte) (network.User, error) {
	return auth.VerifyToken(s.rasPublicKey, data)
}

func (s *Surf) onGateStatus(conn network.Conn, enable bool) {
	log.Info("conn status", "id", conn.ConnId(), "uid", conn.UserID(), "utype", conn.UserRole(), "status", enable)
	if conn.UserRole() == NodeType_Client {
		s.Do(func() {
			if enable {
				s.notifyClientConnect(conn.UserID(), s.ninfo.NodeID(), conn.RemoteAddr())
			} else {
				s.notifyClientDisconnect(conn.UserID(), s.ninfo.NodeID(), msgCore.NotifyClientDisconnect_Disconnect)
			}
		})
	} else if conn.UserRole() == NodeType_Gate {
		s.Do(func() {
			if enable {
				s.gateConnMap[conn.ConnId()] = conn
			} else {
				s.onGateDisconn(conn)
				delete(s.gateConnMap, conn.ConnId())
			}
		})
	} else {
		// do nothing
	}
}

func (s *Surf) onNotifyClientConnect(ctx Context, msg *msgCore.NotifyClientConnect) {
	item := &clientGateItem{
		// conn:     ctx.Conn(),
		connId:   ctx.ConnId(),
		clientIp: msg.IpAddr,
		connAt:   time.Now(),
	}
	s.Do(func() {
		s.clientGateMap[msg.Uid] = item
		if ctx.FromURole() == NodeType_Gate {
			m, has := s.gateHoldUserid[ctx.ConnId()]
			if !has {
				m = make(map[uint32]*clientGateItem)
				s.gateHoldUserid[ctx.ConnId()] = m
			}
			m[msg.Uid] = item
		}
		s.notifyClientConnect(msg.Uid, msg.GateNodeId, msg.IpAddr)
	})
}

func (s *Surf) onNotifyClientDisconnect(ctx Context, msg *msgCore.NotifyClientDisconnect) {
	s.Do(func() {
		delete(s.clientGateMap, msg.Uid)
		if m, has := s.gateHoldUserid[ctx.ConnId()]; has {
			delete(m, msg.Uid)
		}
		s.notifyClientDisconnect(msg.Uid, msg.GateNodeId, msg.Reason)
	})
}

func (s *Surf) notifyClientConnect(uid uint32, gateNodeId uint32, ip string) {
	log.Info("recv notify client connect", "uid", uid, "gateNodeId", gateNodeId, "ip", ip)
	if s.serverInfo.OnClientConnect != nil {
		s.serverInfo.OnClientConnect(uid, gateNodeId, ip)
	}
}

func (s *Surf) notifyClientDisconnect(uid uint32, gateNodeId uint32, reason msgCore.NotifyClientDisconnect_Reason) {
	log.Info("recv notify client disconnect", "uid", uid, "gateNodeId", gateNodeId, "reason", reason)
	if s.serverInfo.OnClientDisconnect != nil {
		s.serverInfo.OnClientDisconnect(uid, gateNodeId, int32(reason))
	}
}

func (s *Surf) onGateDisconn(gateConn network.Conn) {
	m, has := s.gateHoldUserid[gateConn.ConnId()]
	if !has {
		return
	}

	for uid := range m {
		s.serverInfo.OnClientDisconnect(uid, gateConn.UserID(), int32(msgCore.NotifyClientDisconnect_GateClosed))

		delete(s.clientGateMap, uid)
	}
	delete(s.gateHoldUserid, gateConn.ConnId())
}

func (s *Surf) connectGates() {
	if s.NodeType() == NodeType_Gate {
		return
	}

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
}

func (s *Surf) startNodeWatcher() error {
	if s.conf.EtcdConf != nil {
		var err error
		cb := func(ev *etcclientv3.Event) {
			switch ev.Type {
			case etcclientv3.EventTypePut:
				node := &nodeRegistryData{}
				err := json.Unmarshal(ev.Kv.Value, node)
				if err != nil {
					return
				}
				s.nodeGroup.Set(node)
			case etcclientv3.EventTypeDelete:
				nid := 0
				strkey := string(ev.Kv.Key)
				strlist := strings.Split(strkey, "/")
				if len(strlist) == 0 {
					return
				}
				nid, _ = strconv.Atoi(strlist[len(strlist)-1])
				s.nodeGroup.Del(uint32(nid))
			}
		}

		s.nodeWatcher, err = registry.NewEtcdWatcher(*s.conf.EtcdConf, "/reg/node/", cb, etcclientv3.WithPrefix())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Surf) startNodeRegistry() error {
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
	return nil
}
