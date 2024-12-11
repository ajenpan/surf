package core

import (
	"context"
	"fmt"
	"reflect"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/ajenpan/surf/core/marshal"
)

func NewRdsClient(dsn string) *redis.Client {
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		panic(err)
	}
	rds := redis.NewClient(opt)
	if err := rds.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return rds
}

func NewMysqlClient(dsn string) *gorm.DB {
	dbc, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableNestedTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	return dbc
}

func HandleFunc[T proto.Message](s *Surf, f func(Context, T)) {
	var v T
	msgid := GetMsgId(v)
	msgname := string(v.ProtoReflect().Descriptor().Name())
	s.HandleFuncs(msgid, FuncToHandle(f))
	log.Info("HandleRequest", "msgid", msgid, "msgname", msgname)
}

func GetMsgIdFromFunc[T proto.Message](f func(Context, T)) uint32 {
	var v T
	return GetMsgId(v)
}

func FuncToHandle[T proto.Message](fn func(Context, T)) HandlerFunc {
	var v T
	msgType := reflect.TypeOf(v).Elem()
	return func(ctx Context) {
		msg := reflect.New(msgType).Interface().(T)
		pkt := ctx.Packet()

		marshaler := marshal.NewMarshaler(pkt.GetMarshalType())
		if marshaler == nil {
			err := fmt.Errorf("marshaler not found")
			log.Error("err", "err", err)
			ctx.Response(nil, err)
			return
		}

		err := marshaler.Unmarshal(pkt.GetBody(), msg)
		if err != nil {
			log.Error("Unmarshal ", "err", err)
			ctx.Response(nil, err)
			return
		}

		fn(ctx, msg)
	}
}

func GetMsgId(msg proto.Message) uint32 {
	md := msg.ProtoReflect().Descriptor()
	return GetMsgIDFromDesc(md)
}

func GetMsgIDFromDesc(md protoreflect.MessageDescriptor) uint32 {
	msgDesc := md.Enums().ByName("MSGID")
	if msgDesc == nil {
		return 0
	}
	idDesc := msgDesc.Values().ByName("ID")
	if idDesc == nil {
		return 0
	}
	return uint32(idDesc.Number())
}

func Assert(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func SendRequestToNode[RespT any](surf *Surf, ntype uint16, nid uint32, msg proto.Message, fn func(result *ResponseResult, resp *RespT)) error {
	var resp *RespT = new(RespT)
	return surf.SendRequestToNode(ntype, nid, msg, func(result *ResponseResult, pk *RoutePacket) {
		if fn == nil {
			return
		}
		if !result.Ok() {
			fn(result, nil)
			return
		}
		marshaler := marshal.NewMarshaler(pk.GetMarshalType())
		err := marshaler.Unmarshal(pk.Body, resp)
		if err != nil {
			log.Error(err.Error())
			return
		}
		fn(result, resp)
	})
}
