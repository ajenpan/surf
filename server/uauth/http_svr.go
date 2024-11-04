package uauth

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/utils/calltable"
	"google.golang.org/protobuf/proto"
)

type httpCallContext struct {
	w     http.ResponseWriter
	r     *http.Request
	uinfo *auth.UserInfo
}

func (ctx *httpCallContext) Response(msg proto.Message, err error) {
	type httpWrap struct {
		ErrCode int         `json:"errcode"`
		ErrMsg  string      `json:"errmsg"`
		Data    interface{} `json:"data"`
	}

	enc := json.NewEncoder(ctx.w)
	wrap := &httpWrap{Data: msg}

	if err != nil {
		wrap.ErrMsg = err.Error()

		if errs, ok := err.(*errors.Error); ok {
			wrap.ErrCode = int(errs.Code)
		} else {
			wrap.ErrCode = -1
		}
	}

	encerr := enc.Encode(wrap)

	if encerr != nil {
		ctx.w.WriteHeader(http.StatusInternalServerError)
		ctx.w.Write([]byte(encerr.Error()))
	}
}

func (ctx *httpCallContext) Caller() uint32 {
	return ctx.uinfo.UId
}

type HttpSvr struct {
	Addr    string
	Marshal marshal.Marshaler
	Mux     *http.ServeMux
	svr     *http.Server
}

func (s *HttpSvr) Run() error {
	s.svr = &http.Server{
		Addr:    s.Addr,
		Handler: s.Mux,
	}

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	go s.svr.Serve(ln)
	return nil
}

func (s *HttpSvr) Stop() error {
	return s.svr.Close()
}

func (s *HttpSvr) ServerCallTable(ct *calltable.CallTable[string]) {
	if s.Mux == nil {
		s.Mux = &http.ServeMux{}
	}
	ct.Range(func(key string, method *calltable.Method) bool {
		if !strings.HasPrefix(key, "/") {
			key = "/" + key
		}
		cb := s.WrapMethod(method)
		s.Mux.HandleFunc(key, cb)
		return true
	})
}

func (s *HttpSvr) HandleMethod(name string, method *calltable.Method) {
	s.Mux.HandleFunc(name, s.WrapMethod(method))
}

func (s *HttpSvr) WrapMethod(method *calltable.Method) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		req := method.GetRequest()
		defer method.PutRequest(req)

		if err := s.Marshal.Unmarshal(raw, req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		var ctx core.Context = &httpCallContext{
			w: w,
			r: r,
		}
		// here call method
		method.Call(ctx, req)
		// respArgs := method.Call(ctx, req)

		// if len(respArgs) != 2 {
		// 	return
		// }

		// var respErr error

		// if !respArgs[1].IsNil() {
		// 	respErr = respArgs[1].Interface().(error)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	w.Write([]byte(respErr.Error()))
		// 	return
		// }

		// if !respArgs[0].IsNil() {
		// 	respData, err := s.Marshal.Marshal(respArgs[0].Interface())
		// 	if err != nil {
		// 		w.WriteHeader(http.StatusInternalServerError)
		// 		w.Write([]byte(err.Error()))
		// 		return
		// 	}
		// 	w.Write(respData)
		// }
		// w.WriteHeader(http.StatusOK)
	}
}
