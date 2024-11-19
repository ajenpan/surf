package registry

import etcclientv3 "go.etcd.io/etcd/client/v3"

type EtcdWatchOpts struct {
	Key      string             `json:"key"`
	EtcdConf etcclientv3.Config `json:"etcd_conf"`
}

type EtcdWatch struct {
	cli  *etcclientv3.Client
	opts EtcdWatchOpts
}

func NewEtcdWatch(opts EtcdWatchOpts) (*EtcdWatch, error) {
	cli, err := etcclientv3.New(opts.EtcdConf)
	if err != nil {
		return nil, err
	}

	return &EtcdWatch{
		cli:  cli,
		opts: opts,
	}, nil
}
