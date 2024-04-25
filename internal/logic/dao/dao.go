package dao

import (
	"context"
	"time"

	"goim-example/internal/logic/conf"

	"github.com/Shopify/sarama"
	"github.com/redis/go-redis/v9"
)

// Dao dao.
type Dao struct {
	c           *conf.Config
	kafkaPub    sarama.SyncProducer
	redis       *redis.Client
	redisExpire time.Duration
}

// New new a dao and return.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:           c,
		redis:       newRedis(c.Redis),
		redisExpire: time.Duration(c.Redis.Expire),
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
	// sarama.NewRandomPartitioner // 随机分配分区 partition
	// sarama.NewRoundRobinPartitioner // 轮换选择分区
	// sarama.NewHashPartitioner // 通过设置 hash-key 自动 hash 分区，如果没有设置key，则随机选取 msg.Key = sarama.StringEncoder("topic1key")
	// sarama.NewManualPartitioner // 人工指定分区 msg.Partition = 0       // 人工指定分区

	// Kafka客户端会根据Key进行Hash，我们通过把接收用户ID作为Key，这样就能让所有发给某个人的消息落到同一个分区了，也就有序了。
	kc.Producer.Partitioner = sarama.NewHashPartitioner // 设置选择分区的策略为Hash
	kc.Producer.RequiredAcks = sarama.WaitForAll        // 等待所有follower都回复ack，确保Kafka不会丢消息
	kc.Producer.Retry.Max = 10                          // 最多重试10次以生成消息
	kc.Producer.Return.Successes = true
	pub, err := sarama.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return pub
}

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
	/*
		&redis.Pool{
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
	*/
}

// Close close the resource.
func (d *Dao) Close() error {
	if d.c.Consume.KafkaEnable {
		d.kafkaPub.Close()
	}

	d.redis.Close()

	return nil
}

// Ping dao ping.
func (d *Dao) Ping(c context.Context) error {
	return d.pingRedis(c)
}
