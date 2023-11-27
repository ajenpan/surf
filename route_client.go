package surf

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/ajenpan/surf/auth"
	"github.com/ajenpan/surf/server"
	"github.com/ajenpan/surf/utils/rsagen"
	"github.com/ajenpan/surf/utils/signal"
)

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

func LoadAuthKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	rawpr, rawpu, err := ReadRSAKey()
	if err != nil {
		return nil, nil, err
	}
	pr, err := rsagen.ParseRsaPrivateKeyFromPem(rawpr)
	if err != nil {
		return nil, nil, err
	}
	pu, err := rsagen.ParseRsaPublicKeyFromPem(rawpu)
	if err != nil {
		return nil, nil, err
	}
	return pr, pu, nil
}

type RouteClient struct {
	addrs  []string
	jwtstr string
}

func (rc *RouteClient) Start() {
	var err error
	pk, _, err := LoadAuthKey()
	if err != nil {
		panic(err)
	}

	jwtstr, err := auth.GenerateToken(pk, &auth.UserInfo{
		UId:   100010,
		UName: "gameserver01",
		URole: "gameserver",
	}, time.Hour*24*365)

	if err != nil {
		panic(err)
	}
	fmt.Println(jwtstr)

	c := server.NewTcpClient(&server.TcpClientOptions{
		RemoteAddress:        "localhost:80",
		AuthToken:            jwtstr,
		ReconnectDelaySecond: 10,
		OnMessage:            rc.OnMessage,
		OnStatus:             rc.OnStatus,
	})

	err = c.Connect()
	if err != nil {
		fmt.Println(err)
	}

	defer c.Close()
	signal.WaitShutdown()
}

func (rc *RouteClient) Stop() {

}

func (h *RouteClient) OnStatus(s *server.TcpClient, enable bool) {
	fmt.Println("OnConnect:", s.UserID(), s.SessionID(), enable)
}

func (h *RouteClient) OnMessage(s *server.TcpClient, m *server.Message) {
	typ := m.GetMsgtype()
	switch typ {
	case 0:
	case 1:
	case 2:

	}
}
