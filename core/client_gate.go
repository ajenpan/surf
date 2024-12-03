package core

import (
	"github.com/ajenpan/surf/core/network"
)

type ClientGateOptions struct {
	WsListenAddr string

	OnConnPacket network.FuncOnConnPacket
	OnConnEnable network.FuncOnConnEnable
	OnConnAuth   network.FuncOnConnAuth
}

func NewClientGate(opts ClientGateOptions) *ClientGate {
	return &ClientGate{
		opts:       opts,
		ClientConn: NewConnStore(),
	}
}

type ClientGate struct {
	opts       ClientGateOptions
	ClientConn *ConnStore

	wssvr *network.WSServer
}

func (cli *ClientGate) Start() error {
	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   cli.opts.WsListenAddr,
		OnConnPacket: cli.opts.OnConnPacket,
		OnConnEnable: cli.onConnEnable,
		OnConnAuth:   cli.opts.OnConnAuth,
	})
	if err != nil {
		return err
	}
	cli.wssvr = ws
	err = ws.Start()
	return err
}

func (cli *ClientGate) Stop() error {
	if cli.wssvr != nil {
		return cli.wssvr.Stop()
	}
	return nil
}

func (cli *ClientGate) GetConnByUid(uid uint32) (network.Conn, bool) {
	return cli.ClientConn.LoadByUID(uid)
}

func (cli *ClientGate) GetConnByCid(cid string) (network.Conn, bool) {
	return cli.ClientConn.LoadByCID(cid)
}

func (cli *ClientGate) onConnEnable(conn network.Conn, enable bool) {
	if enable {
		log.Info("OnConnEnable", "id", conn.ConnID(), "addr", conn.RemoteAddr(), "uid", conn.UserID(), "urid", conn.UserRole(), "enable", enable)
		currConn, got := cli.ClientConn.SwapByUID(conn)
		if got {
			ud := currConn.GetUserData()
			currConn.SetUserData(nil)

			conn.SetUserData(ud)
			log.Info("OnConnEnable: repeat conn, close old conn", "id", currConn.ConnID(), "uid", currConn.UserID())
			currConn.Close()
		} else {
			cli.opts.OnConnEnable(conn, true)
		}
	} else {
		currConn, got := cli.ClientConn.Delete(conn.ConnID())
		if got {
			cli.opts.OnConnEnable(currConn, false)
		}
	}
}
