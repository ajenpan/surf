package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
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
	app := cli.NewApp()
	app.Version = Version
	app.Name = Name
	app.Action = RealMain
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func RealMain(c *cli.Context) error {
	h := battleHandler.New()
	// calltable := calltable.ExtractMethodFromDesc(msgBattle.File_battle_proto.Messages(), h)

	// RegGames
	battle.RegisterGame("guandan", guandan.NewLogic)
	battle.RegisterGame("niuniu", niuniu.NewLogic)

	opts := &core.ServerInfo{
		Svr:                h,
		OnClientDisconnect: h.OnPlayerDisConn,
	}

	conf := &core.NodeConf{
		SurfConf: core.SurfConfig{
			HttpListenAddr:    ":13300",
			WsListenAddr:      ":13301",
			PublicKeyFilePath: "http://myali01:9999/publickey",
		},
	}
	ninfo := &auth.NodeInfo{
		NId:   10001,
		NName: Name,
		NType: battleHandler.NodeType_Battle,
	}
	surf, err := core.NewSurf(ninfo, conf, opts)
	if err != nil {
		return err
	}

	err = surf.Run()
	return err
}
