package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/game"
	battleHandler "github.com/ajenpan/surf/server/battle"

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

	// RegGames
	game.RegisterGame("guandan", guandan.NewLogic)
	game.RegisterGame("niuniu", niuniu.NewLogic)

	opts := &core.ServerInfo{
		Svr:                h,
		OnClientDisconnect: h.OnPlayerDisConn,
	}

	conf := &core.NodeConf{
		SurfConf: core.SurfConfig{
			HttpListenAddr:    ":10200",
			WsListenAddr:      ":10201",
			PublicKeyFilePath: "http://myali01:9999/publickey",
			GateAddrList: []string{
				"ws://localhost:10101",
			},
		},
	}
	ninfo := &auth.NodeInfo{
		NId:   10201,
		NName: Name,
		NType: battleHandler.NodeType_Battle,
	}
	surf, err := core.NewSurf(ninfo, conf, opts)
	if err != nil {
		return err
	}
	defer surf.Close()
	err = surf.Run()
	return err
}
