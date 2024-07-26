package watch

import (
	"context"
	"fmt"
	"strings"

	etcclientv3 "go.etcd.io/etcd/client/v3"
)

type Options struct {
	NodeId     string
	ServerName string
	NodeData   string
	EtcdConf   etcclientv3.Config `json:"etcd_conf"`
}

type WatchCli struct {
	cli     *etcclientv3.Client
	nodekey string

	Options
}

func NewWatchCli(opts Options) (*WatchCli, error) {
	etcdcli, err := etcclientv3.New(opts.EtcdConf)
	if err != nil {
		return nil, err
	}

	nodekey := strings.Join([]string{"registry", opts.ServerName, opts.NodeId}, "/")

	resp, err := etcdcli.Get(context.Background(), nodekey, etcclientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}

	if resp.Count >= 1 {
		return nil, fmt.Errorf("node alread exist, key:%s", nodekey)
	}

	ret := &WatchCli{
		nodekey: nodekey,
		cli:     etcdcli,
	}

	return ret, nil
}
