package registry

import (
	"context"

	etcclientv3 "go.etcd.io/etcd/client/v3"
)

func NewEtcdWatcher(conf etcclientv3.Config, watchKey string, cb func(ev *etcclientv3.Event), op ...etcclientv3.OpOption) (*EtcdWatcher, error) {
	cli, err := etcclientv3.New(conf)
	if err != nil {
		return nil, err
	}

	watchC := cli.Watch(context.Background(), watchKey, op...)
	resp, err := cli.Get(context.Background(), watchKey, op...)
	if err != nil {
		return nil, err
	}

	res := &EtcdWatcher{
		cli:    cli,
		closed: make(chan struct{}),
	}

	go func() {
		for _, kv := range resp.Kvs {
			ev := &etcclientv3.Event{
				Type:   etcclientv3.EventTypePut,
				Kv:     kv,
				PrevKv: kv,
			}
			cb(ev)
		}
		for resp := range watchC {
			for _, ev := range resp.Events {
				cb(ev)
			}
		}
	}()

	return res, nil
}

type EtcdWatcher struct {
	cli    *etcclientv3.Client
	closed chan struct{}
}

func (wr *EtcdWatcher) Close() {
	select {
	case <-wr.closed:
	default:
		close(wr.closed)

		if wr.cli != nil {
			wr.cli.Close()
			wr.cli = nil
		}
	}
}
