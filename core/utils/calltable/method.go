package calltable

import (
	"reflect"
	"sync"
)

type Method struct {
	Name string
	ID   uint32

	Func reflect.Value

	RequestType  reflect.Type
	ResponseType reflect.Type

	reqPool  *sync.Pool
	respPool *sync.Pool
}

func (m *Method) InitPool() {
	if m.RequestType != nil {
		m.reqPool = &sync.Pool{New: m.NewRequest}
	}

	if m.ResponseType != nil {
		m.respPool = &sync.Pool{New: m.NewResponse}
	}
}

func (m *Method) Call(args ...interface{}) []reflect.Value {
	values := make([]reflect.Value, len(args))
	for i, v := range args {
		values[i] = reflect.ValueOf(v)
	}
	return m.Func.Call(values)
}

func (m *Method) NewRequest() interface{} {
	return reflect.New(m.RequestType).Interface()
}

func (m *Method) NewResponse() interface{} {
	return reflect.New(m.ResponseType).Interface()
}

func (m *Method) GetRequest() interface{} {
	if m.reqPool == nil {
		return m.NewRequest()
	}
	return m.reqPool.Get()
}

func (m *Method) PutRequest(req interface{}) {
	if m.reqPool == nil {
		return
	}
	m.reqPool.Put(req)
}

func (m *Method) GetResponse() interface{} {
	if m.respPool == nil {
		return m.NewResponse()
	}
	return m.respPool.Get()
}

func (m *Method) PutResponse(resp interface{}) {
	if m.respPool == nil {
		return
	}
	m.respPool.Put(resp)
}
