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
	Conn *clientv3.Client
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// 即:	target := fmt.Sprintf("discovery:///%s/%s/%s", env, appid, region)
	prefix := fmt.Sprintf("/%s/", target.Endpoint)

	r := &Resolver{
		Conn:   b.Conn,
		cc:     cc,
		prefix: prefix,
	}
	log.Infof("---> etcdv3 grpc to find target:%s \r\n", prefix)
	go r.watchers()
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
	Conn    *clientv3.Client
	KV      clientv3.KV
	Watcher clientv3.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
}

func (r *Resolver) ResolveNow(rn resolver.ResolveNowOptions) {
	// todo
}

func (r *Resolver) Close() {
	log.Infof("---> etcdv3 -------Close()")
	r.cancel()
	r.Watcher.Close()
	return
}

func (r *Resolver) watchers() {
	r.ctx, r.cancel = context.WithCancel(context.Background())

	r.addresses = make(map[string]resolver.Address)
	r.Watcher = clientv3.NewWatcher(r.Conn)
	r.KV = clientv3.NewKV(r.Conn)
	// log.Infoln("---> etcdv3 xx----func (r *Resolver) Watcher  r.prefix:", r.prefix)
	// 先获取一次
	ins, err := r.KV.Get(r.ctx, r.prefix, clientv3.WithPrefix())
	if err != nil {
		log.Infof("---> etcdv3 err: %s", err.Error())
		return
	}
	for _, kv := range ins.Kvs {
		r.setAddress(string(kv.Key), string(kv.Value))
	}

	r.cc.UpdateState(resolver.State{
		Addresses: r.getAddresses(),
	})

	// 监听key
	watchChan := r.Watcher.Watch(r.ctx, r.prefix, clientv3.WithPrefix(), clientv3.WithRev(0)) // 监听的revision起点
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

	r.Close()
}

func (r *Resolver) setAddress(key, address string) {
	r.Lock()
	defer r.Unlock()
	// log.Infoln("---> etcdv3 ---- setAddress  key:val => ", key, ":", address)

	r.addresses[key] = resolver.Address{Addr: string(address)}
}

func (r *Resolver) delAddress(key string) {
	r.Lock()
	defer r.Unlock()
	delete(r.addresses, key)
}

func (r *Resolver) getAddresses() []resolver.Address {
	items := make([]resolver.Address, 0, len(r.addresses))

	for _, v := range r.addresses {
		items = append(items, v)
	}

	return items
}
