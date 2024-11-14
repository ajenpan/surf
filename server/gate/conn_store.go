package gate

import (
	"sync"

	"github.com/ajenpan/surf/core/network"
)

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

func (cs *ConnStore) LoadByUID(uid uint32) (network.Conn, bool) {
	cs.rwmutex.RLock()
	defer cs.rwmutex.RUnlock()
	cid, ok := cs.uid2cid[uid]
	if !ok {
		return nil, false
	}
	c, ok := cs.conn[cid]
	return c, ok
}

func (cs *ConnStore) SwapByUID(c network.Conn) (network.Conn, bool) {
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

	cs.uid2cid[uid] = c.ConnID()
	cs.conn[c.ConnID()] = c
	return ret, ok
}

func (cs *ConnStore) LoadOrStoreByUID(c network.Conn) (network.Conn, bool) {
	uid := c.UserID()

	cs.rwmutex.Lock()
	defer cs.rwmutex.Unlock()

	cid, ok := cs.uid2cid[uid]
	if !ok {
		cs.uid2cid[uid] = c.ConnID()
		cs.conn[c.ConnID()] = c
		return c, false
	}
	c, ok = cs.conn[cid]
	return c, ok
}

func (cs *ConnStore) LoadByCID(cid string) (network.Conn, bool) {
	cs.rwmutex.RLock()
	defer cs.rwmutex.RUnlock()
	c, ok := cs.conn[cid]
	return c, ok
}

func (cs *ConnStore) Store(c network.Conn) {
	cs.rwmutex.Lock()
	defer cs.rwmutex.Unlock()
	cs.uid2cid[c.UserID()] = c.ConnID()
	cs.conn[c.ConnID()] = c
}

func (cs *ConnStore) Delete(cid string) (network.Conn, bool) {
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
