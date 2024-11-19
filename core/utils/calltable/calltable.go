package calltable

import (
	"sync"
)

type CallTable struct {
	sync.RWMutex
	byID   map[uint32]*Method
	byName map[string]*Method
}

func NewCallTable() *CallTable {
	return &CallTable{
		byID:   make(map[uint32]*Method),
		byName: make(map[string]*Method),
	}
}

func (m *CallTable) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.byID)
}

func (m *CallTable) GetByID(id uint32) *Method {
	m.RLock()
	defer m.RUnlock()
	return m.byID[id]
}

func (m *CallTable) RangeByID(f func(key uint32, value *Method) bool) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.byID {
		if !f(k, v) {
			return
		}
	}
}

func (m *CallTable) RangeByName(f func(key string, value *Method) bool) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.byName {
		if !f(k, v) {
			return
		}
	}
}

func (m *CallTable) Merge(other *CallTable) {
	other.RWMutex.RLock()
	defer other.RWMutex.RUnlock()

	m.Lock()
	defer m.Unlock()

	for k, v := range other.byID {
		m.byID[k] = v
		m.byName[v.Name] = v
	}
}

func (m *CallTable) Add(method *Method) bool {
	m.Lock()
	defer m.Unlock()
	if method.ID > 0 {
		m.byID[method.ID] = method
	}
	if len(method.Name) > 0 {
		m.byName[method.Name] = method
	}
	return true
}
