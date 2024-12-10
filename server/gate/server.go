package gate

import (
	"log/slog"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/network"
	utilsRsa "github.com/ajenpan/surf/core/utils/rsagen"
)

var log = slog.Default()

func StartNodeListener(r *Gate, addr string) (func(), error) {
	ws, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   addr,
		OnConnPacket: r.OnNodePacket,
		OnConnEnable: r.nodeConnStore.OnConnEnable,
		OnConnAuth:   r.OnNodeAuth,
	})
	if err != nil {
		return nil, err
	}
	err = ws.Start()
	log.Info("gate node listen start success", "addr", addr)
	return func() { ws.Stop() }, err
}

func Start(r *Gate, cfg *Config) (func() error, error) {
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

	r.ClientPublicKey = ppk
	r.NodePublicKey = ppk

	r.clientConnStore = core.NewClientConnStore(r.OnConnEnable)
	r.nodeConnStore = core.NewNodeConnStore(r.OnNodeStatus)

	ug, err := network.NewWSServer(network.WSServerOptions{
		ListenAddr:   cfg.ClientListenAddr,
		OnConnPacket: r.OnConnPacket,
		OnConnEnable: r.clientConnStore.OnConnEnable,
		OnConnAuth:   r.OnConnAuth,
	})

	if err != nil {
		return nil, err
	}

	err = ug.Start()
	if err != nil {
		return nil, err
	}

	log.Info("gate client listen start success", "addr", cfg.ClientListenAddr)

	ccloser = func() {
		ug.Stop()
	}

	ncloser, err = StartNodeListener(r, cfg.NodeListenAddr)
	if err != nil {
		return nil, err
	}

	initok = true
	return closer, nil
}
