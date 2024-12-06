package calltable2

type Method interface {
	Call(ctx interface{}, req interface{})
	NewMsg() interface{}
}

type method[CtxT any, ReqT any] struct {
	Func func(ctx CtxT, req *ReqT)
}

func (m *method[CtxT, ReqT]) Call(ctx interface{}, req interface{}) {
	m.Func(ctx.(CtxT), req.(*ReqT))
}

func (m *method[CtxT, ReqT]) NewMsg() interface{} {
	var req *ReqT = new(ReqT)
	return req
}

func FromFunc[CtxT any, ReqT any](f func(ctx CtxT, req *ReqT)) Method {
	return &method[CtxT, ReqT]{Func: f}
}
