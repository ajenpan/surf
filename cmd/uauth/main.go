package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"

	utilSignal "github.com/ajenpan/surf/core/utils/signal"
	"github.com/ajenpan/surf/server/uauth"
)

var Version string = "unknown"
var GitCommit string = "unknown"
var BuildAt string = "unknown"
var BuildBy string = "unknown"
var Name string = "uauth"

var ConfigPath string = ""

func InitConfig() {
	raw, err := os.ReadFile(ConfigPath)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw, &uauth.DefaultConf)
	if err != nil {
		slog.Error("read config err", "err", err)
	}
}

func CmdPrintConfig() {
	raw, err := json.Marshal(&uauth.DefaultConf)
	if err != nil {
		slog.Error("marshal config err", "err", err)
		return
	}
	fmt.Println(string(raw))
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
	app.Commands = []*cli.Command{
		{
			Name:   "printconf",
			Hidden: true,
			Action: func(c *cli.Context) error {
				CmdPrintConfig()
				return nil
			},
		},
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       "config.json",
			Destination: &ConfigPath,
		},
	}

	app.Action = RealMain

	if err := app.Run(os.Args); err != nil {
		slog.Error("run app err", "err", err)
		os.Exit(-1)
	}
}

func RealMain(c *cli.Context) error {
	InitConfig()

	closer, err := uauth.Start(uauth.DefaultConf)
	if err != nil {
		return err
	}
	defer closer()
	signal := utilSignal.WaitShutdown()
	slog.Info("recv signal", "signal", signal.String())
	return nil
}
