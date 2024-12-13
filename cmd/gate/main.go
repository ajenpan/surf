package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	gate "github.com/ajenpan/surf/server/gate"
)

var (
	Version    string = "unknow"
	GitCommit  string = "unknow"
	BuildAt    string = "unknow"
	BuildBy    string = runtime.Version()
	RunnningOS string = runtime.GOOS + "/" + runtime.GOARCH
)

func longVersion() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintln(buf, "project:", core.NodeName_Gate)
	fmt.Fprintln(buf, "version:", Version)
	fmt.Fprintln(buf, "git commit:", GitCommit)
	fmt.Fprintln(buf, "build at:", BuildAt)
	fmt.Fprintln(buf, "build by:", BuildBy)
	fmt.Fprintln(buf, "running OS/Arch:", RunnningOS)
	return buf.String()
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(longVersion())
	}
	app := cli.NewApp()
	app.Version = Version
	app.Name = core.NodeName_Gate
	app.Action = RealMain
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func RealMain(c *cli.Context) error {
	opts := &core.ServerInfo{
		Svr: gate.New(),
	}
	conf := &core.NodeConf{
		SurfConf: core.SurfConfig{
			PublicKeyFilePath: "http://myali01:9999/publickey",
			// EtcdConf:          &registry.EtcdConfig{Endpoints: []string{"test122:2379"}},
		},
	}
	ninfo := &auth.NodeInfo{
		NId:   10000,
		NName: core.NodeName_Gate,
		NType: core.NodeType_Gate,
	}
	surf, err := core.NewSurf(ninfo, conf, opts)
	if err != nil {
		return err
	}
	defer surf.Close()
	err = surf.Run()
	return err
}
