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
	Name       string = "gateclient"
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

var gatesvr = gate.NewGate()

func StartTcp(ppk *rsa.PrivateKey) (func(), error) {
	uinfo := &auth.UserInfo{
		UId:   10001,
		UName: "gdclient",
		URole: 0,
	}
	jwt, _ := auth.GenerateToken(ppk, uinfo, 2400*time.Hour)
	fmt.Println(jwt)

	tcpsvr, err := network.NewTcpServer(network.TcpServerOptions{
		ListenAddr:   ":19999",
		OnConnPacket: gatesvr.OnNodePacket,
		OnConnEnable: gatesvr.OnNodeEnable,
		OnConnAuth: func(data []byte) (auth.User, error) {
			return auth.VerifyToken(&ppk.PublicKey, data)
		}},
	)
	if err != nil {
		return nil, err
	}
	tcpsvr.Start()

	client := network.NewTcpClient(network.TcpClientOptions{
		RemoteAddress:  "localhost:19999",
		AuthToken:      []byte(jwt),
		UInfo:          uinfo,
		ReconnectDelay: time.Second,
		OnConnPacket: func(c network.Conn, h *network.HVPacket) {
			fmt.Println("conn recv pk:", h.Meta.GetType())
		},
		OnConnEnable: func(c network.Conn, b bool) {
			fmt.Printf("conn:%v status:%v\n", c.ConnID(), b)
		},
	})
	client.Start()

	return func() {
		tcpsvr.Stop()
		client.Close()
	}, nil
}

// func StartWs(ppk *rsa.PrivateKey) (func(), error) {
// 	uinfo := &auth.UserInfo{
// 		UId:   10001,
// 		UName: "gdclient",
// 		URole: 0,
// 	}
// 	jwt, _ := auth.GenerateToken(ppk, uinfo, 2400*time.Hour)
// 	fmt.Println(jwt)

// 	tcpsvr, err := network.NewWSServer(network.WSServerOptions{
// 		ListenAddr:   ":19999",
// 		OnConnPacket: gatesvr.OnNodePacket,
// 		OnConnEnable: gatesvr.OnNodeEnable,
// 		OnConnAuth: func(data []byte) (auth.User, error) {
// 			return auth.VerifyToken(&ppk.PublicKey, data)
// 		}},
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	tcpsvr.Start()

// 	client := network.NewTcpClient(network.TcpClientOptions{
// 		RemoteAddress:  "localhost:19999",
// 		AuthToken:      []byte(jwt),
// 		UInfo:          uinfo,
// 		ReconnectDelay: time.Second,
// 		OnConnPacket: func(c network.Conn, h *network.HVPacket) {
// 			fmt.Println("conn recv pk:", h.Meta.GetType())
// 		},
// 		OnConnEnable: func(c network.Conn, b bool) {
// 			fmt.Printf("conn:%v status:%v\n", c.ConnID(), b)
// 		},
// 	})
// 	client.Start()
// 	return func() {
// 		tcpsvr.Stop()
// 		client.Close()
// 	}, nil
// }

func RealMain(c *cli.Context) error {
	privateRaw, _, err := ReadRSAKey()
	if err != nil {
		panic(err)
	}

	ppk, err := rsagen.ParseRsaPrivateKeyFromPem(privateRaw)
	if err != nil {
		return err
	}
	tcpclose, err := StartTcp(ppk)
	if err != nil {
		panic(err)
	}
	defer tcpclose()

	s := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", s.String())
	return nil
}
