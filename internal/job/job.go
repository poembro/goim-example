package job

import (
	"sync"
	"time"

	"goim-example/internal/job/conf"

	"github.com/Shopify/sarama"
	"github.com/redis/go-redis/v9"

	"goim-example/pkg/etcdv3"

	log "github.com/golang/glog"
)

// Job is push job.
type Job struct {
	c *conf.Config

	consumer     sarama.ConsumerGroup
	cometServers map[string]*Comet

	rooms      map[string]*Room
	roomsMutex sync.RWMutex

	redis       *redis.Client
	redisExpire time.Duration
}

// New new a push job.
func New(c *conf.Config) *Job {
	j := &Job{
		c:           c,
		redis:       newRedis(c.Redis),
		redisExpire: time.Duration(c.Redis.Expire),
		rooms:       make(map[string]*Room),
	}
	j.watchComet()

	j.Consume() // 消费端初始化
	return j
}

func (j *Job) Consume() {
	if j.c.Consume.KafkaEnable {
		j.consumer = newKafkaSub(j.c.Kafka)
		go j.ConsumeKafka()
	}

	if j.c.Consume.RedisEnable {
		go j.ConsumeRedis()
	}
}

// Close close the resounces.
func (j *Job) Close() error {
	if j.consumer != nil {
		j.consumer.Close()
	}

	if j.redis != nil {
		j.redis.Close()
	}
	return nil
}

func (j *Job) watchComet() {
	env := j.c.Env.DeployEnv
	appid := j.c.Env.TargetAppId // 目标服务
	region := j.c.Env.Region
	zone := j.c.Env.Zone

	nodes := j.c.Discovery.Nodes
	username := j.c.Discovery.Username
	password := j.c.Discovery.Password
	dis := etcdv3.New(nodes, username, password)
	go func() {
		for {
			items := dis.LoadOnlineNodes(env, appid, region, zone)
			err := j.newAddress(items)
			if err != nil {
				return
			}
			time.Sleep(time.Second * 10)
		}
	}()
}

func (j *Job) newAddress(items map[string]string) error {
	comets := map[string]*Comet{}
	for _, grpcAddr := range items {
		if old, ok := j.cometServers[grpcAddr]; ok {
			comets[grpcAddr] = old
			continue
		}

		c, err := NewComet(grpcAddr, j.c.Comet)
		if err != nil {
			log.Errorf("watchComet NewComet(%+v) error(%v)", grpcAddr, err)
			return err
		}
		comets[grpcAddr] = c
		log.Infof("watchComet AddComet grpc:%+v", grpcAddr)
	}

	for key, old := range j.cometServers {
		if _, ok := comets[key]; !ok {
			old.cancel()
			log.Infof("watchComet DelComet:%s", key)
		}
	}
	j.cometServers = comets
	return nil
}
