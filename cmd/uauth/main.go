package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/utils/rsagen"
	"github.com/ajenpan/surf/server"
	"github.com/ajenpan/surf/server/uauth"
)

var Name string = server.NodeName_UAuth
var Version string = "unknown"
var GitCommit string = "unknown"
var BuildAt string = "unknown"
var BuildBy string = "unknown"

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

const PrivateKeyFile = "private.pem"
const PublicKeyFile = "public.pem"

func ReadRSAKey() ([]byte, []byte, error) {
	privateRaw, err := os.ReadFile(PrivateKeyFile)
	if err != nil {
		privateKey, publicKey, err := rsagen.GenerateRsaPem(512)
		if err != nil {
			return nil, nil, err
		}
		privateRaw = []byte(privateKey)
		os.WriteFile(PrivateKeyFile, []byte(privateKey), 0644)
		os.WriteFile(PublicKeyFile, []byte(publicKey), 0644)
	}
	publicRaw, err := os.ReadFile(PublicKeyFile)
	if err != nil {
		return nil, nil, err
	}
	return privateRaw, publicRaw, nil
}

func RealMain(c *cli.Context) error {
	InitConfig()

	privateRaw, publicRaw, err := ReadRSAKey()
	if err != nil {
		return err
	}

	h := uauth.New(privateRaw, publicRaw)
	opts := &core.ServerInfo{
		Svr: h,
	}

	conf := &core.NodeConf{
		SurfConf: core.SurfConfig{
			HttpListenAddr:    ":9999",
			PublicKeyFilePath: "file://" + PublicKeyFile,
		},
	}
	ninfo := &auth.NodeInfo{
		NId:   10200,
		NName: server.NodeName_UAuth,
		NType: server.NodeType_UAuth,
	}
	surf, err := core.NewSurf(ninfo, conf, opts)
	if err != nil {
		return err
	}
	defer surf.Close()
	err = surf.Run()
	return err
}
