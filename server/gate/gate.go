package gate

import (
	"crypto/rsa"
	"encoding/json"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
	msgCore "github.com/ajenpan/surf/msg/core"
)

func New() *Gate {
	return &Gate{}
}

type Gate struct {
	surf      *core.Surf
	conf      *Config
	calltable *calltable.CallTable

	clientConnStore *core.ClientConnStore
	nodeConnStore   *core.NodeConnStore
	caller          *core.PacketRouteCaller

	ClientPublicKey *rsa.PublicKey
	NodePublicKey   *rsa.PublicKey

	closer func() error
}

func (gate *Gate) OnInit(surf *core.Surf) error {
	var err error

	gate.conf = DefaultConfig
	confStr := surf.ServerConf()
	if len(confStr) > 2 {
		err = json.Unmarshal(confStr, gate.conf)
		if err != nil {
			return err
		}
	}

	gate.closer, err = Start(gate, gate.conf)
	if err != nil {
		return err
	}
	gate.surf = surf
	return err
}

func (gate *Gate) OnReady() {
	gate.surf.UpdateNodeData(core.NodeState_Running, nil)
}

func (gate *Gate) OnStop() error {
	return nil
}

func (gate *Gate) OnConnAuth(data []byte) (network.User, error) {
	return auth.VerifyToken(gate.NodePublicKey, data)
}

func (gate *Gate) OnConnEnable(conn network.Conn, enable bool) {
	if enable {
		gate.onUserOnline(conn)
	} else {
		gate.onUserOffline(conn)
	}
}

func (gate *Gate) OnConnPacket(s network.Conn, pk *network.HVPacket) {
	log.Info("OnConnPacket", "sid", s.ConnId(), "uid", s.UserID(), "type", pk.Meta.GetType(), "datalen", len(pk.Body))

	if pk.Meta.GetType() != network.PacketType_Route || len(pk.Body) < core.RoutePackHeadBytesLen {
		return
	}

	rpk := core.NewRoutePacket(nil).FromHVPacket(pk)
	if rpk == nil {
		log.Warn("OnConnPacket parse route packet failed", "pk", pk)
		return
	}

	target_ntype := rpk.GetToURole()
	target_nid := rpk.GetToUId()

	log.Debug("recv client route pk", "from", rpk.GetFromUId(), "fromrole", rpk.GetFromURole(), "to", target_nid, "torole", target_ntype, "msgid", rpk.GetMsgId(), "msgtype", rpk.GetMsgType())

	if target_ntype == 0 && target_nid == 0 {
		log.Warn("OnConnPacket from server type and nodeid is 0")
		return
	}

	if target_nid == gate.surf.NodeID() || target_ntype == gate.surf.NodeType() {
		gate.OnCall(s, rpk)
		return
	}

	ud := s.GetUserData().(*ConnUserData)
	var targetConn network.Conn
	if target_nid != 0 {
		targetConn, _ = gate.nodeConnStore.LoadByUId(target_nid)
	} else {
		// get nodeid by servertype
		nodeid, has := ud.serverType2NodeID[target_ntype]
		if has {
			targetConn, _ = gate.nodeConnStore.LoadByUId(nodeid)
		}
		if !has || targetConn == nil {
			targetConn = gate.FindIdleNode(target_ntype)
		}
	}

	if targetConn == nil {
		pk.Meta.SetSubFlag(core.RoutePackType_SubFlag_RouteFail)
		gate.SendTo(s, pk)
		return
	}

	if _, has := ud.nodeids[targetConn.UserID()]; !has {
		ud.nodeids[targetConn.UserID()] = struct{}{}
		ud.serverType2NodeID[target_ntype] = targetConn.UserID()

		notify := &msgCore.NotifyClientConnect{
			Uid:        s.UserID(),
			GateNodeId: gate.surf.NodeID(),
			IpAddr:     s.RemoteAddr(),
		}
		gate.surf.SendAsync(targetConn, core.NodeType_Core, targetConn.UserID(), notify)
	}

	rpk.SetFromUId(s.UserID())
	rpk.SetFromURole(s.UserRole())
	rpk.SetToUId(targetConn.UserID())
	rpk.SetToURole(targetConn.UserRole())
	gate.SendTo(targetConn, pk)
}

func (gate *Gate) FindIdleNode(servertype uint16) network.Conn {
	// TODO: how to find the best node
	var ret network.Conn
	gate.nodeConnStore.Range(func(conn network.Conn) bool {
		if uint16(conn.UserRole()) == servertype {
			ret = conn
			return false
		}
		return true
	})
	return ret
}

func (gate *Gate) OnNodeAuth(data []byte) (network.User, error) {
	info := &auth.NodeInfo{}
	err := info.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (gate *Gate) OnNodeStatus(conn network.Conn, enable bool) {
	log.Info("OnNodeStatus", "connid", conn.ConnId(), "nodeid", conn.UserID(), "svrtype", conn.UserRole(), "enable", enable)
}

func (gate *Gate) OnNodePacket(s network.Conn, pk *network.HVPacket) {
	switch pk.Meta.GetType() {
	case network.PacketType_Route:
		rpk := core.NewRoutePacket(nil).FromHVPacket(pk)
		if rpk == nil {
			log.Warn("OnNodePacket parse route packet failed", "type", pk.Meta.GetType())
			return
		}

		log.Debug("OnNodePacket", "from", rpk.GetFromUId(), "fromrole", rpk.GetFromURole(),
			"to", rpk.GetToUId(), "torole", rpk.GetToURole(), "msgid", rpk.GetMsgId(), "msgtype", rpk.GetMsgType())

		clientUID := rpk.GetToUId()
		if clientUID == 0 {
			log.Warn("client not found", "clientUID", clientUID)
			return
		}

		v, found := gate.clientConnStore.LoadByUId(clientUID)
		if !found {
			log.Warn("client not found", "cid", clientUID)
			pk.Meta.SetSubFlag(core.RoutePackType_SubFlag_RouteFail)
			gate.SendTo(s, pk)
			return
		}
		gate.SendTo(v, pk)
	default:
		log.Warn("OnNodePacket unknown packet type", "type", pk.Meta.GetType())
	}
}

func (gate *Gate) SendTo(c network.Conn, pk *network.HVPacket) {
	err := c.Send(pk)
	if err != nil {
		log.Error("send packet failed", "err", err)
	}
}

func (gate *Gate) onServerOnline(s network.Conn) {
	gate.PublishEvent("NodeOnline", map[string]any{
		"conn_id":  s.ConnId(),
		"node_id":  s.UserID(),
		"svr_type": s.UserRole(),
		"enable":   true,
	})
}

func (gate *Gate) onServerOffline(s network.Conn) {
	gate.PublishEvent("NodeOffline", map[string]any{
		"conn_id":   s.ConnId(),
		"node_id":   s.UserID(),
		"node_type": s.UserRole(),
		"enable":    false,
	})
}

func (gate *Gate) onUserOnline(conn network.Conn) {
	conn.SetUserData(NewConnUserData(conn))

	gate.PublishEvent("UserOnline", map[string]any{
		"sid":          conn.ConnId(),
		"uid":          conn.UserID(),
		"gate_node_id": gate.nodeId(),
		"enable":       true,
	})
}

func (gate *Gate) nodeId() uint32 {
	return gate.surf.NodeID()
}

func (gate *Gate) onUserOffline(conn network.Conn) {
	gate.PublishEvent("UserOffline", map[string]any{
		"sid":    conn.ConnId(),
		"uid":    conn.UserID(),
		"enable": false,
	})

	ud := conn.GetUserData().(*ConnUserData)
	conn.SetUserData(nil)

	notify := &msgCore.NotifyClientDisconnect{
		Uid:        conn.UserID(),
		GateNodeId: gate.surf.NodeID(),
		Reason:     msgCore.NotifyClientDisconnect_Disconnect,
	}

	for nodeid := range ud.nodeids {
		node, _ := gate.nodeConnStore.LoadByUId(nodeid)
		if node == nil || !node.Enable() {
			continue
		}
		gate.surf.SendAsync(node, core.NodeType_Core, node.UserID(), notify)
	}
}

func (gate *Gate) PublishEvent(ename string, event any) {
	log.Info("PublishEvent", "ename", ename, "event", event)
}

func (gate *Gate) OnCall(conn network.Conn, pk *core.RoutePacket) {
	ctx := &core.ConnContext{
		ReqConn:   conn,
		Core:      gate.surf,
		ReqPacket: pk,
	}
	gate.caller.Call(ctx)
}
