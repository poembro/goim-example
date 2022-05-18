package dao

import (
	"context"
	"time"

	"github.com/go-redis/redis"

	"goim-demo/conf"
	"goim-demo/pkg/db"
)

// Dao dao
type Dao struct {
	c      *conf.Config
	RdsCli *redis.Client
	expire time.Duration
}

// New init db.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:      c,
		RdsCli: db.InitRedis(c.Global.RedisIP, c.Global.RedisPassword),
		expire: time.Duration(time.Second * 60), //75 * time.Second
	}

	return d
}

// Close  the resource.
func (d *Dao) Close() {
	if d.RdsCli != nil {
		d.RdsCli.Close()
	}
}

// Ping verify server is ok.
func (d *Dao) Ping(c context.Context) (err error) {
	if _, err = d.RdsCli.Ping().Result(); err != nil {
		return
	}

	return
}
