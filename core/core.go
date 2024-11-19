package core

import (
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/errors"
	logger "github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
)

var GSurf = &Surf{}

var log = logger.Default

func Init(opt Options) error {
	return GSurf.Init(opt)
}

func Run() error {
	return GSurf.Run()
}

func SendRequestToClient[T proto.Message](uid uint32, msgid uint32, msg any, cb func(timeout bool, err error, resp *T)) {
	resp := new(T)

	wrapfn := func(timeout bool, rpk *RoutePacket) {
		if timeout {
			cb(true, nil, resp)
			return
		}

		errcode := rpk.GetErrCode()
		if errcode != 0 {
			cb(false, errors.New(int32(errcode), "resp err"), nil)
			return
		}
		err := GSurf.opts.Marshaler.Unmarshal(rpk.GetBody(), resp)
		if err != nil {
			cb(false, err, nil)
			return
		}
		cb(false, nil, resp)
	}

	err := GSurf.SendRequestToClientByUId(uid, msgid, msg, wrapfn)

	if err != nil && cb != nil {
		cb(false, err, nil)
	}
}

func OnRouteAsync(ct *calltable.CallTable, conn network.Conn, pk *network.HVPacket) {
	// var err error
	// rpk := network.RoutePacketRaw(pk.GetBody())

	// method := ct.Get(pk.Head.GetMsgId())
	// if method == nil {
	// 	pk.Head.SetType(network.PacketType_HandleErr)
	// 	pk.Head.SetSubFlag(network.PacketType_HandleErr_MethodNotFound)
	// 	pk.SetBody(nil)
	// 	conn.Send(pk)
	// 	return
	// }

	// msg := method.NewRequest()

	// if pk.Head.GetBodyLen() > 0 {
	// 	mar := &proto.UnmarshalOptions{}
	// 	err := mar.Unmarshal(pk.GetBody(), msg.(proto.Message))
	// 	if err != nil {
	// 		pk.Head.SetType(network.PacketType_HandleErr)
	// 		pk.Head.SetSubFlag(network.PacketType_HandleErr_MethodParseErr)
	// 		pk.SetBody(nil)
	// 		conn.Send(pk)
	// 		return
	// 	}

	// }

	// var ctx Context = &context{
	// 	Conn: conn,
	// 	Core: nil,
	// 	Raw:  pk,
	// }

	// method.Call(ctx, msg)
}

// type HandlerRegister struct {
// 	asyncHLock  sync.RWMutex
// 	asyncHandle map[string]FuncAsyncHandle

// 	requestHLock  sync.RWMutex
// 	requestHandle map[string]FuncRequestHandle
// }

// func (hr *HandlerRegister) getAsyncCallbcak(name string) FuncAsyncHandle {
// 	hr.asyncHLock.RLock()
// 	defer hr.asyncHLock.RUnlock()
// 	return hr.asyncHandle[name]
// }
// func (hr *HandlerRegister) getRequestCallback(name string) FuncRequestHandle {
// 	hr.requestHLock.RLock()
// 	defer hr.requestHLock.RUnlock()
// 	return hr.requestHandle[name]
// }

// func (hr *HandlerRegister) RegisterAysncHandle(name string, cb FuncAsyncHandle) {
// 	hr.asyncHLock.Lock()
// 	defer hr.asyncHLock.Unlock()
// 	hr.asyncHandle[name] = cb
// }

// func (hr *HandlerRegister) RegisterRequestHandle(name string, cb FuncRequestHandle) {
// 	hr.requestHLock.Lock()
// 	defer hr.requestHLock.Unlock()
// 	hr.requestHandle[name] = cb
// }

// func (hr *HandlerRegister) OnServerMsgWraper(ctx *Context, m *network.MsgWraper) bool {
//
//	if m.GetMsgtype() == network.MsgTypeAsync {
//		wrap := &network.AsyncMsg{}
//		proto.Unmarshal(m.GetBody(), wrap)
//		cb := hr.getAsyncCallbcak(wrap.GetName())
//		if cb != nil {
//			cb(ctx, wrap)
//			return true
//		}
//	} else if m.GetMsgtype() == network.MsgTypeRequest {
//
//		wrap := &network.RequestMsg{}
//		proto.Unmarshal(m.GetBody(), wrap)
//		cb := hr.getRequestCallback(wrap.GetName())
//		if cb != nil {
//			cb(ctx, wrap)
//			return true
//		}
//	}
//
// return false
// }
