package registry

import (
	"context"
	"strings"
	"sync"
	"time"

	etcclientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdConfig = etcclientv3.Config

type EtcdRegistryOpts struct {
	NodeId     string     `json:"node_id"`
	NodeType   string     `json:"node_type"`
	TimeoutSec int64      `json:"timeout_sec"`
	EtcdConf   EtcdConfig `json:"etcd_conf"`
}

type EtcdRegistry struct {
	cli      *etcclientv3.Client
	nodekey  string
	opts     EtcdRegistryOpts
	closed   chan struct{}
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

	nodekey := strings.Join([]string{"/nodes", opts.NodeType, opts.NodeId}, "/")
	// resp, err := etcdcli.Get(context.Background(), nodekey, etcclientv3.WithLimit(1))
	// if err != nil {
	// 	return nil, err
	// }
	// if resp.Count >= 1 {
	// 	return nil, fmt.Errorf("node alread exist, key:%s", nodekey)
	// }

	grantResp, err := etcdcli.Grant(context.Background(), opts.TimeoutSec)
	if err != nil {
		return nil, err
	}

	ret := &EtcdRegistry{
		nodekey: nodekey,
		cli:     etcdcli,
		closed:  make(chan struct{}),
		opts:    opts,
		leaseID: grantResp.ID,
	}

	// go ret.keepAlive()
	resp, err := ret.cli.KeepAlive(context.Background(), grantResp.ID)
	if err != nil {
		return nil, err
	}

	go func() {
		for range resp {
			// do nothing
		}
	}()

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
		case <-reg.closed:
			return nil
		}
	}
}

func (reg *EtcdRegistry) Close() error {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	select {
	case <-reg.closed:
		return nil
	default:
		close(reg.closed)
		reg.unregister()
		return reg.cli.Close()
	}
}
