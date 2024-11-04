package uauth

import (
	"fmt"
	"net/http"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/utils/rsagen"
	"github.com/ajenpan/surf/server/uauth/database/cache"
)

const PrivateKeyFile = "private.pem"
const PublicKeyFile = "public.pem"

func ReadRSAKey() ([]byte, []byte, error) {
	privateRaw, err := os.ReadFile(PrivateKeyFile)
	if err != nil {
		privateKey, publicKey, err := rsagen.GenerateRsaPem(1024)
		if err != nil {
			return nil, nil, err
		}
		privateRaw = []byte(privateKey)
		os.WriteFile(PrivateKeyFile, []byte(privateKey), 0644)
		os.WriteFile(PublicKeyFile, []byte(publicKey), 0644)
	}
	publicRaw, err := os.ReadFile(PublicKeyFile)
	if err != nil {
		return nil, nil, err
	}
	return privateRaw, publicRaw, nil
}

func CreateMysqlClient(dsn string) *gorm.DB {
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

func Start(conf *Config) (func() error, error) {
	privateRaw, publicRaw, err := ReadRSAKey()
	if err != nil {
		panic(err)
	}

	pk, err := rsagen.ParseRsaPrivateKeyFromPem(privateRaw)
	if err != nil {
		return nil, err
	}

	h := NewAuth(AuthOptions{
		PK:        pk,
		PublicKey: publicRaw,
		DB:        CreateMysqlClient(conf.DBAddr),
		Cache:     cache.NewMemory(),
	})
	ct := h.CTByName()

	svr := &HttpSvr{
		Marshal: &marshal.JSONPb{
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
		Addr: conf.HttpListenAddr,
	}

	svr.ServerCallTable(ct)

	svr.Mux.HandleFunc("/publickey", func(w http.ResponseWriter, r *http.Request) {
		w.Write(publicRaw)
	})

	err = svr.Run()
	if err != nil {
		return nil, err
	}

	fmt.Println("start http server at:", conf.HttpListenAddr)
	closer := func() error {
		return svr.Stop()
	}
	return closer, nil
}
