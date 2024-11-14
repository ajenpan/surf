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
	utilsRsa "github.com/ajenpan/surf/core/utils/rsagen"
	utilSignal "github.com/ajenpan/surf/core/utils/signal"
	gate "github.com/ajenpan/surf/server/gate"
)

var (
	Name       string = "gate"
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
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(longVersion())
	}
	app := cli.NewApp()
	app.Version = Version
	app.Name = Name
	app.Action = RealMain
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

const PrivateKeyFile = "private.pem"
const PublicKeyFile = "public.pem"

func ReadRSAKey() ([]byte, []byte, error) {
	privateRaw, err := os.ReadFile(PrivateKeyFile)
	if err != nil {
		privateKey, publicKey, err := utilsRsa.GenerateRsaPem(512)
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
	log.Default.SetOutput(os.Stdout)

	_, _, err := ReadRSAKey()
	if err != nil {
		panic(err)
	}

	pk, err := utilsRsa.LoadRsaPrivateKeyFromFile(PrivateKeyFile)
	if err != nil {
		return err
	}

	testuid := uint32(30000)
	testuser := &auth.UserInfo{
		UId:   testuid,
		UName: fmt.Sprintf("yk%d", testuid),
		URole: 1,
	}
	testjwt, err := auth.GenerateToken(pk, testuser, 24*time.Hour*999)
	if err != nil {
		return err
	}
	log.Info("testjwt:", testjwt)

	cfg := &gate.Config{
		RsaPublicKeyFile: "file://" + PublicKeyFile,
		ClientListenAddr: ":11000",
		NodeListenAddr:   ":13000",
	}

	closer, err := gate.Start(cfg)
	if err != nil {
		return err
	}
	defer closer()

	log.Info("config:", cfg.String())
	log.Info("gate server started")

	s := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", s.String())
	return nil
}
