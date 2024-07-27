package network

import (
	"io"
	"net/http"
	"strings"

	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/core/utils/marshal"
)

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
	return s.svr.ListenAndServe()
}

func (s *HttpSvr) Stop() error {
	return s.svr.Close()
}

func (s *HttpSvr) ServerCallTable(ct *calltable.CallTable[string]) {
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

		// here call method
		respArgs := method.Call(r.Context(), req)

		if len(respArgs) != 2 {
			return
		}

		var respErr error

		if !respArgs[1].IsNil() {
			respErr = respArgs[1].Interface().(error)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(respErr.Error()))
			return
		}

		if !respArgs[0].IsNil() {
			respData, err := s.Marshal.Marshal(respArgs[0].Interface())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.Write(respData)
		}
		w.WriteHeader(http.StatusOK)
	}
}
