package handle

import (
	"reflect"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type MessageMethod struct {
	Method reflect.Method

	Req  reflect.Type
	Resp reflect.Type
}

func (mm *MessageMethod) Call(args []reflect.Value) []reflect.Value {
	return mm.Method.Func.Call(args)
}

type CallTable struct {
	sync.RWMutex
	list map[string]*MessageMethod
}

func (m *CallTable) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.list)
}

func (m *CallTable) Has(name string) bool {
	m.RLock()
	defer m.RUnlock()
	_, has := m.list[name]
	return has
}

func (m *CallTable) Get(name string) *MessageMethod {
	m.RLock()
	defer m.RUnlock()

	ret, has := m.list[name]
	if has {
		return ret
	}
	return nil
}

func (m *CallTable) GetCallByMsg(msg proto.Message) *MessageMethod {
	m.RLock()
	defer m.RUnlock()
	name := string(msg.ProtoReflect().Descriptor().Name())
	ret, has := m.list[name]
	if has {
		return ret
	}
	return nil
}

func (m *CallTable) Range(f func(key string, value *MessageMethod) bool) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.list {
		if !f(k, v) {
			return
		}
	}
}

func (m *CallTable) Merge(other *CallTable, overWrite bool) int {
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

func ParseProtoMessageWithSuffix(suffix string, ms protoreflect.MessageDescriptors, handler interface{}) *CallTable {
	ret := &CallTable{
		list: make(map[string]*MessageMethod),
	}

	refHandler := reflect.TypeOf(handler)

	for i := 0; i < ms.Len(); i++ {
		msg := ms.Get(i)
		requestName := string(msg.Name())
		if !strings.HasSuffix(requestName, suffix) {
			continue
		}
		method, has := refHandler.MethodByName(requestName)
		if !has {
			continue
		}
		ret.list[requestName] = &MessageMethod{
			Method: method,
			// Req:    msg,
		}
	}
	return ret
}

//ParseRpcMethod
func ParseRpcMethod(ms protoreflect.ServiceDescriptors, h interface{}) *CallTable {
	ret := &CallTable{
		list: make(map[string]*MessageMethod),
	}

	refh := reflect.TypeOf(h)
	for i := 0; i < ms.Len(); i++ {
		rpcName := string(ms.Get(i).Name())
		rpcMethods := ms.Get(i).Methods()
		for j := 0; j < rpcMethods.Len(); j++ {
			rpcMethod := rpcMethods.Get(j)
			rpcMethodName := string(rpcMethod.Name())

			method, has := refh.MethodByName(rpcMethodName)
			if !has {
				continue
			}
			epn := strings.Join([]string{rpcName, rpcMethodName}, "/")

			ret.list[epn] = &MessageMethod{
				Method: method,
				Req:    method.Type.In(2).Elem(),
				Resp:   method.Type.In(3).Elem(),
				// ReqDesc:  rpcMethod.Input(),
				// RespDesc: rpcMethod.Output(),
			}
		}
	}
	return ret
}
