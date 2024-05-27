package job

import (
	"context"
	"errors"
	"strings"
	"time"

	pb "goim-example/api/logic"
	"goim-example/internal/job/conf"

	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	"github.com/redis/go-redis/v9"
)

func newRedis(c *conf.Redis) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		DB:       0,
		Password: c.Auth,
		//PoolSize: 75,
		//MinIdleConns: c.Idle,
		//DialTimeout:  time.Duration(c.DialTimeout),
		//ReadTimeout:  time.Duration(c.ReadTimeout),
		//WriteTimeout: time.Duration(c.WriteTimeout),
	})
}

// Subscribe
func (j *Job) ConsumeRedis() error {
	for {
		select {
		default:
			values, err := j.redis.BRPop(context.TODO(), time.Second*5, j.c.Kafka.Topic).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					continue
				}

				if strings.Contains(err.Error(), "timeout") {
					continue
				}

				log.Errorf("ConsumeRedis  error(%v)", err)
				return err
			}

			if len(values) < 2 {
				continue
			}
			msg := values[1]
			pushMsg := new(pb.PushMsg)
			if err := proto.Unmarshal([]byte(msg), pushMsg); err != nil {
				log.Errorf("ConsumeRedis proto.Unmarshal(%v) error(%v)", msg, err)
				return err
			}

			log.Infoln("ConsumeRedis message:", pushMsg)
			if err := j.Push(context.Background(), pushMsg); err != nil {
				log.Errorf("ConsumeRedis j.Push(%v) error(%v)", pushMsg, err)
				return err
			}
		}
	}
}
