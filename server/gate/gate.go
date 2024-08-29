package route

import (
	"sync"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/core/utils/marshal"
)

func NewGate() *Gate {
	ret := &Gate{
		// userSession: make(map[uint32]network.Conn),
		// UserSessions :
		marshaler: &marshal.ProtoMarshaler{},
	}
	// ret.ct = calltable.ExtractAsyncMethodByMsgID()
	return ret
}

type Gate struct {
	marshaler marshal.Marshaler

	ct *calltable.CallTable[uint32]

	ClientConn sync.Map
	ServerConn sync.Map
	Selfinfo   *auth.UserInfo
}

func (gate *Gate) OnConnEnable(conn network.Conn, enable bool) {
	log.Debugf("OnConnEnable: id:%v,addr:%v,uid:%v,urid:%v,enable:%v", conn.ConnID(), conn.RemoteAddr(), conn.UserID(), conn.UserRole(), enable)
	if enable {
		currConn, got := gate.ClientConn.Swap(conn.UserID(), conn)
		if got {
			gate.onUserOffline(currConn.(network.Conn))
		} else {
			gate.onUserOnline(conn)
		}
	} else {
		currConn, got := gate.ClientConn.LoadAndDelete(conn.UserID())
		if got {
			gate.onUserOffline(currConn.(network.Conn))
			currConn.(network.Conn).Close()
		}
	}
}

func (gate *Gate) OnConnPacket(s network.Conn, pk *network.HVPacket) {
	if pk.Meta.GetType() != network.PacketType_Route || len(pk.Body) < network.RoutePackHeadLen {
		return
	}

	// network.par
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

	v, found := gate.ServerConn.Load(nodeid)

	if !found {
		rpk.SetErrCode(network.RoutePackType_SubFlag_RouteErrCode_NodeNotFound)
		pk.Meta.SetSubFlag(network.RoutePackType_SubFlag_RouteErr)
		pk.SetHead(rpk.GetHead())
		s.Send(pk)
		return
	}

	v.(network.Conn).Send(pk)
}

func (gate *Gate) FindIdleNode(servertype uint16) uint32 {
	// todo
	return 0
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

func (gate *Gate) OnNodeEnable(conn network.Conn, enable bool) {
	if enable {
		currConn, got := gate.ServerConn.Swap(conn.UserID(), conn)
		if got {
			gate.onServerOffline(currConn.(network.Conn))
		} else {
			gate.onServerOnline(conn)
		}
	} else {
		currConn, got := gate.ServerConn.LoadAndDelete(conn.UserID())
		if got {
			gate.onServerOffline(currConn.(network.Conn))
			currConn.(network.Conn).Close()
		}
	}
}

func (gate *Gate) OnNodePacket(s network.Conn, pk *network.HVPacket) {
	switch pk.Meta.GetType() {
	case network.PacketType_Route:
		rpk := network.RoutePacketHead(pk.GetBody())
		cid := rpk.GetClientId()
		if cid == 0 {
			log.Warnf("client not found cid:%d", cid)
			return
		}

		v, found := gate.ClientConn.Load(cid)
		if !found {
			log.Warnf("client not found cid:%d", cid)
			// pk = rpk.GenHVPacket(network.RouteMsgType_SubFlag_RouteErr)
			// s.Send(pk)
			return
		}
		v.(network.Conn).Send(pk)
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
	log.Printf("[Mock PublishEvent] name:%v,event:%v", ename, event)
}

func (gate *Gate) OnCall(c network.Conn, subflag uint8, pk *network.RoutePacket) {
	var err error
	switch subflag {
	case network.RoutePackType_SubFlag_Request:
		method := gate.ct.Get(pk.GetMsgId())
		if method == nil {
			return
		}
		req := method.NewRequest()
		err = gate.marshaler.Unmarshal(pk.GetBody(), req)
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
