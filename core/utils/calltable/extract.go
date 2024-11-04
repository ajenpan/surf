package calltable

import (
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const MethodPrefix string = "On"
const MsgPrefix string = "Req"
const MsgSuffix string = "Request"

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
				HandleName:   rpcMethodName,
				Func:         methodv.Func,
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
			HandleName:  method.Name,
			Func:        method.Func,
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

func ExtractFunction(f interface{}) *Method {
	refv := reflect.ValueOf(f)
	if refv.Kind() != reflect.Func {
		return nil
	}
	reqtype := refv.Type().In(1).Elem()
	ret := &Method{
		Func:        refv,
		RequestType: reqtype,
	}
	return ret
}

func ExtractMethodFromDesc(ms protoreflect.MessageDescriptors, h interface{}) (*CallTable[uint32], *CallTable[string]) {
	ctByID := NewCallTable[uint32]()
	ctByName := NewCallTable[string]()

	hvalue := reflect.TypeOf(h)
	pbMsgType := reflect.TypeOf((*proto.Message)(nil)).Elem()

	for i := 0; i < ms.Len(); i++ {
		msg := ms.Get(i)
		msgid := GetMessageMsgID(msg)
		msgName := string(msg.Name())

		method, has := hvalue.MethodByName(MethodPrefix + msgName)
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

		hname := msgName
		hname = strings.TrimPrefix(hname, MsgPrefix)
		hname = strings.TrimSuffix(hname, MsgSuffix)

		m := &Method{
			HandleName:  hname,
			HandleMsgid: msgid,
			Func:        method.Func,
			RequestType: reqMsgType.Elem(),
		}
		m.InitPool()

		if msgid > 0 {
			ctByID.Add(msgid, m)
		}
		if len(hname) > 0 {
			ctByName.Add(hname, m)
		}
	}
	return ctByID, ctByName
}
