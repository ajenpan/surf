package gate

import (
	logger "github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/marshal"
	"github.com/ajenpan/surf/core/network"
	utilsRsa "github.com/ajenpan/surf/core/utils/rsagen"
)

var log = logger.Default

func StartClientListener(r *Gate, addr string) (func(), error) {
	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   addr,
		OnConnPacket: r.OnConnPacket,
		OnConnStatus: r.OnConnEnable,
		OnConnAuth:   r.OnConnAuth,
	})
	if err != nil {
		return nil, err
	}
	err = ws.Start()
	return func() {
		ws.Stop()
	}, err
}

func StartNodeListener(r *Gate, addr string) (func(), error) {
	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   addr,
		OnConnPacket: r.OnNodePacket,
		OnConnStatus: r.OnNodeStatus,
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
		ClientConn:      NewConnStore(),
		NodeConn:        NewConnStore(),
		Marshaler:       &marshal.ProtoMarshaler{},
		ClientPublicKey: ppk,
		NodePublicKey:   ppk,
	}

	ccloser, err = StartClientListener(r, cfg.ClientListenAddr)
	if err != nil {
		return nil, err
	}

	ncloser, err = StartNodeListener(r, cfg.NodeListenAddr)
	if err != nil {
		return nil, err
	}

	initok = true
	return closer, nil
}
