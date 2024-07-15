package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/utils/rsagen"
	utilSignal "github.com/ajenpan/surf/core/utils/signal"
)

var (
	Name       string = "route"
	Version    string = "unknow"
	GitCommit  string = "unknow"
	BuildAt    string = "unknow"
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
	if err := Run(); err != nil {
		fmt.Println(err)
	}
}

func Run() error {
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

const PrivateKeyFile = "private.pem"
const PublicKeyFile = "public.pem"

func RealMain(c *cli.Context) error {
	ppk, err := rsagen.LoadRsaPrivateKeyFromFile(PrivateKeyFile)
	if err != nil {
		return err
	}

	_, err = rsagen.LoadRsaPublicKeyFromFile(PublicKeyFile)
	if err != nil {
		return err
	}

	jwt, _ := auth.GenerateToken(ppk, &auth.UserInfo{
		UId:   10001,
		UName: "gdclient",
		URole: 0,
	}, 240*time.Hour)

	fmt.Println(jwt)

	s := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", s.String())
	return nil
}
