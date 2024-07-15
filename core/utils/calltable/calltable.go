package calltable

import (
	"sync"
)

type CallTable[T comparable] struct {
	sync.RWMutex
	list map[T]*Method
}

func NewCallTable[T comparable]() *CallTable[T] {
	return &CallTable[T]{
		list: make(map[T]*Method),
	}
}

func (m *CallTable[T]) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.list)
}

func (m *CallTable[T]) Get(name T) *Method {
	m.RLock()
	defer m.RUnlock()
	return m.list[name]
}

func (m *CallTable[T]) Range(f func(key T, value *Method) bool) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.list {
		if !f(k, v) {
			return
		}
	}
}

func (m *CallTable[T]) Merge(other *CallTable[T], overWrite bool) int {
	ret := 0
	other.RWMutex.RLock()
	defer other.RWMutex.RUnlock()

	m.Lock()
	defer m.Unlock()

	for k, v := range other.list {
		_, has := m.list[k]
		if has && !overWrite {
			continue
		}
		m.list[k] = v
		ret++
	}
	return ret
}

func (m *CallTable[T]) Add(name T, method *Method) bool {
	m.Lock()
	defer m.Unlock()
	if _, has := m.list[name]; has {
		return false
	}
	m.list[name] = method
	return true
}
