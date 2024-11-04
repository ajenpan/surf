package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core/log"
	utilSignal "github.com/ajenpan/surf/core/utils/signal"
	"github.com/ajenpan/surf/server/uauth"
)

var Version string = "unknown"
var GitCommit string = "unknown"
var BuildAt string = "unknown"
var BuildBy string = "unknown"
var Name string = "auth"

var ConfigPath string = ""
var PrintConf bool = false

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
		}, &cli.BoolFlag{
			Name:        "printconf",
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

func RealMain(c *cli.Context) error {
	closer, err := uauth.Start(uauth.DefaultConf)
	if err != nil {
		return err
	}
	defer closer()
	signal := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", signal.String())
	return nil
}
