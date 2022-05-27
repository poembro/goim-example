package etcdv3

import (
	"context"
	"fmt"
	"sync"

	log "github.com/golang/glog"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"google.golang.org/grpc/resolver"
)

type Builder struct {
	client *clientv3.Client
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	prefix := fmt.Sprintf("/%s/", target.Endpoint)
	r := &Resolver{
		client: b.client,
		cc:     cc,
		prefix: prefix,
	}
	log.Infof("---> etcdv3 grpc to find target:%s \r\n", prefix)
	go r.Watcher(prefix)
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

func (r *Builder) Scheme() string {
	return "discovery"
}

////////////////////////////////////////////////////////////

type Resolver struct {
	sync.RWMutex

	cc        resolver.ClientConn
	prefix    string
	addresses map[string]resolver.Address

	//////////etcd//////////
	client     *clientv3.Client
	kvCli      clientv3.KV
	watcherCli clientv3.Watcher
	ctx        context.Context
	cancel     context.CancelFunc
}

func (r *Resolver) ResolveNow(rn resolver.ResolveNowOptions) {
	// todo
}

func (r *Resolver) Close() {
	log.Infof("---> etcdv3 -------Stop()被调用了---------->")
	r.cancel()
	r.watcherCli.Close()
	return
}

func (r *Resolver) Watcher(prefix string) {
	r.ctx, r.cancel = context.WithCancel(context.Background())

	r.addresses = make(map[string]resolver.Address)
	r.watcherCli = clientv3.NewWatcher(r.client)
	r.kvCli = clientv3.NewKV(r.client)

	// 先获取一次
	resp, err := r.kvCli.Get(r.ctx, r.prefix, clientv3.WithPrefix())
	if err != nil {
		log.Infof("---> etcdv3 err: %s", err.Error())
		return
	}
	for _, kv := range resp.Kvs {
		r.setAddress(string(kv.Key), string(kv.Value))
	}

	r.cc.UpdateState(resolver.State{
		Addresses: r.getAddresses(),
	})

	// 监听key
	watchChan := r.watcherCli.Watch(r.ctx, r.prefix, clientv3.WithPrefix(), clientv3.WithRev(0)) // 监听的revision起点
	for response := range watchChan {
		for _, event := range response.Events {
			switch event.Type {
			case mvccpb.PUT:
				r.setAddress(string(event.Kv.Key), string(event.Kv.Value))
			case mvccpb.DELETE:
				r.delAddress(string(event.Kv.Key))
			}
		}

		r.cc.UpdateState(resolver.State{
			Addresses: r.getAddresses(),
		})
	}
	log.Infof("---> etcdv3 -------func (r *Resolver) Watcher 被调用了---------->")
	r.Close()
}

func (r *Resolver) setAddress(key, address string) {
	r.Lock()
	defer r.Unlock()
	r.addresses[key] = resolver.Address{Addr: string(address)}
}

func (r *Resolver) delAddress(key string) {
	r.Lock()
	defer r.Unlock()
	delete(r.addresses, key)
}

func (r *Resolver) getAddresses() []resolver.Address {
	addresses := make([]resolver.Address, 0, len(r.addresses))

	for _, address := range r.addresses {
		addresses = append(addresses, address)
	}

	return addresses
}
