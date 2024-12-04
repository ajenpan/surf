package main

import (
	"crypto/rsa"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/utils/rsagen"
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
	pk, err := LoadAuthPublicKey()
	if err != nil {
		panic(err)
	}

	h := battleHandler.New()
	// calltable := calltable.ExtractMethodFromDesc(msgBattle.File_battle_proto.Messages(), h)

	uid := 10001
	uinfo := &auth.UserInfo{
		UId:   uint32(uid),
		UName: fmt.Sprintf("%s_%d", h.ServerName(), uid),
		URole: uint16(h.ServerType()),
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
	slog.Info("testjwt", "jwt", testjwt)

	RegGames()

	opts := core.Options{
		HttpListenAddr: ":13300",
		WsListenAddr:   ":13301",
		// PublicKeyFilePath: "http://myali01:9999/publickey",
		PublicKeyFilePath: "file://./public.pem",
		GateAddrList: []string{
			"ws://localhost:13000",
		},
		GateToken:          []byte(jwt),
		OnClientDisconnect: h.OnPlayerDisConn,
	}
	surf, err := core.NewSurf2(opts)
	if err != nil {
		return err
	}

	err = surf.Run()
	return err
}
