package calltable2

import (
	"sync"
)

type CallTable[T comparable] struct {
	sync.RWMutex
	methods map[T]Method
}

func NewCallTable[T comparable]() *CallTable[T] {
	return &CallTable[T]{
		methods: make(map[T]Method),
	}
}

func (m *CallTable[T]) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.methods)
}

func (m *CallTable[T]) GetByID(id T) Method {
	m.RLock()
	defer m.RUnlock()
	return m.methods[id]
}

func (m *CallTable[T]) Range(f func(key T, value Method) bool) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.methods {
		if !f(k, v) {
			return
		}
	}
}

func (m *CallTable[T]) Merge(other *CallTable[T]) {
	other.RWMutex.RLock()
	defer other.RWMutex.RUnlock()

	m.Lock()
	defer m.Unlock()

	for k, v := range other.methods {
		m.methods[k] = v
	}
}

func (m *CallTable[T]) Add(key T, method Method) bool {
	m.Lock()
	defer m.Unlock()
	m.methods[key] = method
	return true
}
