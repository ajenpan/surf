package main

import (
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/auth"
	"github.com/ajenpan/surf/log"
	"github.com/ajenpan/surf/server"
)

func NewRouter() (*Router, error) {
	ret := &Router{
		userSession: make(map[uint32]server.Session),
		// UserSessions :
	}

	return ret, nil
}

type Router struct {
	userSession     map[uint32]server.Session
	userSessionLock sync.RWMutex

	// UserSessions *UserSessions

	Selfinfo *auth.UserInfo
}

func (r *Router) OnSessionStatus(s server.Session, enable bool) {
	log.Debugf("OnSessionStatus: %v %v %v %v", s.SessionID(), s.RemoteAddr(), s.UserID(), enable)

	currSession := r.GetUserSession(s.UserID())
	if enable {
		if currSession != nil {
			r.RemoveUserSession(s.UserID())
			r.onUserOffline(currSession)
		}
		r.StoreUserSession(s.UserID(), s)
		r.onUserOnline(s)
	} else {
		if currSession != nil && currSession == s {
			r.RemoveUserSession(s.UserID())
			r.onUserOffline(currSession)
		}
	}
}

func (r *Router) OnSessionMessage(s server.Session, m *server.MsgWraper) {
	var err error
	targetuid := m.GetUid()

	if targetuid == 0 {
		//call my self
		r.OnCall(s, m)
		return
	}

	targetSess := r.GetUserSession(targetuid)
	if targetSess == nil {
		//TODO: send err to source
		log.Warnf("session uid:%v not found", targetuid)
		return
	}

	if !r.forwardEnable(s, targetSess, m) {
		return
	}

	m.SetUid(s.UserID())
	err = targetSess.Send(m)
	if err != nil {
		log.Error(err)
	}
}

func (r *Router) GetUserSession(uid uint32) server.Session {
	r.userSessionLock.RLock()
	defer r.userSessionLock.RUnlock()
	return r.userSession[uid]
}

func (r *Router) StoreUserSession(uid uint32, s server.Session) {
	r.userSessionLock.Lock()
	defer r.userSessionLock.Unlock()
	r.userSession[uid] = s
}

func (r *Router) RemoveUserSession(uid uint32) {
	r.userSessionLock.Lock()
	defer r.userSessionLock.Unlock()
	delete(r.userSession, uid)
}

func (r *Router) onUserOnline(s server.Session) {
	// uinfo.Groups.Range(func(k, v interface{}) bool {
	// 	r.gm.AddTo(k.(string), uinfo.UID, s)
	// 	return true
	// })

	// r.PublishEvent(&msg.UserStatChange{
	// 	Sid:  s.ID(),
	// 	Uid:  uinfo.UID,
	// 	Stat: msg.UserStatChange_Online,
	// })
}

func (r *Router) onUserOffline(s server.Session) {
	// uinfo.Groups.Range(func(k, v interface{}) bool {
	// 	r.gm.RemoveFromGroup(k.(string), uinfo.UID, s)
	// 	return true
	// })

	// r.PublishEvent(&msg.UserStatChange{
	// 	Sid:  s.ID(),
	// 	Uid:  uinfo.UID,
	// 	Stat: msg.UserStatChange_Offline,
	// })
}

func (r *Router) PublishEvent(event proto.Message) {
	log.Printf("[Mock PublishEvent] name:%v, msg:%v\n", string(proto.MessageName(event).Name()), event)
}

func (r *Router) OnCall(s server.Session, msg *server.MsgWraper) {
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
}

func (r *Router) forwardEnable(s server.Session, target server.Session, msg *server.MsgWraper) bool {
	// suinfo := GetSocketUserInfo(s)
	// tuinfo := GetSocketUserInfo(target)
	// if suinfo == nil || tuinfo == nil {
	// 	return false
	// }
	// return r.Permitc.ForwardEnable(suinfo.Role, tuinfo.Role, msgid)
	return true
}
