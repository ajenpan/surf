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

func (ug *ClientGate) Start() error {
	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   ug.opts.WsListenAddr,
		OnConnPacket: ug.opts.OnConnPacket,
		OnConnEnable: ug.onConnEnable,
		OnConnAuth:   ug.opts.OnConnAuth,
	})
	if err != nil {
		return err
	}
	ug.wssvr = ws
	err = ws.Start()
	return err
}

func (ug *ClientGate) Stop() error {
	if ug.wssvr != nil {
		return ug.wssvr.Stop()
	}
	return nil
}

func (ug *ClientGate) GetConnByUid(uid uint32) (network.Conn, bool) {
	return ug.ClientConn.LoadByUID(uid)
}

func (ug *ClientGate) GetConnByCid(cid string) (network.Conn, bool) {
	return ug.ClientConn.LoadByCID(cid)
}

func (ug *ClientGate) onConnEnable(conn network.Conn, enable bool) {
	if enable {
		log.Debugf("OnConnEnable: id:%v,addr:%v,uid:%v,urid:%v,enable:%v", conn.ConnID(), conn.RemoteAddr(), conn.UserID(), conn.UserRole(), enable)
		currConn, got := ug.ClientConn.SwapByUID(conn)
		if got {
			ud := currConn.GetUserData()
			currConn.SetUserData(nil)

			conn.SetUserData(ud)
			log.Infof("OnConnEnable: repeat conn, close old conn:%v,uid:%v", currConn.ConnID(), currConn.UserID())
			currConn.Close()
		} else {
			ug.opts.OnConnEnable(conn, true)
		}
	} else {
		currConn, got := ug.ClientConn.Delete(conn.ConnID())
		if got {
			ug.opts.OnConnEnable(currConn, false)
		}
	}
}
