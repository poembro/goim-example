package etcdv3

import (
	"context"
	"sync"

	log "github.com/golang/glog"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"google.golang.org/grpc/resolver"
)

type Builder struct {
	etcdConn *clientv3.Client
}

func (s *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// 即:	target := fmt.Sprintf("discovery:///%s/%s/%s", env, appid, region)

	r := &Resolver{
		etcdConn:   s.etcdConn,
		targetConn: cc,
		Prefix:     target.URL.Path,
		Items:      make(map[string]resolver.Address),
	}
	log.Infof("etcdv3 grpc to find target:%#v \r\n", target.URL)
	go r.watchers()
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

func (r *Builder) Scheme() string {
	return "discovery"
}

////////////////////////////////////////////////////////////

type Resolver struct {
	lock sync.RWMutex

	targetConn resolver.ClientConn
	Items      map[string]resolver.Address
	Prefix     string

	//////////etcd//////////
	etcdConn *clientv3.Client
}

func (r *Resolver) ResolveNow(rn resolver.ResolveNowOptions) {
	// todo
}

func (r *Resolver) Close() {
	log.Infof("etcdv3 -------Close()")
	return
}

func (r *Resolver) watchers() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := clientv3.NewWatcher(r.etcdConn)
	defer watcher.Close()

	etcdkv := clientv3.NewKV(r.etcdConn)
	log.Infoln("etcdv3  r.Prefix  :", r.Prefix)
	// 先获取一次
	items, err := etcdkv.Get(ctx, r.Prefix, clientv3.WithPrefix())
	if err != nil {
		log.Infof("etcdv3 err: %s", err.Error())
		return
	}
	log.Infoln("etcdv3 items:", items.Kvs)
	for _, kv := range items.Kvs {
		r.creates(string(kv.Key), string(kv.Value))
	}

	r.targetConn.UpdateState(resolver.State{
		Addresses: r.findAll(),
	})

	// 监听key
	watchChan := watcher.Watch(ctx, r.Prefix, clientv3.WithPrefix(), clientv3.WithRev(0)) // 监听的revision起点
	for response := range watchChan {
		for _, event := range response.Events {
			switch event.Type {
			case mvccpb.PUT:
				r.creates(string(event.Kv.Key), string(event.Kv.Value))
			case mvccpb.DELETE:
				r.remove(string(event.Kv.Key))
			}
		}

		r.targetConn.UpdateState(resolver.State{
			Addresses: r.findAll(),
		})
	}
}

func (r *Resolver) creates(key, address string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	log.Infoln("etcdv3 ---- setAddrs  key:val => ", key, ":", address)
	r.Items[key] = resolver.Address{Addr: address}
}

func (r *Resolver) findAll() []resolver.Address {
	r.lock.RLock()
	defer r.lock.RUnlock()

	items := make([]resolver.Address, 0, len(r.Items))
	for _, v := range r.Items {
		items = append(items, v)
	}
	return items
}

func (r *Resolver) remove(key string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.Items, key)
}
