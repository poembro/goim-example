package job

import (
	"sync"
	"time"

	"goim-demo/internal/job/conf"

	cluster "github.com/bsm/sarama-cluster"
	"github.com/gomodule/redigo/redis"

	"goim-demo/pkg/etcdv3"

	log "github.com/golang/glog"
)

// Job is push job.
type Job struct {
	c *conf.Config

	consumer     *cluster.Consumer
	cometServers map[string]*Comet

	rooms      map[string]*Room
	roomsMutex sync.RWMutex

	redis       *redis.Pool
	redisExpire int32
}

// New new a push job.
func New(c *conf.Config) *Job {
	j := &Job{
		c:           c,
		redis:       newRedis(c.Redis),
		redisExpire: int32(time.Duration(c.Redis.Expire) / time.Second),
		rooms:       make(map[string]*Room),
	}
	j.watchComet()

	j.subscribe()
	return j
}

func (j *Job) subscribe() {
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
	etcdAddr := j.c.Discovery.Nodes
	region := j.c.Env.Region
	zone := j.c.Env.Zone
	env := j.c.Env.DeployEnv
	appid := "goim.comet"

	go func() {
		for {
			ins := etcdv3.DiscoveryEtcd(etcdAddr, env, appid, region, zone)
			err := j.newAddress(ins)
			if err != nil {
				return
			}
			time.Sleep(time.Second * 10)
		}
	}()
}

func (j *Job) newAddress(ins map[string]string) error {
	comets := map[string]*Comet{}
	for _, grpcAddr := range ins {
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
