package main

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/ajenpan/surf/core/network"

	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/core/utils/marshal"

	"github.com/ajenpan/surf/server/auth/common"
	"github.com/ajenpan/surf/server/auth/handler"
	"github.com/ajenpan/surf/server/auth/store/cache"
	"github.com/ajenpan/surf/server/auth/store/models"

	utilSignal "github.com/ajenpan/surf/core/utils/signal"

	"github.com/ajenpan/surf/core/utils/rsagen"

	log "github.com/ajenpan/surf/core/log"
)

var Version string = "unknown"
var GitCommit string = "unknown"
var BuildAt string = "unknown"
var BuildBy string = "unknown"
var Name string = "auth"

var ConfigPath string = ""
var ListenAddr string = ""
var PrintConf bool = false

const PrivateKeyFile = "private.pem"
const PublicKeyFile = "public.pem"

func ReadRSAKey() ([]byte, []byte, error) {
	privateRaw, err := os.ReadFile(PrivateKeyFile)
	if err != nil {
		privateKey, publicKey, err := rsagen.GenerateRsaPem(512)
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

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("project:", Name)
		fmt.Println("version:", Version)
		fmt.Println("git commit:", GitCommit)
		fmt.Println("build at:", BuildAt)
		fmt.Println("build by:", BuildBy)
	}

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       "config.yaml",
			Destination: &ConfigPath,
		}, &cli.StringFlag{
			Name:        "listen",
			Aliases:     []string{"l"},
			Value:       ":30020",
			Destination: &ListenAddr,
		}, &cli.BoolFlag{
			Name:        "print-config",
			Destination: &PrintConf,
			Hidden:      true,
		},
	}

	app.Action = RealMain

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
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

func CreateSQLiteClient(dsn string) *gorm.DB {
	dbc, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dbc.AutoMigrate(
		&models.Users{},
	)

	if err != nil {
		panic(err)
	}
	return dbc
}

func RealMain(c *cli.Context) error {
	privateRaw, publicRaw, err := ReadRSAKey()
	if err != nil {
		panic(err)
	}
	pk, err := rsagen.ParseRsaPrivateKeyFromPem(privateRaw)
	if err != nil {
		return err
	}
	publicKey, err := rsagen.ParseRsaPublicKeyFromPem(publicRaw)
	if err != nil {
		return err
	}

	h := handler.NewAuth(handler.AuthOptions{
		PK:        pk,
		PublicKey: publicRaw,
		DB:        CreateSQLiteClient("auth.db"),
		// DB:    CreateMysqlClient("root:123456@tcp(test122:13306)/auth?charset=utf8mb4&parseTime=True&loc=Local"),
		Cache: cache.NewMemory(),
	})

	AuthWrapper := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authorstr := r.Header.Get("Authorization")
			authorstr = strings.TrimPrefix(authorstr, "Bearer ")
			if authorstr == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, err := common.VerifyToken(publicKey, authorstr)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			h(w, r)
		}
	}

	mux := http.NewServeMux()

	NewMethod := func(f interface{}) *calltable.Method {
		refv := reflect.ValueOf(f)
		if refv.Kind() != reflect.Func {
			return nil
		}
		ret := &calltable.Method{
			Func:         refv,
			RequestType:  refv.Type().In(1).Elem(),
			ResponseType: refv.Type().Out(0).Elem(),
		}
		ret.InitPool()
		return ret
	}

	svr := &network.HttpSvr{
		Mux: mux,
		Marshal: &marshal.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseEnumNumbers: true,
				UseProtoNames:  true,
			},
		},
		Addr: ListenAddr,
	}

	mux.Handle("/auth/Login", svr.WrapMethod(NewMethod(h.Login)))
	mux.Handle("/auth/Register", svr.WrapMethod(NewMethod(h.Register)))
	mux.Handle("/auth/UserInfo", AuthWrapper(svr.WrapMethod(NewMethod(h.UserInfo))))
	mux.Handle("/auth/RefreshToken", AuthWrapper(svr.WrapMethod(NewMethod(h.RefreshToken))))

	go svr.Start()
	defer svr.Stop()

	fmt.Println("start http server at:", ListenAddr)

	signal := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", signal.String())
	return nil
}
