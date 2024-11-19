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

	NodeConn *core.ConnStore

	NodeID uint32

	ClientPublicKey *rsa.PublicKey
	NodePublicKey   *rsa.PublicKey

	ClientConn *core.ClientGate

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
	log.Infof("OnConnPacket sid:%v,uid:%v,type:%v,datalen:%v", s.ConnID(), s.UserID(), pk.Meta.GetType(), len(pk.Body))

	if pk.Meta.GetType() != network.PacketType_Route || len(pk.Body) < core.RoutePackHeadLen {
		return
	}

	rpk := core.NewRoutePacket(nil).FromHVPacket(pk)
	if rpk == nil {
		log.Warnf("OnConnPacket parse route packet failed: %v", pk)
		return
	}

	target_svrtype := rpk.GetToURole()
	target_nodeid := rpk.GetToUID()

	if target_svrtype == 0 && target_nodeid == 0 {
		log.Warnf("OnConnPacket from server type and nodeid is 0")
		return
	}

	if target_nodeid == gate.NodeID || target_svrtype == gate.ServerType() {
		gate.OnCall(s, rpk)
		return
	}

	ud := s.GetUserData().(*ConnUserData)

	var targetNode network.Conn

	if target_nodeid != 0 {
		targetNode, _ = gate.NodeConn.LoadByUID(target_nodeid)
	} else {
		// get nodeid by servertype
		nodeid, has := ud.serverType2NodeID[target_svrtype]
		if has {
			targetNode, _ = gate.NodeConn.LoadByUID(nodeid)
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
	rpk.SetFromURole(core.ServerType_User)
	gate.SendTo(targetNode, pk)
}

func (gate *Gate) FindIdleNode(servertype uint16) network.Conn {
	// TODO:
	var ret network.Conn
	gate.NodeConn.Range(func(conn network.Conn) bool {
		if uint16(conn.UserRole()) == servertype {
			ret = conn
			return false
		}
		return true
	})
	return ret
}

func (gate *Gate) OnNodeAuth(data []byte) (network.User, error) {
	return auth.VerifyToken(gate.NodePublicKey, data)
}

func (gate *Gate) OnNodeStatus(conn network.Conn, enable bool) {
	if enable {
		_, loaded := gate.NodeConn.LoadOrStoreByUID(conn)
		if loaded {
			// 重复连接, 则直接关闭
			conn.Close()
			log.Errorf("repeat server conn:%v, nodeid:%d,svrtype:%d", conn.ConnID(), conn.UserID(), conn.UserRole())
		} else {
			gate.onServerOnline(conn)
		}
	} else {
		currConn, got := gate.NodeConn.Delete(conn.ConnID())
		if got {
			gate.onServerOffline(currConn)
		}
	}
}

func (gate *Gate) OnNodePacket(s network.Conn, pk *network.HVPacket) {
	log.Infof("OnNodePacket sid:%v,uid:%v,type:%v,datalen:%v", s.ConnID(), s.UserID(), pk.Meta.GetType(), len(pk.Body))
	switch pk.Meta.GetType() {
	case network.PacketType_Route:
		rpk := core.NewRoutePacket(nil).FromHVPacket(pk)

		clientUID := rpk.GetToUID()
		if clientUID == 0 {
			log.Warnf("client not found clientUID:%d", clientUID)
			return
		}

		v, found := gate.ClientConn.GetConnByUid(clientUID)
		if !found {
			log.Warnf("client not found cid:%d", clientUID)
			pk.Meta.SetSubFlag(core.RoutePackType_SubFlag_RouteFail)
			gate.SendTo(s, pk)
			return
		}
		gate.SendTo(v, pk)
	case network.PacketType_Node:
	}
}

func (gate *Gate) SendTo(c network.Conn, pk *network.HVPacket) {
	err := c.Send(pk)
	if err != nil {
		log.Errorf("send packet failed: %v", err)
	}
}

func (gate *Gate) onServerOnline(s network.Conn) {
	gate.PublishEvent("NodeOnline", map[string]any{
		"conn_id":  s.ConnID(),
		"node_id":  s.UserID(),
		"svr_type": s.UserRole(),
		"enable":   true,
	})
}

func (gate *Gate) onServerOffline(s network.Conn) {
	gate.PublishEvent("NodeOffline", map[string]any{
		"conn_id":  s.ConnID(),
		"node_id":  s.UserID(),
		"svr_type": s.UserRole(),
		"enable":   false,
	})
}

func (gate *Gate) onUserOnline(conn network.Conn) {
	conn.SetUserData(NewConnUserData(conn))

	gate.PublishEvent("UserOnline", map[string]any{
		"sid":      conn.ConnID(),
		"uid":      conn.UserID(),
		"node_id":  0,
		"svr_type": gate.ServerType(),
		"enable":   true,
	})
}

func (gate *Gate) onUserOffline(conn network.Conn) {
	gate.PublishEvent("UserOffline", map[string]any{
		"sid":    conn.ConnID(),
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
		log.Errorf("marshal notify failed: %v", err)
		return
	}

	msgid := calltable.GetMessageMsgID(notify.ProtoReflect().Descriptor())

	for nodeid := range ud.nodeids {
		node, _ := gate.NodeConn.LoadByUID(nodeid)
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
	log.Infof("[Mock PublishEvent] name:%v,event:%v", ename, event)
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
