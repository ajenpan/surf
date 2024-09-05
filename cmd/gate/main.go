package main

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/rsagen"
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
		privateKey, publicKey, err := rsagen.GenerateRsaPem(2048)
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

func StartClientListener(ppk *rsa.PrivateKey, r *gate.Gate) (func(), error) {
	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   ":9999",
		OnConnPacket: r.OnConnPacket,
		OnConnEnable: r.OnConnEnable,
		OnConnAuth: func(data []byte) (auth.User, error) {
			return auth.VerifyToken(&ppk.PublicKey, data)
		},
	})
	if err != nil {
		return nil, err
	}
	ws.Start()
	return func() {
		ws.Stop()
	}, nil
}

func StartNodeListener(ppk *rsa.PrivateKey, r *gate.Gate) (func(), error) {
	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   ":9998",
		OnConnPacket: r.OnNodePacket,
		OnConnEnable: r.OnNodeEnable,
		OnConnAuth: func(data []byte) (auth.User, error) {
			return auth.VerifyToken(&ppk.PublicKey, data)
		},
	})
	if err != nil {
		return nil, err
	}
	ws.Start()
	return func() {
		ws.Stop()
	}, nil

	// tcpsvr, err := network.NewTcpServer(network.TcpServerOptions{
	// 	ListenAddr:   ":9998",
	// 	OnConnPacket: r.OnNodePacket,
	// 	OnConnEnable: r.OnNodeEnable,
	// 	OnConnAuth: func(data []byte) (auth.User, error) {
	// 		return auth.VerifyToken(&ppk.PublicKey, data)
	// 	}},
	// )
	// if err != nil {
	// 	return nil, err
	// }
	// tcpsvr.Start()
	// return func() {
	// 	tcpsvr.Stop()
	// }, nil
}

func RealMain(c *cli.Context) error {
	log.Default.SetOutput(os.Stdout)

	privateRaw, _, err := ReadRSAKey()
	if err != nil {
		panic(err)
	}

	ppk, err := rsagen.ParseRsaPrivateKeyFromPem(privateRaw)
	if err != nil {
		return err
	}

	jwt, _ := auth.GenerateToken(ppk, &auth.UserInfo{
		UId:   10001,
		UName: "gdclient",
		URole: 0,
	}, 240*time.Hour)

	fmt.Println(jwt)

	r := gate.NewGate()
	closer, err := StartClientListener(ppk, r)
	if err != nil {
		return err
	}
	defer closer()

	nodeCloser, err := StartNodeListener(ppk, r)
	if err != nil {
		return err
	}
	defer nodeCloser()

	s := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", s.String())
	return nil
}
