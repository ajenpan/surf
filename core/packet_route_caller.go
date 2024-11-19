package core

import (
	"sync"
	"sync/atomic"
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
	calltable *calltable.CallTable
	Handler   interface{}

	respWatier sync.Map
	synIdx     uint32
}

func (s *PacketRouteCaller) GetSYN() uint32 {
	ret := atomic.AddUint32(&s.synIdx, 1)
	if ret == 0 {
		return atomic.AddUint32(&s.synIdx, 1)
	}
	return ret
}

func (s *PacketRouteCaller) pushRespCallback(syn uint32, cb RequestCallbackFunc) error {
	timeout := time.AfterFunc(3*time.Second, func() {
		info := s.popRespCallback(syn)
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

	s.respWatier.Store(syn, cache)
	return nil
}

func (s *PacketRouteCaller) popRespCallback(syn uint32) *RequestCallbackCache {
	cache, ok := s.respWatier.Load(syn)
	if !ok {
		return nil
	}
	return cache.(*RequestCallbackCache)
}

func (p *PacketRouteCaller) Call(ctx *ConnContext) {
	rpk := ctx.ReqPacket

	if rpk.GetSubType() != 0 {
		log.Error("recv err packet subtype:", rpk.GetSubType())
		return
	}

	switch rpk.GetMsgType() {
	case RoutePackMsgType_Async:
		fallthrough
	case RoutePackMsgType_Request:
		method := p.calltable.GetByID(rpk.GetMsgId())
		if method == nil {
			log.Errorf("not found msg handler by msgid:%v,from_uid:%v,from_svrtype:%v,to_uid:%v,to_svrtype:%v",
				rpk.GetMsgId(), rpk.GetFromUID(), rpk.GetFromURole(), rpk.GetToUID(), rpk.GetToURole())
			return
		}
		marshaler := marshal.NewMarshalerById(rpk.GetMarshalType())
		if marshaler == nil {
			log.Error("invalid marshaler type:", rpk.GetMarshalType())
			//todo send error packet
			return
		}
		req := method.NewRequest()
		err := marshaler.Unmarshal(rpk.GetBody(), req)
		if err != nil {
			log.Error("unmarshal request body failed:", err)
			//todo send error packet
			return
		}

		method.Call(p.Handler, ctx, req)
	case RoutePackMsgType_Response:
		cbinfo := p.popRespCallback(rpk.GetSYN())
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
