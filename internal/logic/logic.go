package logic

import (
	"context"

	"goim-demo/internal/logic/conf"
	"goim-demo/internal/logic/dao"
)

// Logic struct
type Logic struct {
	c *conf.Config
	//dis *naming.Discovery
	dao *dao.Dao
	// online
	totalIPs   int64
	totalConns int64
	roomCount  map[string]int32
	// load balancer
	//nodes        []*naming.Instance
	//loadBalancer *LoadBalancer
	regions map[string]string // province -> region
}

// New init
func New(c *conf.Config) (l *Logic) {
	l = &Logic{
		c:   c,
		dao: dao.New(c),
		//loadBalancer: NewLoadBalancer(),
		regions: make(map[string]string),
	}
	l.initRegions() //初始化regions属性 l.regions[上海] = sh

	return l
}

// Ping ping resources is ok.
func (l *Logic) Ping(c context.Context) (err error) {
	return l.dao.Ping(c)
}

// Close close resources.
func (l *Logic) Close() {
	l.dao.Close()
}

func (l *Logic) initRegions() {
	for region, ps := range l.c.Regions {
		for _, province := range ps {
			l.regions[province] = region
		}
	}
}
