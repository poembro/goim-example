package dao

import (
	"context"
	"time"

	"goim-demo/internal/logic/conf"

	"github.com/Shopify/sarama"
	"github.com/gomodule/redigo/redis"
)

// Dao dao.
type Dao struct {
	c           *conf.Config
	kafkaPub    sarama.SyncProducer
	redis       *redis.Pool
	redisExpire int32
}

// New new a dao and return.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:           c,
		redis:       newRedis(c.Redis),
		redisExpire: int32(time.Duration(c.Redis.Expire) / time.Second),
	}

	// 初始化 kafka 连接
	if c.Consume.KafkaEnable {
		d.kafkaPub = newKafkaPub(c.Kafka)
	}
	return d
}

func newKafkaPub(c *conf.Kafka) sarama.SyncProducer {
	var err error
	kc := sarama.NewConfig()
	kc.Version = sarama.V2_8_1_0
	kc.Producer.Partitioner = sarama.NewHashPartitioner
	kc.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	kc.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	kc.Producer.Return.Successes = true
	pub, err := sarama.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return pub
}

func newRedis(c *conf.Redis) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.Idle,
		MaxActive:   c.Active,
		IdleTimeout: time.Duration(c.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(c.Network, c.Addr,
				redis.DialConnectTimeout(time.Duration(c.DialTimeout)),
				redis.DialReadTimeout(time.Duration(c.ReadTimeout)),
				redis.DialWriteTimeout(time.Duration(c.WriteTimeout)),
				redis.DialPassword(c.Auth),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
}

// Close close the resource.
func (d *Dao) Close() error {
	return d.redis.Close()
}

// Ping dao ping.
func (d *Dao) Ping(c context.Context) error {
	return d.pingRedis(c)
}
