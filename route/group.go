package route

import (
	"sync"

	"github.com/ajenpan/surf/core/network"

	"github.com/emirpasic/gods/maps/treemap"
)

type Group struct {
	imp  *treemap.Map
	lock sync.RWMutex
}

func NewGroup() *Group {
	return &Group{
		imp: treemap.NewWithIntComparator(),
	}
}

func (g *Group) Add(uid uint64, s *network.Conn) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.imp.Put(uid, s)
}

func (g *Group) RemoveIfSame(uid uint64, s *network.Conn) {
	g.lock.Lock()
	defer g.lock.Unlock()
	if v, found := g.imp.Get(uid); found {
		if v.(*network.Conn) == s {
			g.imp.Remove(uid)
		}
	}
}

func (g *Group) Get(uid uint64) *network.Conn {
	g.lock.RLock()
	defer g.lock.RUnlock()
	if v, found := g.imp.Get(uid); found {
		return v.(*network.Conn)
	}
	return nil
}

func (g *Group) Size() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.imp.Size()
}
func (g *Group) GetAll() []*network.Conn {
	g.lock.RLock()
	defer g.lock.RUnlock()

	ret := make([]*network.Conn, 0, g.imp.Size())
	g.imp.All(func(key, value interface{}) bool {
		ret = append(ret, value.(*network.Conn))
		return true
	})
	return ret
}

func (g *Group) Range(startAt, endAt int) ([]*network.Conn, int) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	if endAt > g.imp.Size() {
		endAt = g.imp.Size()
	}
	if startAt > endAt {
		startAt = endAt
	}

	total := g.imp.Size()
	iter := g.imp.Iterator()

	for i := 0; i < startAt; i++ {
		if !iter.Next() {
			return nil, total
		}
	}

	cnt := endAt - startAt
	ret := make([]*network.Conn, 0, cnt)
	for i := 0; i < cnt; i++ {
		ret = append(ret, iter.Value().(*network.Conn))
		if !iter.Next() {
			break
		}
	}
	return ret, total
}
