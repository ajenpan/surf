package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	etcclientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegistryOpts struct {
	NodeId     string             `json:"node_id"`
	ServerType uint16             `json:"server_type"`
	ServerName string             `json:"server_name"`
	NodeData   string             `json:"node_data"`
	TimeoutSec int64              `json:"timeout_sec"`
	EtcdConf   etcclientv3.Config `json:"etcd_conf"`
}

type EtcdRegistry struct {
	cli      *etcclientv3.Client
	nodekey  string
	opts     EtcdRegistryOpts
	closeCh  chan struct{}
	nodeData string
	mu       sync.RWMutex
	leaseID  etcclientv3.LeaseID
}

func NewEtcdRegistry(opts EtcdRegistryOpts) (*EtcdRegistry, error) {
	etcdcli, err := etcclientv3.New(opts.EtcdConf)
	if err != nil {
		return nil, err
	}

	if opts.TimeoutSec < 5 {
		opts.TimeoutSec = 5
	}

	nodekey := strings.Join([]string{"registry", opts.ServerName, opts.NodeId}, "/")
	resp, err := etcdcli.Get(context.Background(), nodekey, etcclientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}

	if resp.Count >= 1 {
		return nil, fmt.Errorf("node alread exist, key:%s", nodekey)
	}

	grantResp, err := etcdcli.Grant(context.Background(), opts.TimeoutSec)
	if err != nil {
		return nil, err
	}

	ret := &EtcdRegistry{
		nodekey: nodekey,
		cli:     etcdcli,
		closeCh: make(chan struct{}),
		opts:    opts,
		leaseID: grantResp.ID,
	}

	err = ret.UpdateNodeData(opts.NodeData)

	if err != nil {
		return nil, err
	}

	go ret.keepAlive()
	return ret, nil
}

func (reg *EtcdRegistry) unregister() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(reg.opts.TimeoutSec)*time.Second)
	defer cancel()
	_, err := reg.cli.Delete(ctx, reg.nodekey)
	return err
}

func (reg *EtcdRegistry) UpdateNodeData(newdata string) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(reg.opts.TimeoutSec)*time.Second)
	defer cancel()

	_, err := reg.cli.Put(ctx, reg.nodekey, newdata, etcclientv3.WithLease(reg.leaseID))
	if err != nil {
		return err
	}

	reg.nodeData = newdata
	return nil
}

func (reg *EtcdRegistry) keepAlive() error {
	ticker := time.NewTicker(time.Duration(reg.opts.TimeoutSec/2) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := reg.cli.KeepAliveOnce(context.Background(), reg.leaseID)
			if err != nil {
				return err
			}
		case <-reg.closeCh:
			return nil
		}
	}
}

func (reg *EtcdRegistry) Close() error {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	select {
	case <-reg.closeCh:
		return nil
	default:
		close(reg.closeCh)
		reg.unregister()
		return reg.cli.Close()
	}
}
