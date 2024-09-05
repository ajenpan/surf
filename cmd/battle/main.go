package main

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/core/utils/rsagen"
	innerbattleMsg "github.com/ajenpan/surf/msg/innerproto/battle"
	battleMsg "github.com/ajenpan/surf/msg/openproto/battle"

	_ "github.com/ajenpan/surf/game/niuniu"
	battleHandler "github.com/ajenpan/surf/server/battle/handler"
)

var (
	Name       string = "unknown"
	Version    string = "unknown"
	GitCommit  string = "unknown"
	BuildAt    string = "unknown"
	BuildBy    string = runtime.Version()
	RunnningOS string = runtime.GOOS + "/" + runtime.GOARCH
)

func longVersion() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintln(buf, "project:", Name)
	fmt.Fprintln(buf, "version:", Version)
	fmt.Fprintln(buf, "git commit:", GitCommit)
	fmt.Fprintln(buf, "build at:", BuildAt)
	fmt.Fprintln(buf, "build by:", BuildBy)
	fmt.Fprintln(buf, "running OS/Arch:", RunnningOS)
	return buf.String()
}

func main() {
	err := Run()
	if err != nil {
		fmt.Println(err)
	}
}

func Run() error {
	Name = "battle"

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(longVersion())
	}

	app := cli.NewApp()
	app.Version = Version
	app.Name = Name
	app.Action = RealMain

	err := app.Run(os.Args)
	return err
}

var listenAt string = ":12345"

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
	ct := calltable.ExtractMethodByMsgID(battleMsg.File_battle_proto.Messages(), h)
	ct.Merge(calltable.ExtractMethodByMsgID(innerbattleMsg.File_battle_inner_proto.Messages(), h), false)

	uinfo := &auth.UserInfo{
		UId:   20001,
		UName: "battle",
		URole: 1,
	}
	jwt, err := auth.GenerateToken(pk, uinfo, 2400*time.Hour)
	if err != nil {
		return err
	}

	err = core.Init(core.Options{
		Server:         h,
		TcpListenAddr:  listenAt,
		HttpListenAddr: ":18080",
		CTById:         ct,
		PublicKey:      &pk.PublicKey,
		GateAddrList: []string{
			"ws://localhost:9998",
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
