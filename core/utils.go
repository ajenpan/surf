package core

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ajenpan/surf/core/marshal"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

func FuncToHandle[T proto.Message](f func(Context, T)) HandlerFunc {
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

		f(ctx, msg)
	}
}
