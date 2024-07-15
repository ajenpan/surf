package calltable

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func ExtractParseGRpcMethod(ms protoreflect.ServiceDescriptors, h interface{}) *CallTable[string] {
	refh := reflect.TypeOf(h)

	ret := NewCallTable[string]()

	// ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	pbMsgType := reflect.TypeOf((*proto.Message)(nil)).Elem()
	errType := reflect.TypeOf((*error)(nil)).Elem()

	for i := 0; i < ms.Len(); i++ {
		service := ms.Get(i)
		methods := service.Methods()
		// svrName := string(service.Name())

		for j := 0; j < methods.Len(); j++ {
			rpcMethod := methods.Get(j)
			rpcMethodName := string(rpcMethod.Name())

			methodv, has := refh.MethodByName(rpcMethodName)
			if !has {
				continue
			}
			methodt := methodv.Type

			if methodt.NumIn() != 2 || methodt.NumOut() != 2 {
				continue
			}
			// if method.In(0) != ctxType {
			// 	continue
			// }
			if !methodt.In(1).Implements(pbMsgType) {
				continue
			}
			if !methodt.Out(0).Implements(pbMsgType) {
				continue
			}
			if methodt.Out(1) != errType {
				continue
			}
			// epn := strings.Join([]string{svrName, rpcMethodName}, "/")
			reqType := methodt.In(1).Elem()
			respType := methodt.Out(0).Elem()

			m := &Method{
				FuncName:     rpcMethodName,
				Func:         methodv.Func,
				Style:        StyleGRpc,
				RequestType:  reqType,
				ResponseType: respType,
			}
			m.InitPool()

			ret.list[rpcMethodName] = m
		}
	}
	return ret
}

func ExtractAsyncMethod(ms protoreflect.MessageDescriptors, h interface{}) *CallTable[string] {
	const MethodPrefix string = "On"
	refh := reflect.TypeOf(h)

	ret := NewCallTable[string]()
	pbMsgType := reflect.TypeOf((*proto.Message)(nil)).Elem()

	for i := 0; i < ms.Len(); i++ {
		msg := ms.Get(i)
		msgName := string(msg.Name())
		method, has := refh.MethodByName(MethodPrefix + msgName)
		if !has {
			continue
		}

		if method.Type.NumIn() != 3 {
			continue
		}

		reqMsgType := method.Type.In(2)
		if reqMsgType.Kind() != reflect.Ptr {
			continue
		}
		if !reqMsgType.Implements(pbMsgType) {
			continue
		}

		m := &Method{
			FuncName:    method.Name,
			Func:        method.Func,
			Style:       StyleAsync,
			RequestType: reqMsgType.Elem(),
		}
		m.InitPool()
		ret.list[msgName] = m
	}
	return ret
}

func ExtractProtoFile(fd protoreflect.FileDescriptor, handler interface{}) *CallTable[string] {
	ret := NewCallTable[string]()

	rpcTable := ExtractParseGRpcMethod(fd.Services(), handler)
	asyncTalbe := ExtractAsyncMethod(fd.Messages(), handler)

	ret.Merge(rpcTable, false)
	ret.Merge(asyncTalbe, false)

	return ret
}

func GetMessageMsgID(msg protoreflect.MessageDescriptor) uint32 {
	MSGIDDesc := msg.Enums().ByName("MSGID")
	if MSGIDDesc == nil {
		return 0
	}
	IDDesc := MSGIDDesc.Values().ByName("ID")
	if IDDesc == nil {
		return 0
	}
	return uint32(IDDesc.Number())
}

func ExtractAsyncMethodByMsgID(ms protoreflect.MessageDescriptors, h interface{}) *CallTable[uint32] {
	const MethodPrefix string = "On"
	hvalue := reflect.TypeOf(h)

	ret := NewCallTable[uint32]()
	pbMsgType := reflect.TypeOf((*proto.Message)(nil)).Elem()

	for i := 0; i < ms.Len(); i++ {
		msg := ms.Get(i)
		msgid := GetMessageMsgID(msg)
		if msgid == 0 {
			continue
		}
		msgName := string(msg.Name())
		method, has := hvalue.MethodByName(MethodPrefix + msgName)
		if !has {
			continue
		}

		if method.Type.NumIn() != 2 {
			continue
		}

		reqMsgType := method.Type.In(1)
		if reqMsgType.Kind() != reflect.Ptr {
			continue
		}

		if !reqMsgType.Implements(pbMsgType) {
			continue
		}
		m := &Method{
			Func:        method.Func,
			Style:       StyleAsync,
			RequestType: reqMsgType.Elem(),
		}
		m.InitPool()
		ret.Add(msgid, m)
	}
	return ret
}
