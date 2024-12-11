package core

import (
	"fmt"
	"sync"
	"time"
)

type ResponseResult struct {
	Timeout bool
	Errcode int16
}

func (r *ResponseResult) Ok() bool {
	return !r.Timeout && r.Errcode == 0
}

func (r *ResponseResult) Failed() bool {
	return r.Timeout || r.Errcode != 0
}

func (r *ResponseResult) String() string {
	return fmt.Sprintf("timeout:%v,errcode:%v", r.Timeout, r.Errcode)
}

var responseOK = &ResponseResult{}

type RequestCallbackFunc func(result *ResponseResult, pk *RoutePacket)

type RequestCallbackCache struct {
	cbfun   RequestCallbackFunc
	timeout *time.Timer
}

type RequestRouteKey = uint32

type ResponseRouteKey struct {
	NId uint32
	SYN uint32
}

func NewPacketRouteCaller() *PacketRouteCaller {
	return &PacketRouteCaller{
		handlers: NewHandlerRoute[RequestRouteKey](),
	}
}

type PacketRouteCaller struct {
	respWatier sync.Map
	handlers   *HandlerRoute[RequestRouteKey]
}

func (s *PacketRouteCaller) PushRespCallback(key ResponseRouteKey, timeoutsec uint32, cb RequestCallbackFunc) error {
	var timeout *time.Timer = nil

	if timeoutsec < 1 {
		timeoutsec = 1
	}

	timeout = time.AfterFunc(time.Duration(timeoutsec)*time.Second, func() {
		info := s.PopRespCallback(key)
		if info != nil && info.cbfun != nil {
			info.cbfun(&ResponseResult{true, 0}, nil)
		}
	})

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
		fallthrough
	case RoutePackMsgType_Request:
		handles := p.handlers.Get(rpk.GetMsgId())
		if handles == nil {
			log.Error("not found msg handler", "msgid", rpk.GetMsgId(), "from_uid", rpk.GetFromUId(), "from_svrtype", rpk.GetFromURole(), "to_uid", rpk.GetToUId(), "to_svrtype", rpk.GetToURole())
			return
		}
		for _, h := range handles {
			h(ctx)
		}
	case RoutePackMsgType_Response:
		cbinfo := p.PopRespCallback(ResponseRouteKey{ctx.FromUId(), rpk.GetSYN()})
		if cbinfo == nil {
			log.Error("not found resp handler", "msgid", rpk.GetMsgId(), "from_uid", rpk.GetFromUId(), "from_svrtype", rpk.GetFromURole(), "to_uid", rpk.GetToUId(), "to_svrtype", rpk.GetToURole())
			return
		}
		if cbinfo.cbfun != nil {
			cbinfo.cbfun(&ResponseResult{false, rpk.GetErrCode()}, rpk)
		}
	default:
		log.Error("unknow msg type", "msgid", rpk.GetMsgId(), "from_uid", rpk.GetFromUId(), "from_svrtype", rpk.GetFromURole(), "to_uid", rpk.GetToUId(), "to_svrtype", rpk.GetToURole())
	}
}
