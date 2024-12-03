package core

import (
	"sync"
	"time"

	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/utils/calltable"
)

type RequestCallbackFunc func(timeout bool, pk *RoutePacket)

type RequestCallbackCache struct {
	cbfun   RequestCallbackFunc
	timeout *time.Timer
}

type PacketRouteCaller struct {
	Calltable *calltable.CallTable
	Handler   interface{}

	respWatier sync.Map
}

type storeKey struct {
	nodeid uint32
	syn    uint32
}

func (s *PacketRouteCaller) PushRespCallback(nodeid uint32, syn uint32, timeoutsec uint32, cb RequestCallbackFunc) error {
	timeout := time.AfterFunc(time.Duration(timeoutsec)*time.Second, func() {
		info := s.PopRespCallback(nodeid, syn)
		if info != nil && info.cbfun != nil {
			info.cbfun(true, nil)
		}

		if info.timeout != nil {
			info.timeout.Stop()
			info.timeout = nil
		}
	})

	cache := &RequestCallbackCache{
		cbfun:   cb,
		timeout: timeout,
	}

	s.respWatier.Store(storeKey{nodeid, syn}, cache)
	return nil
}

func (s *PacketRouteCaller) PopRespCallback(nodeid uint32, syn uint32) *RequestCallbackCache {
	cache, ok := s.respWatier.Load(storeKey{nodeid, syn})
	if !ok {
		return nil
	}
	return cache.(*RequestCallbackCache)
}

func (p *PacketRouteCaller) Call(ctx *ConnContext) {
	rpk := ctx.ReqPacket

	if rpk.GetSubType() != 0 {
		log.Error("recv err packet subtype", "subtype", rpk.GetSubType())
		return
	}

	switch rpk.GetMsgType() {
	case RoutePackMsgType_Async:
		fallthrough
	case RoutePackMsgType_Request:
		method := p.Calltable.GetByID(rpk.GetMsgId())
		if method == nil {
			log.Error("not found msg handler", "msgid", rpk.GetMsgId(), "from_uid", rpk.GetFromUID(), "from_svrtype", rpk.GetFromURole(), "to_uid", rpk.GetToUID(), "to_svrtype", rpk.GetToURole())
			return
		}
		marshaler := marshal.NewMarshalerById(rpk.GetMarshalType())
		if marshaler == nil {
			log.Error("invalid marshaler type", "type", rpk.GetMarshalType())
			//todo send error packet
			return
		}
		req := method.NewRequest()
		err := marshaler.Unmarshal(rpk.GetBody(), req)
		if err != nil {
			log.Error("unmarshal request body failed", "err", err)
			//todo send error packet
			return
		}

		method.Call(p.Handler, ctx, req)
	case RoutePackMsgType_Response:
		cbinfo := p.PopRespCallback(ctx.Conn.UserID(), rpk.GetSYN())
		if cbinfo == nil {
			return
		}
		if cbinfo.timeout != nil {
			cbinfo.timeout.Stop()
		}
		if cbinfo.cbfun != nil {
			cbinfo.cbfun(false, rpk)
		}
	default:
	}
}
