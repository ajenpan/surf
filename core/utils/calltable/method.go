package calltable

import (
	"reflect"
	"sync"
)

type Method struct {
	Handler interface{}

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
	var values []reflect.Value

	if m.Handler == nil {
		argc := len(args)
		values = make([]reflect.Value, argc)
		for i, v := range args {
			values[i] = reflect.ValueOf(v)
		}
	} else {
		argc := len(args) + 1
		values = make([]reflect.Value, argc)
		values[0] = reflect.ValueOf(m.Handler)
		for i, v := range args {
			values[i+1] = reflect.ValueOf(v)
		}
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
