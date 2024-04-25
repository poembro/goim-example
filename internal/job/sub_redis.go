package job

import (
	"context"

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
		PoolSize: 75,
		//MinIdleConns: c.Idle,
		//DialTimeout:  time.Duration(c.DialTimeout),
		//ReadTimeout:  time.Duration(c.ReadTimeout),
		//WriteTimeout: time.Duration(c.WriteTimeout),
	})
}

// Subscribe
func (j *Job) ConsumeRedis() error {
	ctx := context.TODO()
	pubsub := j.redis.Subscribe(ctx, j.c.Kafka.Topic)
	defer pubsub.Close()
	// 在发布任何内容之前，请等待确认已创建订阅
	_, err := pubsub.Receive(ctx)
	if err != nil {
		panic(err)
	}

	ch := pubsub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			if len(msg.Payload) <= 0 {
				continue
			}

			pushMsg := new(pb.PushMsg)
			if err := proto.Unmarshal([]byte(msg.Payload), pushMsg); err != nil {
				log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
				return err
			}

			log.Infoln("Subscribe message:", pushMsg)
			if err := j.Push(context.Background(), pushMsg); err != nil {
				log.Errorf("j.Push(%v) error(%v)", pushMsg, err)
				return err
			}
		}
	}

}
