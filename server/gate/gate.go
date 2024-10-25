package gate

import (
	"crypto/rsa"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
)

type Gate struct {
	Calltable *calltable.CallTable[uint32]
	Marshaler marshal.Marshaler

	ClientConn *ConnStore
	NodeConn   *ConnStore

	Selfinfo *auth.UserInfo

	ClientPublicKey *rsa.PublicKey
	NodePublicKey   *rsa.PublicKey
}

func (gate *Gate) OnConnAuth(data []byte) (auth.User, error) {
	return auth.VerifyToken(gate.NodePublicKey, data)
}

func (gate *Gate) OnConnEnable(conn network.Conn, enable bool) {
	if enable {
		log.Debugf("OnConnEnable: id:%v,addr:%v,uid:%v,urid:%v,enable:%v", conn.ConnID(), conn.RemoteAddr(), conn.UserID(), conn.UserRole(), enable)
		currConn, got := gate.ClientConn.SwapByUID(conn)
		if got {
			currConn := currConn.(network.Conn)
			currConn.Close()
		}
		gate.onUserOnline(conn)
	} else {
		currConn, got := gate.ClientConn.Delete(conn.ConnID())
		if got {
			gate.onUserOffline(currConn)
		}
	}
}

func (gate *Gate) OnConnPacket(s network.Conn, pk *network.HVPacket) {
	log.Infof("OnConnPacket sid:%v,uid:%v,type:%v,datalen:%v", s.ConnID(), s.UserID(), pk.Meta.GetType(), len(pk.Body))

	if pk.Meta.GetType() != network.PacketType_Route || len(pk.Body) < network.RoutePackHeadLen {
		return
	}

	rpk := network.ParseRoutePacket(pk)
	rpk.SetClientId(s.UserID())
	svrtype := rpk.GetSvrType()

	if svrtype == 0 {
		gate.OnCall(s, pk.Meta.GetSubFlag(), rpk)
		return
	}

	nodeid := rpk.GetNodeId()
	servertype := rpk.GetSvrType()
	if nodeid == 0 {
		// get nodeid by servertype
		nodeid = gate.FindIdleNode(servertype)
	}

	v, found := gate.NodeConn.LoadByUID(nodeid)

	if !found {
		rpk.SetErrCode(network.RoutePackType_SubFlag_RouteErrCode_NodeNotFound)
		pk.Meta.SetSubFlag(network.RoutePackType_SubFlag_RouteErr)
		pk.SetHead(rpk.GetHead())
		gate.SendTo(s, pk)
		return
	}

	gate.SendTo(v, pk)
}

func (gate *Gate) FindIdleNode(servertype uint16) uint32 {
	// TODO:
	retappid := uint32(0)
	gate.NodeConn.Range(func(conn network.Conn) bool {
		if uint16(conn.UserRole()) == servertype {
			retappid = conn.UserID()
			return false
		}
		return true
	})
	return retappid
}

// func (r *Router) GetUserSession(uid uint32) network.Conn {
// 	r.userSessionLock.RLock()
// 	defer r.userSessionLock.RUnlock()
// 	return r.userSession[uid]
// }

// func (r *Router) StoreUserSession(uid uint32, s network.Conn) {
// 	r.userSessionLock.Lock()
// 	defer r.userSessionLock.Unlock()
// 	r.userSession[uid] = s
// }

//	func (r *Router) RemoveUserSession(uid uint32) {
//		r.userSessionLock.Lock()
//		defer r.userSessionLock.Unlock()
//		delete(r.userSession, uid)
//	}

func (gate *Gate) OnNodeAuth(data []byte) (auth.User, error) {
	return auth.VerifyToken(gate.NodePublicKey, data)
}

func (gate *Gate) OnNodeStatus(conn network.Conn, enable bool) {
	if enable {
		_, loaded := gate.NodeConn.LoadOrStoreByUID(conn)
		if loaded {
			// 重复连接, 则直接关闭
			conn.Close()
			log.Error("repeat server conn ", conn.ConnID(), conn.UserID())
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
		rpk := network.RoutePacketHead(pk.GetBody())
		clientUID := rpk.GetClientId()
		if clientUID == 0 {
			log.Warnf("client not found clientUID:%d", clientUID)
			return
		}

		v, found := gate.ClientConn.LoadByUID(clientUID)
		if !found {
			log.Warnf("client not found cid:%d", clientUID)
			pk = rpk.GenHVPacket(network.RoutePackType_SubFlag_RouteErr)
			gate.SendTo(s, pk)
			return
		}
		gate.SendTo(v, pk)
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
		"sid":    s.ConnID(),
		"uid":    s.UserID(),
		"enable": true,
	})
}

func (gate *Gate) onServerOffline(s network.Conn) {
	gate.PublishEvent("NodeOffline", map[string]any{
		"sid":    s.ConnID(),
		"uid":    s.UserID(),
		"enable": true,
	})
}

func (gate *Gate) onUserOnline(s network.Conn) {
	ud := NewConnUserData()
	s.SetUserData(ud)
	gate.PublishEvent("ConnEnable", map[string]any{
		"sid":    s.ConnID(),
		"uid":    s.UserID(),
		"enable": true,
	})
}

func (gate *Gate) onUserOffline(s network.Conn) {
	gate.PublishEvent("ConnEnable", map[string]any{
		"sid":    s.ConnID(),
		"uid":    s.UserID(),
		"enable": false,
	})
	s.SetUserData(nil)
}

func (gate *Gate) PublishEvent(ename string, event any) {
	log.Infof("[Mock PublishEvent] name:%v,event:%v", ename, event)
}

func (gate *Gate) OnCall(c network.Conn, subflag uint8, pk *network.RoutePacket) {
	var err error
	switch subflag {
	case network.RoutePackType_SubFlag_Request:
		method := gate.Calltable.Get(pk.GetMsgId())
		if method == nil {
			return
		}
		req := method.NewRequest()
		err = gate.Marshaler.Unmarshal(pk.GetBody(), req)
		if err != nil {
			return
		}
		// method.Call(r, ctx, req)
	default:
	}

	// if msgtype == network.RouteMsgType_Response {

	// } else {
	// 	method := r.ct.Get(msgid)
	// 	req := method.NewRequest()
	// 	err = r.marshaler.Unmarshal(pk.GetBody(), req)
	// 	if err != nil {

	// 		pk.SetMsgType(network.RouteMsgType_RouteErr)
	// 		pk.SetErrCode(network.RouteMsgErrCode_NodeNotFound)

	// 		// c.Send(pk)
	// 		// c.Send()
	// 	}
	// }

	// askid := head.GetAskID()
	// method := r.ct.Get(msgid)
	// if method == nil {
	// 	log.Print("not found method,msgid:", msgid)
	// 	dealSocketErrCnt(s)
	// 	return
	// }

	// reqRaw := method.NewRequest()
	// if reqRaw == nil {
	// 	log.Print("not found request,msgid:", msgid)
	// 	return
	// }

	// req := reqRaw.(proto.Message)
	// err = proto.Unmarshal(body, req)

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// ctx := context.WithValue(context.Background(), tcpSocketKey, s)
	// ctx = context.WithValue(ctx, tcpPacketKey, p)

	// result := method.Call(r, ctx, req)

	// if len(result) != 2 {
	// 	return
	// }
	// // if err is not nil, only return err
	// resperrI := result[1].Interface()
	// if resperrI != nil {
	// 	var senderr error
	// 	switch resperr := resperrI.(type) {
	// 	case *msg.Error:
	// 		senderr = r.SendMessage(s, askid, RouteTypRespErr, resperr)
	// 	case error:
	// 		senderr = r.SendMessage(s, askid, RouteTypRespErr, &msg.Error{
	// 			Code:   -1,
	// 			Detail: resperr.Error(),
	// 		})
	// 	default:
	// 		log.Print("not support error type:")
	// 	}
	// 	if senderr != nil {
	// 		log.Print("send err failed:", senderr)
	// 	}
	// 	return
	// }

	// respI := result[0].Interface()
	// if respI != nil {
	// 	resp, ok := respI.(proto.Message)
	// 	if !ok {
	// 		return
	// 	}
	// 	respMsgTyp := head.GetMsgTyp()
	// 	if respMsgTyp == RouteTypRequest {
	// 		respMsgTyp = RouteTypResponse
	// 	}

	// 	r.SendMessage(s, askid, respMsgTyp, resp)
	// 	log.Printf("oncall sid:%v,uid:%v,msgid:%v,askid:%v,req:%v,resp:%v\n", s.ID(), s.UID(), msgid, askid, req, resp)
	// 	return
	// }
}
