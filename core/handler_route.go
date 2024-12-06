package core

import (
	"sync"
)

type HandlerRoute[T comparable] struct {
	sync.RWMutex
	methods map[T]HandlerFunc
}

func NewHandlerRoute[T comparable]() *HandlerRoute[T] {
	return &HandlerRoute[T]{
		methods: make(map[T]HandlerFunc),
	}
}

func (m *HandlerRoute[T]) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.methods)
}

func (m *HandlerRoute[T]) Get(id T) HandlerFunc {
	m.RLock()
	defer m.RUnlock()
	return m.methods[id]
}

func (m *HandlerRoute[T]) Range(f func(key T, value HandlerFunc) bool) {
	m.RLock()
	defer m.RUnlock()
	for k, v := range m.methods {
		if !f(k, v) {
			return
		}
	}
}

func (m *HandlerRoute[T]) Merge(other *HandlerRoute[T]) {
	other.RWMutex.RLock()
	defer other.RWMutex.RUnlock()

	m.Lock()
	defer m.Unlock()

	for k, v := range other.methods {
		m.methods[k] = v
	}
}

func (m *HandlerRoute[T]) Add(key T, method HandlerFunc) bool {
	m.Lock()
	defer m.Unlock()
	m.methods[key] = method
	return true
}

func (m *HandlerRoute[T]) Delete(key T) {
	m.Lock()
	defer m.Unlock()
	delete(m.methods, key)
}

func (m *HandlerRoute[T]) LoadAndDelete(key T) (HandlerFunc, bool) {
	m.Lock()
	defer m.Unlock()
	v, has := m.methods[key]
	if !has {
		return nil, false
	}
	delete(m.methods, key)
	return v, has
}
