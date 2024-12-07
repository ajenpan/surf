package calltable

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const MethodPrefix string = "On"
const MsgPrefix string = "Req"
const MsgSuffix string = "Request"

func ExtractParseGRpcMethod(ms protoreflect.ServiceDescriptors, h interface{}) *CallTable {
	refh := reflect.TypeOf(h)

	ret := NewCallTable()

	pbMsgType := reflect.TypeOf((*proto.Message)(nil)).Elem()
	errType := reflect.TypeOf((*error)(nil)).Elem()

	for i := 0; i < ms.Len(); i++ {
		service := ms.Get(i)
		methods := service.Methods()

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

			if !methodt.In(1).Implements(pbMsgType) {
				continue
			}
			if !methodt.Out(0).Implements(pbMsgType) {
				continue
			}
			if methodt.Out(1) != errType {
				continue
			}
			reqType := methodt.In(1).Elem()
			respType := methodt.Out(0).Elem()

			m := &Method{
				Name:         rpcMethodName,
				Func:         methodv.Func,
				RequestType:  reqType,
				ResponseType: respType,
			}
			m.InitPool()

			ret.Add(m)
		}
	}
	return ret
}

func ExtractAsyncMethod(ms protoreflect.MessageDescriptors, h interface{}) *CallTable {
	const MethodPrefix string = "On"
	refh := reflect.TypeOf(h)

	ret := NewCallTable()
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
			Name:        method.Name,
			Func:        method.Func,
			RequestType: reqMsgType.Elem(),
		}
		m.InitPool()
		ret.Add(m)
	}
	return ret
}

func GetMsgId(msg proto.Message) uint32 {
	md := msg.ProtoReflect().Descriptor()
	return GetMsgIDFromDesc(md)
}

func GetMsgIDFromDesc(md protoreflect.MessageDescriptor) uint32 {
	msgDesc := md.Enums().ByName("MSGID")
	if msgDesc == nil {
		return 0
	}
	idDesc := msgDesc.Values().ByName("ID")
	if idDesc == nil {
		return 0
	}
	return uint32(idDesc.Number())
}

func MustExtractFunction(f interface{}) *Method {
	m, err := ExtractFunction(f)
	if err != nil {
		panic(err)
	}
	return m
}

func ExtractFunction(f interface{}) (*Method, error) {
	refv := reflect.ValueOf(f)
	if refv.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}
	if refv.Type().NumIn() != 2 {
		return nil, fmt.Errorf("param num error")
	}
	reqtype := refv.Type().In(1).Elem()
	msg, ok := reflect.New(reqtype).Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("not a proto message")
	}

	id := GetMsgId(msg)
	ret := &Method{
		Name:        string(msg.ProtoReflect().Descriptor().Name()),
		ID:          id,
		Func:        refv,
		RequestType: reqtype,
	}
	return ret, nil
}

func ExtractMethodFromDesc(ms protoreflect.MessageDescriptors, h interface{}) *CallTable {
	ret := NewCallTable()

	hvalue := reflect.TypeOf(h)
	pbMsgType := reflect.TypeOf((*proto.Message)(nil)).Elem()

	for i := 0; i < ms.Len(); i++ {
		msg := ms.Get(i)
		msgid := GetMsgIDFromDesc(msg)
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
			Name:        hname,
			ID:          msgid,
			Func:        method.Func,
			RequestType: reqMsgType.Elem(),
		}
		m.InitPool()

		ret.Add(m)
	}
	return ret
}
