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

	_ "github.com/ajenpan/surf/game/niuniu"
	battleHandler "github.com/ajenpan/surf/server/battle/handler"
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

func RealMain(c *cli.Context) error {
	log.Default.SetOutput(os.Stdout)

	pk, err := LoadAuthPublicKey()
	if err != nil {
		panic(err)
	}

	h := battleHandler.New()
	CTByID, CTByName := calltable.ExtractMethodFromDesc(battleMsg.File_battle_proto.Messages(), h)

	uid := 10001
	uname := fmt.Sprintf("battle_%d", uid)
	uinfo := &auth.UserInfo{
		UId:   uint32(uid),
		UName: uname,
		URole: 6000,
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

	err = core.Init(core.Options{
		Server:            h,
		HttpListenAddr:    ":13300",
		WsListenAddr:      ":13301",
		CTById:            CTByID,
		CTByName:          CTByName,
		PublicKeyFilePath: "http://myali01:9999/publickey",
		GateAddrList: []string{
			"ws://localhost:13000",
		},
		GateToken: []byte(jwt),
		UInfo:     uinfo,
	})

	if err != nil {
		return err
	}

	err = core.Run()
	return err
}
