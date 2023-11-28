package route

import (
	"sync"

	"github.com/ajenpan/surf/server"
)

// type SessionMap struct {
// 	imp sync.Map
// }
// func (sm *SessionMap) Remove(sid string) bool {
// 	_, ok := sm.imp.LoadAndDelete(sid)
// 	return ok
// }

func NewSessionMap() *SessionMap {
	return &SessionMap{}
}

type SessionMap struct {
	imp sync.Map
}

func (sm *SessionMap) RemoveAll() int {
	cnt := 0
	sm.imp.Range(func(key, value interface{}) bool {
		s := value.(server.Session)
		s.Close()
		sm.imp.Delete(key)
		cnt++
		return true
	})
	return cnt
}

func (sm *SessionMap) Remove(sid string) bool {
	_, ok := sm.imp.LoadAndDelete(sid)
	return ok
}

func (sm *SessionMap) Store(s server.Session) bool {
	sm.imp.Store(s.UserID(), s)
	return true
}

type UserSessions struct {
	rwlock sync.RWMutex
	imp    map[uint64]*SessionMap
}

func (us *UserSessions) MustGetSessionMap(uid uint64) *SessionMap {
	us.rwlock.Lock()
	defer us.rwlock.Unlock()
	sm, has := us.imp[uid]
	if !has {
		sm = NewSessionMap()
		us.imp[uid] = sm
	}
	return sm
}

// func (us *UserSessions) RemoveSession(s session.Session) bool {
// 	us.rwlock.Lock()
// 	defer us.rwlock.Unlock()
// 	sm, has := us.imp[s.UID()]
// 	if !has {
// 		return false
// 	}
// 	return sm.Remove(s.ID())
// }
