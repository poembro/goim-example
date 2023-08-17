package etcdv3

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	log "github.com/golang/glog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

var (
	once     sync.Once
	etcdConn *clientv3.Client = nil
)

func newClient(etcdAddr string) (*clientv3.Client, error) {
	if etcdConn != nil {
		return etcdConn, nil
	}
	var err error
	once.Do(func() {
		etcdConn, err = clientv3.New(clientv3.Config{
			Endpoints:          strings.Split(etcdAddr, ","),
			DialTimeout:        time.Second * time.Duration(5),
			MaxCallSendMsgSize: 2 * 1024 * 1024,
		})
	})

	return etcdConn, err
}

type Options struct {
	ctx       context.Context
	namespace string
	ttl       time.Duration
	maxRetry  int
}

// Registry is etcd registry.
type Registry struct {
	Opts  *Options
	Conn  *clientv3.Client
	KV    clientv3.KV
	lease clientv3.Lease
}

// New creates etcd registry
func New(nodes string) (r *Registry) {
	op := &Options{
		ctx:       context.Background(),
		namespace: "",
		ttl:       time.Second * 15,
		maxRetry:  5, // 重试 5次
	}
	client, err := newClient(nodes)
	if err != nil {
		log.Infof("---> etcdv3  err: \"%s\" ", err.Error())
		return nil
	}

	return &Registry{
		Opts: op,
		Conn: client,
		KV:   clientv3.NewKV(client),
	}
}

func (r *Registry) ResolverEtcd() {
	builder := &Builder{
		Conn: r.Conn,
	}

	resolver.Register(builder)
}

// Register the registration.
func (r *Registry) Register(env, appid, region, zone, ip, port string) error {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s:%s", env, appid, region, zone, ip, port)
	value := fmt.Sprintf("%s:%s", ip, port)
	log.Infof("---> etcdv3 service register to etcd \"%s\" ", key)
	r.Opts.namespace = key
	if r.lease != nil {
		r.lease.Close()
	}
	r.lease = clientv3.NewLease(r.Conn)
	leaseID, err := r.registerWithKV(r.Opts.ctx, key, value)
	if err != nil {
		return err
	}

	go r.heartBeat(r.Opts.ctx, leaseID, key, value)
	return nil
}

// Deregister the registration.
func (r *Registry) Deregister() error {
	defer func() {
		if r.lease != nil {
			r.lease.Close()
		}
	}()
	_, err := r.Conn.Delete(r.Opts.ctx, r.Opts.namespace)
	return err
}

// AllService return the service instances in memory according to the service name.
func (r *Registry) ServiceList(env, appid, region, zone string) map[string]string {
	dst := make(map[string]string)
	key := fmt.Sprintf("/%s/%s/%s", env, appid, region) // 服务发现 上海所有节点
	resp, err := r.KV.Get(r.Opts.ctx, key, clientv3.WithPrefix())
	if err != nil {
		log.Infof("---> etcdv3 err k:\"%s\"  v:\"%s\" ", key, err.Error())
		return dst
	}

	for _, kv := range resp.Kvs {
		k := string(kv.Key)
		v := string(kv.Value)
		dst[k] = v
	}

	return dst
}

// registerWithKV create a new lease, return current leaseID
func (r *Registry) registerWithKV(ctx context.Context, key string, value string) (clientv3.LeaseID, error) {
	grant, err := r.lease.Grant(ctx, int64(r.Opts.ttl.Seconds()))
	if err != nil {
		return 0, err
	}
	_, err = r.Conn.Put(ctx, key, value, clientv3.WithLease(grant.ID))
	if err != nil {
		return 0, err
	}
	return grant.ID, nil
}

func (r *Registry) heartBeat(ctx context.Context, leaseID clientv3.LeaseID, key string, value string) {
	curLeaseID := leaseID
	kac, err := r.Conn.KeepAlive(ctx, leaseID)
	if err != nil {
		curLeaseID = 0
	}
	rand.Seed(time.Now().Unix())

	for {
		if curLeaseID == 0 {
			// try to registerWithKV
			retreat := []int{}
			for retryCnt := 0; retryCnt < r.Opts.maxRetry; retryCnt++ {
				if ctx.Err() != nil {
					return
				}
				// prevent infinite blocking
				idChan := make(chan clientv3.LeaseID, 1)
				errChan := make(chan error, 1)
				cancelCtx, cancel := context.WithCancel(ctx)
				go func() {
					defer cancel()
					id, registerErr := r.registerWithKV(cancelCtx, key, value)
					if registerErr != nil {
						errChan <- registerErr
					} else {
						idChan <- id
					}
				}()

				select {
				case <-time.After(3 * time.Second):
					cancel()
					continue
				case <-errChan:
					continue
				case curLeaseID = <-idChan:
				}

				kac, err = r.Conn.KeepAlive(ctx, curLeaseID)
				if err == nil {
					break
				}
				retreat = append(retreat, 1<<retryCnt)
				time.Sleep(time.Duration(retreat[rand.Intn(len(retreat))]) * time.Second)
			}
			if _, ok := <-kac; !ok {
				// retry failed
				return
			}
		}

		select {
		case _, ok := <-kac:
			if !ok {
				if ctx.Err() != nil {
					// channel closed due to context cancel
					return
				}
				// need to retry registration
				curLeaseID = 0
				continue
			}
		case <-r.Opts.ctx.Done():
			return
		}
	}
}
