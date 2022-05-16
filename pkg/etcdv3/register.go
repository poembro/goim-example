package etcdv3

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
)

func newClient(etcdAddr string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:          strings.Split(etcdAddr, ","),
		DialTimeout:        time.Second * time.Duration(5),
		MaxCallSendMsgSize: 2 * 1024 * 1024,
	})
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
	fmt.Println("---服务注册至etcd----->", key)
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
		fmt.Printf("2022-02-25 09:44:38.690	DEBUG	etcdv3/register.go:73	%s \r\n", "关闭续租")
	}()

	closeEtcd := func() {
		_, _ = cli.Revoke(ctx, lease.ID)
		cancel()
		fmt.Printf("2022-02-25 09:44:38.690	DEBUG	etcdv3/register.go:79	%s \r\n", "关闭etcd连接")
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
			fmt.Printf("--发现服务---->k:  %s   v:  %s \r\n ", k, v)
			ins[k] = v
		}
	}
	return ins
}
