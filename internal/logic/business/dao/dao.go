package dao

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"goim-example/internal/logic/conf"
)

// Dao dao
type Dao struct {
	c      *conf.Config
	RDSCli *redis.Client
	expire time.Duration
}

// newRedis 初始化Redis
func newRedis(c *conf.Redis) *redis.Client {
	cli := redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		DB:           0,
		Password:     c.Auth,
		PoolSize:     75,
		MinIdleConns: 50,
	})

	return cli
}

// New init db.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:      c,
		RDSCli: newRedis(c.Redis),
		expire: time.Duration(time.Minute * 1800), //75 * time.Second
	}

	return d
}

// Close  the resource.
func (d *Dao) Close() {
	if d.RDSCli != nil {
		d.RDSCli.Close()
	}
}

// Ping verify server is ok.
func (d *Dao) Ping(c context.Context) (err error) {
	if _, err = d.RDSCli.Ping(c).Result(); err != nil {
		return
	}

	return
}
