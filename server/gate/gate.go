package gate

import (
	"crypto/rsa"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
	msgCore "github.com/ajenpan/surf/msg/core"
)

type Gate struct {
	Calltable *calltable.CallTable
	Marshaler marshal.Marshaler

	NodeID uint32

	ClientPublicKey *rsa.PublicKey
	NodePublicKey   *rsa.PublicKey

	clientConnStore *core.ClientConnStore
	nodeConnStore   *core.NodeConnStore

	caller *core.PacketRouteCaller
}

func (gate *Gate) ServerType() uint16 {
	return core.ServerType_Gate
}

func (gate *Gate) ServerName() string {
	return "gate"
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

	target_svrtype := rpk.GetToURole()
	target_nodeid := rpk.GetToUID()

	log.Debug("recv route pk", "from", rpk.GetFromUID(), "fromrole", rpk.GetFromURole(), "to", target_nodeid, "torole", target_svrtype, "msgid", rpk.GetMsgId(), "msgtype", rpk.GetMsgType())

	if target_svrtype == 0 && target_nodeid == 0 {
		log.Warn("OnConnPacket from server type and nodeid is 0")
		return
	}

	if target_nodeid == gate.NodeID || target_svrtype == gate.ServerType() {
		gate.OnCall(s, rpk)
		return
	}

	ud := s.GetUserData().(*ConnUserData)

	var targetNode network.Conn

	if target_nodeid != 0 {
		targetNode, _ = gate.nodeConnStore.LoadByUId(target_nodeid)
	} else {
		// get nodeid by servertype
		nodeid, has := ud.serverType2NodeID[target_svrtype]
		if has {
			targetNode, _ = gate.nodeConnStore.LoadByUId(nodeid)
		}
		if !has || targetNode == nil {
			targetNode = gate.FindIdleNode(target_svrtype)
			ud.serverType2NodeID[target_svrtype] = targetNode.UserID()
		}
	}

	if targetNode == nil {
		pk.Meta.SetSubFlag(core.RoutePackType_SubFlag_RouteFail)
		gate.SendTo(s, pk)
		return
	}

	if _, has := ud.nodeids[targetNode.UserID()]; !has {
		ud.nodeids[targetNode.UserID()] = struct{}{}
	}

	rpk.SetFromUID(s.UserID())
	rpk.SetFromURole(s.UserRole())
	rpk.SetToUID(targetNode.UserID())
	rpk.SetToURole(targetNode.UserRole())

	gate.SendTo(targetNode, pk)
}

func (gate *Gate) FindIdleNode(servertype uint16) network.Conn {
	// TODO: the battle node
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

		log.Debug("OnNodePacket", "from", rpk.GetFromUID(), "fromrole", rpk.GetFromURole(),
			"to", rpk.GetToUID(), "torole", rpk.GetToURole(), "msgid", rpk.GetMsgId(), "msgtype", rpk.GetMsgType())

		clientUID := rpk.GetToUID()
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
		"conn_id":  s.ConnId(),
		"node_id":  s.UserID(),
		"svr_type": s.UserRole(),
		"enable":   false,
	})
}

func (gate *Gate) onUserOnline(conn network.Conn) {
	conn.SetUserData(NewConnUserData(conn))

	gate.PublishEvent("UserOnline", map[string]any{
		"sid":      conn.ConnId(),
		"uid":      conn.UserID(),
		"node_id":  0,
		"svr_type": gate.ServerType(),
		"enable":   true,
	})
}

func (gate *Gate) onUserOffline(conn network.Conn) {
	gate.PublishEvent("UserOffline", map[string]any{
		"sid":    conn.ConnId(),
		"uid":    conn.UserID(),
		"enable": false,
	})
	notify := &msgCore.NotifyClientDisconnect{
		Uid:        conn.UserID(),
		GateNodeId: gate.NodeID,
		Reason:     msgCore.NotifyClientDisconnect_Disconnect,
	}
	ud := conn.GetUserData().(*ConnUserData)
	conn.SetUserData(nil)

	marshaler := &marshal.ProtoMarshaler{}
	body, err := marshaler.Marshal(notify)
	if err != nil {
		log.Error("marshal notify failed", "err", err)
		return
	}

	msgid := calltable.GetMessageMsgID(notify.ProtoReflect().Descriptor())

	for nodeid := range ud.nodeids {
		node, _ := gate.nodeConnStore.LoadByUId(nodeid)
		if node == nil {
			continue
		}

		npk := core.NewRoutePacket(body)
		npk.SetMsgId(msgid)
		npk.SetFromUID(gate.NodeID)
		npk.SetFromURole(gate.ServerType())

		npk.SetToUID(nodeid)
		// this msg is from gate to core
		npk.SetToURole(core.ServerType_Core)

		gate.SendTo(node, npk.ToHVPacket())
	}
}

func (gate *Gate) PublishEvent(ename string, event any) {
	log.Info("PublishEvent", "ename", ename, "event", event)
}

func (gate *Gate) OnCall(c network.Conn, pk *core.RoutePacket) {
	var err error
	switch pk.GetMsgType() {
	case core.RoutePackMsgType_Async:
		fallthrough
	case core.RoutePackMsgType_Request:
		method := gate.Calltable.GetByID(pk.GetMsgId())
		if method == nil {
			return
		}
		req := method.NewRequest()
		marshaler := marshal.NewMarshalerById((uint8)(pk.GetMarshalType()))
		if marshaler == nil {
			return
		}
		err = marshaler.Unmarshal(pk.GetBody(), req)
		if err != nil {
			return
		}
		ctx := &gateContext{
			Conn:      c,
			ReqPacket: pk,
			caller:    c.UserID(),
			Marshal:   marshaler,
		}
		method.Call(gate, ctx, req)
	default:
	}
}

type gateContext struct {
	Conn      network.Conn
	ReqPacket *core.RoutePacket
	caller    uint32
	Marshal   marshal.Marshaler
}

func (ctx gateContext) Response(msg proto.Message, err error) {

}

func (ctx gateContext) Caller() uint32 {
	return ctx.caller
}
