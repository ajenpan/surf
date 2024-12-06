package core

import (
	"sync"

	"github.com/ajenpan/surf/core/network"
)

func NewClientConnStore(fn network.FuncOnConnEnable) *ClientConnStore {
	return &ClientConnStore{
		fn:        fn,
		ConnStore: NewConnStore(),
	}
}

type ClientConnStore struct {
	fn network.FuncOnConnEnable
	*ConnStore
}

func (store *ClientConnStore) OnConnEnable(conn network.Conn, enable bool) {
	if enable {
		log.Info("OnConnEnable", "id", conn.ConnId(), "addr", conn.RemoteAddr(), "uid", conn.UserID(), "urid", conn.UserRole(), "enable", enable)
		currConn, got := store.SwapByUId(conn)
		if got {
			ud := currConn.GetUserData()
			currConn.SetUserData(nil)

			conn.SetUserData(ud)
			log.Info("OnConnEnable: repeat conn, close old conn", "id", currConn.ConnId(), "uid", currConn.UserID())
			currConn.Close()
		} else {
			store.safeOnConnEnable(conn, true)
		}
	} else {
		currConn, got := store.DeleteByConnId(conn.ConnId())
		if got {
			store.safeOnConnEnable(currConn, false)
		}
	}
}

func (store *ClientConnStore) safeOnConnEnable(conn network.Conn, enable bool) {
	if store.fn != nil {
		store.fn(conn, enable)
	}
}

func NewNodeConnStore(fn network.FuncOnConnEnable) *NodeConnStore {
	return &NodeConnStore{
		fn:        fn,
		ConnStore: NewConnStore(),
	}
}

type NodeConnStore struct {
	fn network.FuncOnConnEnable
	*ConnStore
}

func (store *NodeConnStore) OnConnEnable(conn network.Conn, enable bool) {
	if enable {
		_, loaded := store.LoadOrStoreByUId(conn)
		if loaded {
			conn.Close()
			log.Error("repeat server conn", "connid", conn.ConnId(), "nodeid", conn.UserID(), "svrtype", conn.UserRole())
		} else {
			store.safeOnConnEnable(conn, true)
		}
	} else {
		currConn, got := store.DeleteByConnId(conn.ConnId())
		if got {
			store.safeOnConnEnable(currConn, false)
		}
	}
}

func (store *NodeConnStore) safeOnConnEnable(conn network.Conn, enable bool) {
	if store.fn != nil {
		store.fn(conn, enable)
	}
}

func NewConnStore() *ConnStore {
	return &ConnStore{
		uid2cid: make(map[uint32]string),
		conn:    make(map[string]network.Conn),
	}
}

type ConnStore struct {
	rwmutex sync.RWMutex
	uid2cid map[uint32]string
	conn    map[string]network.Conn
}

func (cs *ConnStore) LoadByUId(uid uint32) (network.Conn, bool) {
	cs.rwmutex.RLock()
	defer cs.rwmutex.RUnlock()
	cid, ok := cs.uid2cid[uid]
	if !ok {
		return nil, false
	}
	c, ok := cs.conn[cid]
	return c, ok
}

func (cs *ConnStore) SwapByUId(c network.Conn) (network.Conn, bool) {
	uid := c.UserID()

	cs.rwmutex.Lock()
	defer cs.rwmutex.Unlock()

	var ret network.Conn
	var ok bool = false

	cid, has := cs.uid2cid[uid]
	if has {
		ret, ok = cs.conn[cid]
		if ok {
			delete(cs.conn, cid)
		}
	}

	cs.uid2cid[uid] = c.ConnId()
	cs.conn[c.ConnId()] = c
	return ret, ok
}

func (cs *ConnStore) LoadOrStoreByUId(c network.Conn) (network.Conn, bool) {
	uid := c.UserID()

	cs.rwmutex.Lock()
	defer cs.rwmutex.Unlock()

	cid, ok := cs.uid2cid[uid]
	if !ok {
		cs.uid2cid[uid] = c.ConnId()
		cs.conn[c.ConnId()] = c
		return c, false
	}
	c, ok = cs.conn[cid]
	return c, ok
}

func (cs *ConnStore) LoadByConnId(cid string) (network.Conn, bool) {
	cs.rwmutex.RLock()
	defer cs.rwmutex.RUnlock()
	c, ok := cs.conn[cid]
	return c, ok
}

func (cs *ConnStore) Store(c network.Conn) {
	cs.rwmutex.Lock()
	defer cs.rwmutex.Unlock()
	cs.uid2cid[c.UserID()] = c.ConnId()
	cs.conn[c.ConnId()] = c
}

func (cs *ConnStore) DeleteByConnId(cid string) (network.Conn, bool) {
	cs.rwmutex.Lock()
	defer cs.rwmutex.Unlock()

	c, ok := cs.conn[cid]
	if !ok {
		return nil, false
	}
	delete(cs.uid2cid, c.UserID())
	delete(cs.conn, cid)
	return c, true
}

func (cs *ConnStore) Range(fn func(c network.Conn) bool) {
	cs.rwmutex.RLock()
	defer cs.rwmutex.RUnlock()
	for _, c := range cs.conn {
		if !fn(c) {
			break
		}
	}
}
