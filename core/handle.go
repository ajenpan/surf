package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/auth"
	xerr "github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
)

func RequestFuncToConnHandler[ReqT, RespT proto.Message](fn func(ctx context.Context, in ReqT, out RespT) error) ConnHandler {
	var _req ReqT
	var _resp RespT

	var reqType = reflect.TypeOf(_req).Elem()
	var respType = reflect.TypeOf(_resp).Elem()

	respMsgId := GetMsgId(_resp)

	return func(conn network.Conn, rpk *RoutePacket) {
		var err error
		reqBody := rpk.Body()
		marshaltype := rpk.MarshalType()

		req := reflect.New(reqType).Interface().(ReqT)
		resp := reflect.New(respType).Interface().(RespT)

		marshaler := marshal.NewMarshaler(marshaltype)
		if marshaler == nil {
			err := fmt.Errorf("marshaler not found")
			log.Error("err", "err", err)
			return
		}

		err = marshaler.Unmarshal(reqBody, req)
		if err != nil {
			return
		}

		fromUser := &auth.UserInfo{
			UId:   rpk.FromUId(),
			URole: rpk.FromURole(),
		}

		ctx := CtxWithUser(context.Background(), fromUser)
		ctx = CtxWithConnId(ctx, conn.ConnId())

		err = fn(ctx, req, resp)

		if err != nil {
			errcode := int16(-1)
			if verr, ok := err.(*xerr.Error); ok && verr != nil {
				errcode = int16(verr.Code)
			}
			rpk.SetErrCode(errcode)
		}

		respBody, err := marshaler.Marshal(resp)

		if err != nil {
			log.Error("resp Marshal err", "what", err.Error())
		}

		log.Debug("handler", "reqname", req.ProtoReflect().Descriptor().Name(), "req", req, "resp", resp, "err", err)

		rpk.SetMsgId(respMsgId)
		rpk.SetMsgType(RoutePackMsgType_Response)
		rpk.SetFromUId(rpk.ToUId())
		rpk.SetFromURole(rpk.ToURole())
		rpk.SetToUId(fromUser.UId)
		rpk.SetToURole(fromUser.URole)
		rpk.SetBody(respBody)

		err = conn.Send(rpk.ToHVPacket())

		if err != nil {
			log.Error("resp Send err", "what", err.Error())
		}
	}
}

func RequestFuncToHttpHandler[ReqT, RespT proto.Message](fn func(ctx context.Context, in ReqT, out RespT) error) http.HandlerFunc {
	var _req ReqT
	var _resp RespT

	var reqType = reflect.TypeOf(_req).Elem()
	var respType = reflect.TypeOf(_resp).Elem()

	return func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		req := reflect.New(reqType).Interface().(ReqT)
		resp := reflect.New(respType).Interface().(RespT)

		marshaltype := marshal.NameToId(r.Header.Get("Content-Type"))

		marshaler := marshal.NewMarshaler(marshaltype)
		if marshaler == nil {
			err := fmt.Errorf("marshaler not found")
			log.Error("err", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		err = marshaler.Unmarshal(reqBody, req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		ctx := CtxWithConnId(r.Context(), "http-non-connid")
		err = fn(ctx, req, resp)
		if err != nil {
			errcode := (-1)
			if verr, ok := err.(*xerr.Error); ok && verr != nil {
				errcode = int(verr.Code)
			}
			w.Header().Set("errcode", strconv.Itoa(errcode))
			w.Header().Set("errmsg", err.Error())
		}
		respBody, err := marshaler.Marshal(resp)
		if err != nil {
			return
		}
		w.Write(respBody)
	}
}

func AsyncFuncToConnHandler[ReqT proto.Message](fn func(ctx context.Context, in ReqT)) ConnHandler {
	var _req ReqT
	var reqType = reflect.TypeOf(_req).Elem()

	return func(conn network.Conn, rpk *RoutePacket) {
		var err error
		reqBody := rpk.Body()
		req := reflect.New(reqType).Interface().(ReqT)

		marshaltype := rpk.MarshalType()
		marshaler := marshal.NewMarshaler(marshaltype)
		if marshaler == nil {
			err := fmt.Errorf("marshaler not found")
			log.Error("err", "err", err)
			return
		}
		err = marshaler.Unmarshal(reqBody, req)
		if err != nil {
			return
		}
		fromUser := &auth.UserInfo{
			UId:   rpk.FromUId(),
			URole: rpk.FromURole(),
		}
		ctx := CtxWithUser(context.Background(), fromUser)
		ctx = CtxWithConnId(ctx, conn.ConnId())
		fn(ctx, req)
	}
}

func AsyncFuncToHttpHandler[ReqT proto.Message](fn func(ctx context.Context, in ReqT)) http.HandlerFunc {
	var _req ReqT
	var reqType = reflect.TypeOf(_req).Elem()
	return func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		req := reflect.New(reqType).Interface().(ReqT)
		marshaltype := marshal.NameToId(r.Header.Get("Content-Type"))
		marshaler := marshal.NewMarshaler(marshaltype)
		if marshaler == nil {
			err := fmt.Errorf("marshaler not found")
			log.Error("err", "err", err)
			return
		}
		err = marshaler.Unmarshal(reqBody, req)
		if err != nil {
			return
		}
		fn(r.Context(), req)
	}
}

func HandleRequestFromConn[ReqT, RespT proto.Message](surf *Surf, fn func(ctx context.Context, in ReqT, out RespT) error) {
	var _req ReqT
	reqMsgId := GetMsgId(_req)
	surf.HandleFuncs(reqMsgId, RequestFuncToConnHandler(fn))
}

func HandleAsyncFromConn[ReqT proto.Message](surf *Surf, fn func(ctx context.Context, in ReqT)) {
	var _req ReqT
	reqMsgId := GetMsgId(_req)
	surf.HandleFuncs(reqMsgId, AsyncFuncToConnHandler(fn))
}

func HandleRequestFromHttp[ReqT, RespT proto.Message](surf *Surf, pattern string, fn func(ctx context.Context, in ReqT, out RespT) error) {
	surf.httpMux.HandleFunc(pattern, RequestFuncToHttpHandler(fn))
}

func HandleAsyncFromHttp[ReqT proto.Message](surf *Surf, pattern string, fn func(ctx context.Context, in ReqT)) {
	surf.httpMux.HandleFunc(pattern, AsyncFuncToHttpHandler(fn))
}
