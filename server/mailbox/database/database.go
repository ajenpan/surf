package database

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"gamemail/conf"
	"gamemail/log"
)

func CreateMysqlClient(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableNestedTransaction: true, //关闭嵌套事务
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: log.NewGormLogrus(),
	})
}

var WPropsDB *gorm.DB

func InitDB(c *conf.Config) error {
	var err error

	if WPropsDB, err = CreateMysqlClient(c.WPropsDBDSN); err != nil {
		return err
	}

	return err
}

var Rds *redis.Client

func LatestMailID() uint32 {
	return 0
}

func UserRecvLatestMailID(uid uint32) uint32 {
	return 0
}

func StoreNewMail() (uint32, error) {
	return 0, nil
}
