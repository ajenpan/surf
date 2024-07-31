package route

import (
	"sync"
	"time"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
)

func NewRouter() *Router {
	ret := &Router{
		// userSession: make(map[uint32]network.Conn),
		// UserSessions :
	}
	return ret
}

type Router struct {
	// userSession     map[uint32]network.Conn
	// userSessionLock sync.RWMutex
	// UserSessions *UserSessions

	UserConn   sync.Map
	ServerConn sync.Map

	Selfinfo *auth.UserInfo
}

func (r *Router) OnConnEnable(conn network.Conn, enable bool) {
	log.Debugf("OnConnEnable: id:%v,addr:%v,uid:%v,urid:%v,enable:%v", conn.ConnID(), conn.RemoteAddr(), conn.UserID(), conn.UserRole(), enable)

	switch conn.UserRole() {
	case auth.UserRole_User:
		if enable {
			currConn, got := r.UserConn.Swap(conn.UserID(), conn)
			if got {
				r.onUserOffline(currConn.(network.Conn))
			}
			r.onUserOnline(conn)
		} else {
			currConn, got := r.UserConn.LoadAndDelete(conn.UserID())
			if got {
				r.onUserOffline(currConn.(network.Conn))
				currConn.(network.Conn).Close()
			}
		}
	case auth.UserRole_Server:
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
}

func (r *Router) OnConnPacket(s network.Conn, pk *network.HVPacket) {
	// var err error
	// targetuid := m.GetUid()

	if pk.GetFlag() == 1 {
		return
	}

	// if targetuid == 0 {
	// 	//call my self
	// 	r.OnCall(s, m)
	// 	return
	// }

	// targetSess := r.GetUserSession(targetuid)
	// if targetSess == nil {
	// 	//TODO: send err to source
	// 	log.Warnf("session uid:%v not found", targetuid)
	// 	return
	// }

	// if !r.forwardEnable(s, targetSess, m) {
	// 	return
	// }

	// m.SetUid(s.UserID())
	// err = targetSess.Send(m)
	// if err != nil {
	// 	log.Error(err)
	// }
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

func (r *Router) onServerOnline(s network.Conn) {

}

func (r *Router) onServerOffline(s network.Conn) {

}

func (r *Router) onUserOnline(s network.Conn) {
	s.Store("loginat", time.Now())

	// uinfo.Groups.Range(func(k, v interface{}) bool {
	// 	r.gm.AddTo(k.(string), uinfo.UID, s)
	// 	return true
	// })

	r.PublishEvent("ConnEnable", map[string]any{
		"sid":    s.ConnID(),
		"uid":    s.UserID(),
		"enable": true,
	})
}

func (r *Router) onUserOffline(s network.Conn) {

	// uinfo.Groups.Range(func(k, v interface{}) bool {
	// 	r.gm.RemoveFromGroup(k.(string), uinfo.UID, s)
	// 	return true
	// })

	r.PublishEvent("ConnEnable", map[string]any{
		"sid":    s.ConnID(),
		"uid":    s.UserID(),
		"enable": false,
	})
}

func (r *Router) PublishEvent(ename string, event any) {
	log.Printf("[Mock PublishEvent] name:%v,event:%v", ename, event)
}

// func (r *Router) OnCall(s *network.Conn, msg *server.MsgWraper) {
// var err error

// msgid := int(head.GetMsgID())

// enable := r.callEnable(s, uint32(msgid))
// if !enable {
// 	log.Print("not enable to call this method:", msgid)
// 	dealSocketErrCnt(s)
// 	return
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
// }
