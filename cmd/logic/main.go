package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	"goim-example/internal/business"
	"goim-example/internal/logic"
	"goim-example/internal/logic/conf"
	"goim-example/internal/logic/grpc"
	"goim-example/pkg/etcdv3"

	//"goim-example/internal/logic/user"  //加的业务

	log "github.com/golang/glog"
)

const (
	ver = "2.0.0"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Infof("goim-logic [version: %s env: %+v] start", ver, conf.Conf.Env)

	// logic
	srv := logic.New(conf.Conf)
	httpSrv := business.New(conf.Conf, srv)
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)
	//可以在此 追加业务代码  抄grpc目录 然后目录下做 业务认证逻辑

	// discovery
	dis := etcdv3.New(conf.Conf.Discovery.Nodes, conf.Conf.Discovery.Username, conf.Conf.Discovery.Password)
	Register(dis, conf.Conf.RPCServer.Addr, conf.Conf.Env)

	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			dis.Deregister() // 移除 etcd 中的节点
			srv.Close()
			httpSrv.Close()
			rpcSrv.GracefulStop()
			log.Infof("goim-logic [version: %s] exit", ver)
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

// Register 服务注册
func Register(dis *etcdv3.Registry, node string, c *conf.Env) error {
	// 当前grpc 服务的 外网ip 端口
	_, port, _ := net.SplitHostPort(node)
	env := c.DeployEnv
	appid := c.AppId
	region := c.Region
	zone := c.Zone
	ip := c.Host
	// 服务注册至ETCD
	return dis.Register(env, appid, region, zone, ip, port)
}
