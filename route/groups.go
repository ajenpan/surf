package route

import (
	"sync"

	"github.com/ajenpan/surf/core/network"
)

type Groups struct {
	groups sync.Map
}

func (m *Groups) MustGetGroup(name string) *Group {
	v, _ := m.groups.LoadOrStore(name, NewGroup())
	return v.(*Group)
}

func (m *Groups) GetGroup(name string) *Group {
	if v, has := m.groups.Load(name); has {
		return v.(*Group)
	}
	return nil
}

func (m *Groups) RemoveFromGroup(name string, uid uint64, s *network.Conn) {
	if v, has := m.groups.Load(name); has {
		v.(*Group).RemoveIfSame(uid, s)
	}
}

func (m *Groups) AddTo(name string, uid uint64, s *network.Conn) {
	g := m.MustGetGroup(name)
	g.Add(uid, s)
}
