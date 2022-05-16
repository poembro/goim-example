package etcdv3

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/golang/glog"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
)

var etcdCli *clientv3.Client = nil

func newClient(etcdAddr string) (*clientv3.Client, error) {
	if etcdCli != nil {
		return etcdCli, nil
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:          strings.Split(etcdAddr, ","),
		DialTimeout:        time.Second * time.Duration(5),
		MaxCallSendMsgSize: 2 * 1024 * 1024,
	})

	etcdCli = cli
	return cli, err
}

// RegisterEtcd  注册服务
func RegisterEtcd(etcdAddr string, ttl int, env, appid, region, zone, ip, port string) (func(), error) {
	cli, err := newClient(etcdAddr)
	if err != nil {
		return nil, err
	}

	//lease
	ctx, cancel := context.WithCancel(context.Background())
	lease, err := cli.Grant(ctx, int64(ttl))
	if err != nil {
		cancel()
		return nil, err
	}

	key := fmt.Sprintf("/%s/%s/%s/%s/%s:%s", env, appid, region, zone, ip, port)
	value := fmt.Sprintf("%s:%s", ip, port)
	log.Infof("etcdv3 service register to etcd \"%s\" ", key)
	if _, err := cli.Put(ctx, key, value, clientv3.WithLease(lease.ID)); err != nil {
		cancel()
		return nil, err
	}
	keepAlive, err := cli.KeepAlive(ctx, lease.ID)
	if err != nil {
		cancel()
		return nil, err
	}

	go func() {
		for ka := range keepAlive { // keepAlive是1个channel
			if ka == nil {
				break
			}
			//fmt.Println("-->续约成功", ka)
		}
		log.Infof("etcdv3 %s \r\n", "关闭续租")
	}()

	closeEtcd := func() {
		_, _ = cli.Revoke(ctx, lease.ID)
		cancel()
		log.Infof("etcdv3 %s \r\n", "关闭etcd连接")
	}

	return closeEtcd, nil
}

func ResolverEtcd(etcdAddr string) {
	cli, err := newClient(etcdAddr)
	if err != nil {
		return
	}
	builder := &Builder{
		Client: cli,
	}
	resolver.Register(builder)
}

func DiscoveryEtcd(etcdAddr string, env, appid, region, zone string) map[string]string {
	ins := make(map[string]string)

	cli, err := newClient(etcdAddr)
	if err != nil {
		return ins
	}

	key := fmt.Sprintf("/%s/%s/%s", env, appid, region) // 服务发现 上海所有节点
	response, err := cli.Get(context.Background(), key, clientv3.WithPrefix())
	if err == nil {
		for _, kv := range response.Kvs {
			k := string(kv.Key)
			v := string(kv.Value)
			log.Infof("etcdv3 service  Discovery from etcd k:\"%s\"  v:\"%s\" ", k, v)
			ins[k] = v
		}
	}
	return ins
}
