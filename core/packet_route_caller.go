package core

import (
	"sync"
	"time"
)

type RequestCallbackFunc func(timeout bool, pk *RoutePacket)

type RequestCallbackCache struct {
	cbfun   RequestCallbackFunc
	timeout *time.Timer
}

type RequestRouteKey struct {
	NType uint16
	MsgId uint32
}

type AyncRouteKey struct {
	NType uint16
	MsgId uint32
}

type ResponseRouteKey struct {
	NId uint32
	SYN uint32
}

func NewPacketRouteCaller() *PacketRouteCaller {
	return &PacketRouteCaller{
		requestRoute: NewHandlerRoute[RequestRouteKey](),
		ayncRoute:    NewHandlerRoute[AyncRouteKey](),
	}
}

type PacketRouteCaller struct {
	respWatier   sync.Map
	requestRoute *HandlerRoute[RequestRouteKey]
	ayncRoute    *HandlerRoute[AyncRouteKey]
}

func (s *PacketRouteCaller) PushRespCallback(key ResponseRouteKey, timeoutsec uint32, cb RequestCallbackFunc) error {
	var timeout *time.Timer = nil

	if timeoutsec >= 1 {
		timeout = time.AfterFunc(time.Duration(timeoutsec)*time.Second, func() {
			info := s.PopRespCallback(key)
			if info != nil && info.cbfun != nil {
				info.cbfun(true, nil)
			}
		})
	}

	cache := &RequestCallbackCache{
		cbfun:   cb,
		timeout: timeout,
	}

	s.respWatier.Store(key, cache)
	return nil
}

func (s *PacketRouteCaller) PopRespCallback(key ResponseRouteKey) *RequestCallbackCache {
	cache, ok := s.respWatier.LoadAndDelete(key)
	if !ok {
		return nil
	}
	ret := cache.(*RequestCallbackCache)
	if ret.timeout != nil {
		ret.timeout.Stop()
		ret.timeout = nil
	}
	return ret
}

func (p *PacketRouteCaller) Call(ctx Context) {
	rpk := ctx.Packet()

	if rpk.GetSubType() != 0 {
		log.Error("recv err packet subtype", "subtype", rpk.GetSubType())
		return
	}

	switch rpk.GetMsgType() {
	case RoutePackMsgType_Async:
		handles := p.ayncRoute.Get(AyncRouteKey{ctx.FromUserRole(), rpk.GetMsgId()})
		if handles == nil {
			log.Error("not found msg handler", "msgid", rpk.GetMsgId(), "from_uid", rpk.GetFromUID(), "from_svrtype", rpk.GetFromURole(), "to_uid", rpk.GetToUID(), "to_svrtype", rpk.GetToURole())
			return
		}
		for _, h := range handles {
			h(ctx)
		}
	case RoutePackMsgType_Request:
		handles := p.requestRoute.Get(RequestRouteKey{rpk.GetFromURole(), rpk.GetMsgId()})
		if handles == nil {
			log.Error("not found msg handler", "msgid", rpk.GetMsgId(), "from_uid", rpk.GetFromUID(), "from_svrtype", rpk.GetFromURole(), "to_uid", rpk.GetToUID(), "to_svrtype", rpk.GetToURole())
			return
		}
		for _, h := range handles {
			h(ctx)
		}
	case RoutePackMsgType_Response:
		cbinfo := p.PopRespCallback(ResponseRouteKey{ctx.FromUserID(), rpk.GetSYN()})
		if cbinfo == nil {
			return
		}
		if cbinfo.cbfun != nil {
			cbinfo.cbfun(false, rpk)
		}
	default:
	}
}
