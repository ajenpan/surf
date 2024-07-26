package main

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	utilSignal "github.com/ajenpan/surf/core/utils/signal"

	"github.com/ajenpan/surf/core/utils/rsagen"

	log "github.com/ajenpan/surf/core/log"
)

var Version string = "unknown"
var GitCommit string = "unknown"
var BuildAt string = "unknown"
var BuildBy string = "unknown"
var Name string = "allinone"

var ConfigPath string = ""
var ListenAddr string = ""
var PrintConf bool = false

func ReadRSAKey() (*rsa.PrivateKey, error) {
	const privateFile = "private.pem"
	const publicFile = "public.pem"

	raw, err := os.ReadFile(privateFile)
	if err != nil {
		privateKey, publicKey, err := rsagen.GenerateRsaPem(2048)
		if err != nil {
			return nil, err
		}
		raw = []byte(privateKey)
		os.WriteFile(privateFile, []byte(privateKey), 0644)
		os.WriteFile(publicFile, []byte(publicKey), 0644)
	}
	return rsagen.ParseRsaPrivateKeyFromPem(raw)
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
			Value:       ":10110",
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

func createMysqlClient(dsn string) *gorm.DB {
	dbc, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableNestedTransaction: true, //关闭嵌套事务
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	return dbc
}

var PK *rsa.PrivateKey

func RealMain(c *cli.Context) error {
	var err error
	PK, err = ReadRSAKey()
	if err != nil {
		panic(err)
	}

	signal := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", signal.String())
	return nil
}
