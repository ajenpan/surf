package core

import (
	"context"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/ajenpan/surf/core/marshal"
)

// dsn='redis://<user>:<password>@<host>:<port>/<db_number>'
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

func Request[RespT any](surf *Surf, ntype uint16, nid uint32, msg proto.Message, fn func(result *ResponseResult, resp *RespT)) error {
	var resp *RespT = new(RespT)
	return surf.SendRequestToNode(ntype, nid, msg, func(result *ResponseResult, pk *RoutePacket) {
		if fn == nil {
			return
		}
		if !result.Ok() {
			fn(result, nil)
			return
		}
		marshaler := marshal.NewMarshaler(pk.MarshalType())
		err := marshaler.Unmarshal(pk.body, resp)
		if err != nil {
			log.Error(err.Error())
			return
		}
		fn(result, resp)
	})
}
