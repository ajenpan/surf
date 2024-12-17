package core

import (
	"sync"
	"sync/atomic"
)

// todo:
// type RotateSelect interface {
// 	Next() ItemT
// 	Add(key ItemT)
// 	Del(key ItemT)
// }

type RotateSelect[ItemT comparable] struct {
	rwl  sync.RWMutex
	idx  atomic.Uint32
	list []ItemT
}

func (ns *RotateSelect[ItemT]) Next() ItemT {
	ns.rwl.RLock()
	defer ns.rwl.RUnlock()

	if len(ns.list) == 0 {
		var v ItemT
		return v
	}
	idx := int(ns.idx.Add(1)) % len(ns.list)
	return ns.list[idx]
}

func (ns *RotateSelect[ItemT]) Add(key ItemT) {
	ns.rwl.Lock()
	defer ns.rwl.Unlock()

	ns.list = append(ns.list, key)
}

func (ns *RotateSelect[ItemT]) Del(key ItemT) {
	ns.rwl.Lock()
	defer ns.rwl.Unlock()

	for i, value := range ns.list {
		if value == key {
			ns.list = append(ns.list[:i], ns.list[i+1:]...)
			break
		}
	}
}

func (ns *RotateSelect[ItemT]) Size() int {
	ns.rwl.RLock()
	defer ns.rwl.RUnlock()

	return len(ns.list)
}

type NodeGroup struct {
	m         map[uint32]*nodeRegistryData
	selecters map[uint16]*RotateSelect[uint32]
	lock      sync.RWMutex
}

func NewGroup() *NodeGroup {
	return &NodeGroup{
		m: make(map[uint32]*nodeRegistryData),
	}
}

func (g *NodeGroup) Set(item *nodeRegistryData) {
	g.lock.Lock()
	defer g.lock.Unlock()

	_, has := g.m[item.Node.NId]
	g.m[item.Node.NodeID()] = item

	if !has {
		selecter, got := g.selecters[item.Node.NType]
		if !got {
			selecter = &RotateSelect[uint32]{}
			g.selecters[item.Node.NType] = selecter
		}
		selecter.idx.Add(item.Node.NId)
	}
}

func (g *NodeGroup) Choice(ntype uint16) *nodeRegistryData {
	g.lock.RLock()
	defer g.lock.RUnlock()

	selecter, got := g.selecters[ntype]
	if !got {
		return nil
	}
	size := selecter.Size()

	var ret *nodeRegistryData
	for i := 0; i < size; i++ {
		nid := selecter.Next()
		v, has := g.m[nid]
		if !has {
			continue
		}

		if v.Status == NodeState_Running {
			ret = v
			break
		}
	}
	return ret
}

func (g *NodeGroup) Del(uid uint32) *nodeRegistryData {
	g.lock.Lock()
	defer g.lock.Unlock()

	info, has := g.m[uid]
	if !has {
		return nil
	}
	if selecter := g.selecters[info.Node.NType]; selecter != nil {
		selecter.Del(uid)
	}
	return info
}

func (g *NodeGroup) Get(nid uint32) *nodeRegistryData {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.m[nid]
}

func (g *NodeGroup) Size() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return len(g.m)
}
