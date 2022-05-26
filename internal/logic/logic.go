package logic

import (
	"context"
	"time"

	"goim-demo/internal/logic/business"
	"goim-demo/internal/logic/conf"
	"goim-demo/internal/logic/dao"
	"goim-demo/internal/logic/model"

	"goim-demo/pkg/etcdv3"

	log "github.com/golang/glog"
)

const (
	_onlineTick     = time.Second * 10
	_onlineDeadline = time.Minute * 5
)

// Logic struct
type Logic struct {
	c        *conf.Config
	dao      *dao.Dao
	Business *business.Business
	// online
	totalIPs   int64
	totalConns int64
	roomCount  map[string]int32

	regions map[string]string // province -> region
}

// New init
func New(c *conf.Config) (l *Logic) {
	l = &Logic{
		c:        c,
		dao:      dao.New(c),
		Business: business.New(c), // 第三方业务
		regions:  make(map[string]string),
	}
	l.initRegions() //初始化regions属性 l.regions[上海] = sh
	go l.watchComet()
	return l
}

// Ping ping resources is ok.
func (l *Logic) Ping(c context.Context) (err error) {
	return l.dao.Ping(c)
}

// Close close resources.
func (l *Logic) Close() {
	l.dao.Close()
	l.Business.Close()
}

func (l *Logic) initRegions() {
	for region, ps := range l.c.Regions {
		for _, province := range ps {
			l.regions[province] = region
		}
	}
}

func (l *Logic) watchComet() {
	etcdAddr := l.c.Discovery.Nodes
	region := l.c.Env.Region
	zone := l.c.Env.Zone
	env := l.c.Env.DeployEnv
	appid := "goim.comet"
	dis := etcdv3.New(etcdAddr)
	for {
		time.Sleep(_onlineTick)
		ins := dis.GetService(env, appid, region, zone)
		if err := l.loadOnline(ins); err != nil {
			log.Errorf("watchComet error(%v)", err)
		}
	}
}

func (l *Logic) loadOnline(ins map[string]string) (err error) {
	var (
		roomCount = make(map[string]int32)
		online    *model.Online
	)
	for _, grpcAddr := range ins {
		online, err = l.dao.GetServerOnline(context.Background(), grpcAddr)
		if err != nil {
			return
		}
		if time.Since(time.Unix(online.Updated, 0)) > _onlineDeadline {
			_ = l.dao.DelServerOnline(context.Background(), grpcAddr)
			continue
		}
		for roomID, count := range online.RoomCount {
			roomCount[roomID] += count
		}
	}
	l.roomCount = roomCount
	return
}
