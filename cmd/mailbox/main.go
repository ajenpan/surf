package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ajenpan/surf/core/log"
	proto "github.com/ajenpan/surf/msg/mailbox"
	"github.com/ajenpan/surf/server/mailbox"
)

var ConfigPath string = ""
var ListenAddr string = ""
var PrintConf bool = false

var GHandler *mailbox.Handler

var Version string = ""
var GitCommit string = ""
var BuildAt string = ""

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("version:", Version)
		fmt.Println("git commit:", GitCommit)
		fmt.Println("build at:", BuildAt)
	}

	app := &cli.App{
		Name: "gamemail",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "./conf/config.yaml",
				Destination: &ConfigPath,
			}, &cli.StringFlag{
				Name:        "listen",
				Aliases:     []string{"l"},
				Value:       ":9020",
				Destination: &ListenAddr,
			}, &cli.BoolFlag{
				Name:        "print-config",
				Destination: &PrintConf,
				Hidden:      true,
			},
		},
		Version: Version,
	}

	app.Action = func(c *cli.Context) error {
		var err error

		if mailbox.DefaultConf, err = mailbox.ConfInit(ConfigPath, PrintConf); err != nil {
			log.Error(err)
			return err
		}

		if GHandler = mailbox.NewHandler(mailbox.DefaultConf); GHandler == nil {
			err := fmt.Errorf("create handler failed")
			log.Panic(err)
			return err
		}

		if err := httpsvr(); err != nil {
			log.Error(err)
		}
		log.Info("on exsiting...")
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

type HandlerFunc = func(rw http.ResponseWriter, r *http.Request)
type HandlerFuncWarp = func(HandlerFunc) HandlerFunc

func MultiWarp(funcs ...HandlerFuncWarp) HandlerFuncWarp {
	return func(f func(rw http.ResponseWriter, r *http.Request)) func(rw http.ResponseWriter, r *http.Request) {
		for i := len(funcs) - 1; i >= 0; i-- {
			f = funcs[i](f)
		}
		return f
	}
}

func HttpHeaderForward(f func(rw http.ResponseWriter, r *http.Request)) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		// for nginx
		values, has := r.Header["X-Real-Ip"]
		if has && len(values) > 0 {
			r = r.WithContext(context.WithValue(r.Context(), mailbox.CtxXRealIp, values[0]))
		} else {
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
				r = r.WithContext(context.WithValue(r.Context(), mailbox.CtxXRealIp, r.RemoteAddr))
			} else {
				r = r.WithContext(context.WithValue(r.Context(), mailbox.CtxXRealIp, host))
			}
		}
		f(rw, r)
	}
}

var AllowList = map[string]bool{
	"/MailBox/Announcement": true,
}

func parserAdminBearer(r *http.Request) (context.Context, error) {
	bearerToekn := r.Header.Get("Authorization")
	bearerToekn = strings.TrimPrefix(bearerToekn, "Bearer ")
	if len(bearerToekn) < 1 {
		return nil, fmt.Errorf("token is required")
	}
	claims, err := mailbox.VerifyAdminTokenWithClaims(bearerToekn)
	if err != nil {
		return nil, fmt.Errorf("token verify failed: %v", err)
	}
	newCtx := context.WithValue(r.Context(), mailbox.CtxClaims, claims)
	newCtx = context.WithValue(newCtx, mailbox.CtxAdminUID, claims["uid"])
	newCtx = context.WithValue(newCtx, mailbox.CtxCallerRole, "admin")
	return newCtx, nil
}

func httpsvr() error {
	ct := mailbox.ParseRpcMethod(proto.File_mailbox_proto.Services(), GHandler)
	ct.Range(func(key string, method *mailbox.MessageMethod) bool {
		key = "/" + key
		fmt.Println("handle path:", key)

		handleCall := func(rw http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error(err)
					log.Error(string(debug.Stack()))
				}
			}()

			mapRaw := make(map[string]interface{})
			mapRaw["code"] = 1

			// 这里是否要限制一下 body 太大的情况, 如果客户端直接发了一个大于内存的数据, 会发生什么
			defer r.Body.Close()
			raw, err := ioutil.ReadAll(r.Body)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(err.Error()))
				return
			}

			req := reflect.New(method.Req).Interface().(protoreflect.ProtoMessage)
			resp := reflect.New(method.Resp).Interface().(protoreflect.ProtoMessage)

			if err = protojson.Unmarshal([]byte(raw), req); err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(err.Error()))
				return
			}

			callResult := method.Method.Func.Call([]reflect.Value{reflect.ValueOf(GHandler), reflect.ValueOf(r.Context()), reflect.ValueOf(req), reflect.ValueOf(resp)})

			if !callResult[0].IsNil() {
				err = callResult[0].Interface().(error)
			}

			if err != nil {
				log.Error(err)
				mapRaw["code"] = -1
				mapRaw["message"] = err.Error()
			} else {
				raw, err := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}.Marshal(resp)
				if err == nil {
					mapRaw["data"] = json.RawMessage(raw)
					mapRaw["code"] = 0
					mapRaw["message"] = "ok"
				} else {
					log.Error(err)
					mapRaw["code"] = -1
					mapRaw["message"] = err.Error()
				}
			}

			respRaw, err := json.Marshal(mapRaw)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(err.Error()))
				return
			}

			rw.Header().Set("Content-Type", "application/json; charset=utf-8")

			_, err = rw.Write(respRaw)

			//the err after log
			if err != nil {
				log.Error("write resp err:", err)
			}
		}

		h := MultiWarp(HttpHeaderForward)(handleCall)
		http.HandleFunc(key, h)
		return true
	})

	fmt.Println("http listen at ", ListenAddr)

	return http.ListenAndServe(ListenAddr, nil)
}
