package job

import (
	"context"
	"fmt"
	"time"

	pb "goim-example/api/logic"
	"goim-example/internal/job/conf"

	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
)

func newRedis(c *conf.Redis) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.Idle,
		MaxActive:   c.Active,
		IdleTimeout: time.Duration(c.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(c.Network, c.Addr,
				redis.DialConnectTimeout(time.Duration(c.DialTimeout)),
				// redis.DialReadTimeout(time.Duration(c.ReadTimeout)),
				redis.DialWriteTimeout(time.Duration(c.WriteTimeout)),
				redis.DialPassword(c.Auth),
			)
			if err != nil {
				log.Infoln(err)
				return nil, err
			}
			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

// Subscribe
func (j *Job) ConsumeRedis() error {
	psc := redis.PubSubConn{Conn: j.redis.Get()}

	channel := []string{"channel", j.c.Kafka.Topic}
	if err := psc.Subscribe(redis.Args{}.AddFlat(channel)...); err != nil {
		return err
	}
	consume := func(msg redis.Message) error {
		pushMsg := new(pb.PushMsg)
		if err := proto.Unmarshal(msg.Data, pushMsg); err != nil {
			log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
			return err
		}

		log.Infoln("Subscribe message:", pushMsg)
		if err := j.Push(context.Background(), pushMsg); err != nil {
			log.Errorf("j.Push(%v) error(%v)", pushMsg, err)
			return err
		}
		return nil
	}
	// health check
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	// start a new goroutine to receive message
	go func() {
		// 一个连接可以在不同的 goroutine 并发调用 Receive() 和 Subscribe()（subscribe调用了send和flush） ，
		// 但是却不能再有其他并发操作（比如 Close()）
		defer psc.Close()
		for {
			switch msg := psc.Receive().(type) { //连接配置中 要取消读取超时配置 (默认 读取不会超时)
			case error:
				done <- fmt.Errorf("redis pubsub receive err: %v", msg)
				cancel()
				return
			case redis.Message:
				if err := consume(msg); err != nil {
					done <- err
					return
				}
			case redis.Subscription:
				log.Infoln("redis Subscription:", msg)
				if msg.Count == 0 {
					// all channels are unsubscribed
					done <- nil
					return
				}
			}
		}
	}()

	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := psc.Unsubscribe(); err != nil {
				log.Infof("redis pubsub unsubscribe err: %s", err.Error())
				return err
			}
			return nil
		case err := <-done:
			log.Infof("redis pubsub unsubscribe  done  err: %s", err.Error())
			return err
		case <-tick.C:
			if err := psc.Ping(""); err != nil {
				return err
			}
		}
	}

}
