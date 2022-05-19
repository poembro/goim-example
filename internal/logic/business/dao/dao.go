package dao

import (
	"context"
	"time"

	"github.com/go-redis/redis"

	"goim-demo/internal/logic/conf"
)

// Dao dao
type Dao struct {
	c      *conf.Config
	RdsCli *redis.Client
	expire time.Duration
}

// newRedis 初始化Redis
func newRedis(c *conf.Redis) *redis.Client {
	cli := redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		DB:       0,
		Password: c.Auth,
	})

	_, err := cli.Ping().Result()
	if err != nil {
		panic(err)
	}
	return cli
}

// New init db.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:      c,
		RdsCli: newRedis(c.Redis),
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
