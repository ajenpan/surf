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

type RequestRouteKey = uint32

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

func (p *PacketRouteCaller) Call(ctx *ConnContext) {
	rpk := ctx.ReqPacket

	if rpk.GetSubType() != 0 {
		log.Error("recv err packet subtype", "subtype", rpk.GetSubType())
		return
	}

	switch rpk.GetMsgType() {
	case RoutePackMsgType_Async:
		h := p.ayncRoute.Get(AyncRouteKey{ctx.Conn.UserRole(), rpk.GetMsgId()})
		if h == nil {
			log.Error("not found msg handler", "msgid", rpk.GetMsgId(), "from_uid", rpk.GetFromUID(), "from_svrtype", rpk.GetFromURole(), "to_uid", rpk.GetToUID(), "to_svrtype", rpk.GetToURole())
			return
		}
		h(ctx)
	case RoutePackMsgType_Request:
		h := p.requestRoute.Get(rpk.GetMsgId())
		if h == nil {
			log.Error("not found msg handler", "msgid", rpk.GetMsgId(), "from_uid", rpk.GetFromUID(), "from_svrtype", rpk.GetFromURole(), "to_uid", rpk.GetToUID(), "to_svrtype", rpk.GetToURole())
			return
		}
		h(ctx)
		// marshaler := marshal.NewMarshalerById(rpk.GetMarshalType())
		// if marshaler == nil {
		// 	log.Error("invalid marshaler type", "type", rpk.GetMarshalType())
		// 	//todo send error packet
		// 	return
		// }
		// req := method.NewRequest()
		// err := marshaler.Unmarshal(rpk.GetBody(), req)
		// if err != nil {
		// 	log.Error("unmarshal request body failed", "err", err)
		// 	//todo send error packet
		// 	return
		// }
		// method.Call(p.Handler, ctx, req)
	case RoutePackMsgType_Response:
		cbinfo := p.PopRespCallback(ResponseRouteKey{ctx.Conn.UserID(), rpk.GetSYN()})
		if cbinfo == nil {
			return
		}
		if cbinfo.cbfun != nil {
			cbinfo.cbfun(false, rpk)
		}
	default:
	}
}
