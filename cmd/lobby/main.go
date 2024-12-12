package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/server/lobby"
)

var (
	Version   string = "unknown"
	GitCommit string = "unknown"
	BuildAt   string = "unknown"
)

func main() {
	app := cli.NewApp()
	app.Version = Version
	app.Name = lobby.NodeName()
	app.Action = RealMain
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func RealMain(c *cli.Context) error {
	h := lobby.New()

	opts := &core.ServerInfo{
		Svr:                h,
		OnClientDisconnect: h.OnClientDisconnect,
		OnClientConnect:    h.OnClientConnect,
	}

	conf := &core.NodeConf{
		SurfConf: core.SurfConfig{
			HttpListenAddr:    ":10300",
			WsListenAddr:      ":10301",
			PublicKeyFilePath: "http://myali01:9999/publickey",
			// EtcdConf:          &registry.EtcdConfig{},
		},
	}
	ninfo := &auth.NodeInfo{
		NId:   10300,
		NName: lobby.NodeName(),
		NType: lobby.NodeType(),
	}
	surf, err := core.NewSurf(ninfo, conf, opts)
	if err != nil {
		return err
	}
	defer surf.Close()
	err = surf.Run()
	return err
}
