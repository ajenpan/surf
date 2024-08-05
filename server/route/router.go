package route

import (
	"sync"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/marshal"
)

func NewRouter() *Router {
	ret := &Router{
		// userSession: make(map[uint32]network.Conn),
		// UserSessions :
		marshaler: &marshal.ProtoMarshaler{},
	}
	// ret.ct = calltable.ExtractAsyncMethodByMsgID()
	return ret
}

type Router struct {
	marshaler marshal.Marshaler

	//ct *calltable.CallTable[uint32]

	ClientConn sync.Map
	ServerConn sync.Map
	Selfinfo   *auth.UserInfo
}

func (r *Router) OnConnEnable(conn network.Conn, enable bool) {
	log.Debugf("OnConnEnable: id:%v,addr:%v,uid:%v,urid:%v,enable:%v", conn.ConnID(), conn.RemoteAddr(), conn.UserID(), conn.UserRole(), enable)
	if enable {
		currConn, got := r.ClientConn.Swap(conn.UserID(), conn)
		if got {
			r.onUserOffline(currConn.(network.Conn))
		}
		r.onUserOnline(conn)
	} else {
		currConn, got := r.ClientConn.LoadAndDelete(conn.UserID())
		if got {
			r.onUserOffline(currConn.(network.Conn))
			currConn.(network.Conn).Close()
		}
	}
}

func (r *Router) OnConnPacket(s network.Conn, pk *network.HVPacket) {

	pk.Head.SetClientId(s.UserID())
	svrtype := pk.Head.GetSvrType()

	if svrtype == 0 {
		//TODO:
		return
	}

	svrid := pk.Head.GetSvrId()
	if svrid == 0 {
		// TODO:
		return
	}

	v, found := r.ServerConn.Load(svrid)

	if !found {
		pk.Head.SetType(network.PacketType_RouteErr)
		pk.Head.SetSubFlag(network.PacketType_RouteErr_NodeNotFound)
		s.Send(pk)
		return
	}

	v.(network.Conn).Send(pk)
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

func (r *Router) OnNodeEnable(conn network.Conn, enable bool) {
	if enable {
		currConn, got := r.ServerConn.Swap(conn.UserID(), conn)
		if got {
			r.onServerOffline(currConn.(network.Conn))
		}
		r.onServerOnline(conn)
	} else {
		currConn, got := r.ServerConn.LoadAndDelete(conn.UserID())
		if got {
			r.onServerOffline(currConn.(network.Conn))
			currConn.(network.Conn).Close()
		}
	}
}

func (r *Router) OnNodePacket(s network.Conn, pk *network.HVPacket) {
	cid := pk.Head.GetClientId()
	if cid == 0 {
		// todo handler call
		return
	}

	v, found := r.ClientConn.Load(cid)
	if !found {
		pk.Head.SetType(network.PacketType_RouteErr)
		pk.Head.SetSubFlag(network.PacketType_RouteErr_NodeNotFound)
		s.Send(pk)
		return
	}
	v.(network.Conn).Send(pk)
}

func (r *Router) onServerOnline(s network.Conn) {

}

func (r *Router) onServerOffline(s network.Conn) {

}

func (r *Router) onUserOnline(s network.Conn) {
	ud := NewConnUserData()
	s.SetUserData(ud)

	r.PublishEvent("ConnEnable", map[string]any{
		"sid":    s.ConnID(),
		"uid":    s.UserID(),
		"enable": true,
	})
}

func (r *Router) onUserOffline(s network.Conn) {
	r.PublishEvent("ConnEnable", map[string]any{
		"sid":    s.ConnID(),
		"uid":    s.UserID(),
		"enable": false,
	})
	s.SetUserData(nil)
}

func (r *Router) PublishEvent(ename string, event any) {
	log.Printf("[Mock PublishEvent] name:%v,event:%v", ename, event)
}

func (r *Router) OnCall(c network.Conn, pk *network.HVPacket) {
	// var err error

	// msgid := pk.GetMsgId()
	// msgtype := pk.GetMsgType()

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
