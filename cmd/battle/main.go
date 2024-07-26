package main

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/rsagen"
	utilSignal "github.com/ajenpan/surf/core/utils/signal"
	battleHandler "github.com/ajenpan/surf/server/battle/handler"
	_ "github.com/ajenpan/surf/server/games/niuniu"
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

func LoadAuthPublicKey() (*rsa.PublicKey, error) {
	publicRaw, err := os.ReadFile("public.pem")
	if err != nil {
		return nil, err
	}
	pk, err := rsagen.ParseRsaPublicKeyFromPem(publicRaw)
	return pk, err
}

func RealMain(c *cli.Context) error {
	pk, err := LoadAuthPublicKey()
	if err != nil {
		panic(err)
	}

	h := battleHandler.New()

	listener, err := network.NewTcpServer(network.TcpServerOptions{
		ListenAddr:   listenAt,
		OnConnPacket: h.OnMessage,
		OnConnEnable: h.OnConn,
		OnConnAuth: func(tokenRaw []byte) (auth.User, error) {
			return auth.VerifyToken(pk, tokenRaw)
		},
	})

	if err != nil {
		panic(err)
	}

	go listener.Start()
	defer listener.Stop()

	s := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", s.String())
	return nil
}
