package gate

import (
	"github.com/ajenpan/surf/core"
	logger "github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	utilsRsa "github.com/ajenpan/surf/core/utils/rsagen"
)

var log = logger.Default

func StartNodeListener(r *Gate, addr string) (func(), error) {
	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   addr,
		OnConnPacket: r.OnNodePacket,
		OnConnEnable: r.OnNodeStatus,
		OnConnAuth:   r.OnNodeAuth,
	})
	if err != nil {
		return nil, err
	}
	err = ws.Start()
	return func() {
		ws.Stop()
	}, err
}

func Start(cfg *Config) (func() error, error) {
	ppk, err := utilsRsa.LoadRsaPublicKeyFromUrl(cfg.RsaPublicKeyFile)
	if err != nil {
		return nil, err
	}

	initok := false
	var ccloser func()
	var ncloser func()

	closer := func() error {
		if ccloser != nil {
			ccloser()
		}
		if ncloser != nil {
			ncloser()
		}
		return nil
	}

	defer func() {
		if !initok {
			closer()
		}
	}()

	r := &Gate{
		NodeConn:        core.NewConnStore(),
		NodeID:          cfg.NodeID,
		Marshaler:       &marshal.ProtoMarshaler{},
		ClientPublicKey: ppk,
		NodePublicKey:   ppk,
	}

	ug := core.NewClientGate(core.ClientGateOptions{
		WsListenAddr: cfg.ClientListenAddr,
		OnConnPacket: r.OnConnPacket,
		OnConnEnable: r.OnConnEnable,
		OnConnAuth:   r.OnConnAuth,
	})
	err = ug.Start()
	if err != nil {
		return nil, err
	}
	log.Infof("gate client listen on %s start success", cfg.ClientListenAddr)
	ccloser = func() {
		ug.Stop()
	}

	ncloser, err = StartNodeListener(r, cfg.NodeListenAddr)
	if err != nil {
		return nil, err
	}

	r.ClientConn = ug
	initok = true
	return closer, nil
}
