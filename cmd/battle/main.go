package main

import (
	"crypto/rsa"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/core/utils/rsagen"
	battleMsg "github.com/ajenpan/surf/msg/battle"
	"github.com/ajenpan/surf/server/battle"
	battleHandler "github.com/ajenpan/surf/server/battle/handler"

	// games
	"github.com/ajenpan/surf/game/guandan"
	"github.com/ajenpan/surf/game/niuniu"
)

var (
	Name      string = "battle"
	Version   string = "unknown"
	GitCommit string = "unknown"
	BuildAt   string = "unknown"
)

func main() {
	err := Run()
	if err != nil {
		fmt.Println(err)
	}
}

func Run() error {
	info := core.NewServerInfo()
	info.Name = Name
	info.Version = Version
	info.GitCommit = GitCommit
	info.BuildAt = BuildAt

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(info.LongVersion())
	}

	app := cli.NewApp()
	app.Version = Version
	app.Name = Name
	app.Action = RealMain

	err := app.Run(os.Args)
	return err
}

func LoadAuthPublicKey() (*rsa.PrivateKey, error) {
	publicRaw, err := os.ReadFile("private.pem")
	if err != nil {
		return nil, err
	}
	pk, err := rsagen.ParseRsaPrivateKeyFromPem(publicRaw)
	return pk, err
}

func RegGames() {
	battle.RegisterGame("guandan", guandan.NewLogic)
	battle.RegisterGame("niuniu", niuniu.NewLogic)
}

func RealMain(c *cli.Context) error {
	log.Default.SetOutput(os.Stdout)

	pk, err := LoadAuthPublicKey()
	if err != nil {
		panic(err)
	}

	h := battleHandler.New()
	CTByID := calltable.ExtractMethodFromDesc(battleMsg.File_battle_proto.Messages(), h)

	uid := 10001

	uinfo := &auth.UserInfo{
		UId:   uint32(uid),
		UName: fmt.Sprintf("%s_%d", h.ServerName(), uid),
		URole: uint32(h.ServerType()),
	}

	jwt, err := auth.GenerateToken(pk, uinfo, 2400*time.Hour)
	if err != nil {
		return err
	}
	testuid := uint32(rand.Int31n(300000) + 30000)
	testuser := &auth.UserInfo{
		UId:   testuid,
		UName: fmt.Sprintf("yk%d", testuid),
		URole: 1,
	}
	testjwt, err := auth.GenerateToken(pk, testuser, 24*time.Hour*999)
	if err != nil {
		return err
	}
	log.Info("testjwt:", testjwt)

	RegGames()

	opts := core.Options{
		Server:         h,
		HttpListenAddr: ":13300",
		WsListenAddr:   ":13301",
		RouteCallTable: CTByID,
		// PublicKeyFilePath: "http://myali01:9999/publickey",
		PublicKeyFilePath: "file://./public.pem",
		GateAddrList: []string{
			"ws://localhost:13000",
		},
		GateToken: []byte(jwt),
		UInfo:     uinfo,
	}
	surf, err := core.NewSurf(opts)
	if err != nil {
		return err
	}
	err = surf.Run()
	return err
}
