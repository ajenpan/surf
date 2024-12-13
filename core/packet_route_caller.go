package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/ajenpan/surf/core/network"
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

func (p *PacketRouteCaller) Call(ctx network.Conn, rpk *RoutePacket) {
	if rpk.SubType() != 0 {
		log.Error("recv err packet subtype", "subtype", rpk.SubType())
		return
	}
	switch rpk.MsgType() {
	case RoutePackMsgType_Async:
		fallthrough
	case RoutePackMsgType_Request:
		h := p.handlers.Get(rpk.MsgId())
		if h == nil {
			log.Error("not found msg handler", "msgid", rpk.MsgId(), "from_uid", rpk.FromUId(), "from_svrtype", rpk.FromURole(), "to_uid", rpk.ToUId(), "to_svrtype", rpk.ToURole())
			return
		}
		h(ctx, rpk)
	case RoutePackMsgType_Response:
		cbinfo := p.PopRespCallback(ResponseRouteKey{rpk.FromUId(), rpk.SYN()})
		if cbinfo == nil {
			log.Error("not found resp handler", "msgid", rpk.MsgId(), "from_uid", rpk.FromUId(), "from_svrtype", rpk.FromURole(), "to_uid", rpk.ToUId(), "to_svrtype", rpk.ToURole())
			return
		}
		if cbinfo.cbfun != nil {
			cbinfo.cbfun(&ResponseResult{false, rpk.ErrCode()}, rpk)
		}
	default:
		log.Error("unknow msg type", "msgid", rpk.MsgId(), "from_uid", rpk.FromUId(), "from_svrtype", rpk.FromURole(), "to_uid", rpk.ToUId(), "to_svrtype", rpk.ToURole())
	}
}
