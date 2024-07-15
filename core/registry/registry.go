package registry

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

type Registry struct {
	cli     *etcclientv3.Client
	nodekey string

	Options
}

func NewRegistry(opts Options) (*Registry, error) {
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

	ret := &Registry{
		nodekey: nodekey,
		cli:     etcdcli,
	}

	err = ret.UpdateNodeData(opts.NodeData)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (reg *Registry) UpdateNodeData(newdata string) error {
	reg.NodeData = newdata
	_, err := reg.cli.Put(context.Background(), reg.nodekey, newdata)
	if err != nil {
		return err
	}
	return nil
}
