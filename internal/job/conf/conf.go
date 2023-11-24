package conf

import (
	"flag"
	"os"
	"time"

	xtime "goim-demo/pkg/time"

	"github.com/BurntSushi/toml"
)

const (
	appId       = "goim.job"
	targetAppId = "goim.comet"
)

var (
	confPath  string
	region    string
	zone      string
	deployEnv string
	host      string
	// Conf config
	Conf *Config
)

//注意 注意  没有Addr字段
func init() {
	var (
		defHost, _ = os.Hostname()
	)
	flag.StringVar(&confPath, "conf", "job-example.toml", "default config path")
	flag.StringVar(&region, "region", os.Getenv("REGION"), "avaliable region. or use REGION env variable, value: sh etc.")
	flag.StringVar(&zone, "zone", os.Getenv("ZONE"), "avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.")
	flag.StringVar(&deployEnv, "deploy.env", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "host", defHost, "machine hostname. or use default machine hostname.")
}

// Init init config.
func Init() (err error) {
	Conf = Default()
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

// Default new a config with specified defualt value.
func Default() *Config {
	return &Config{
		Env: &Env{Region: region, AppId: appId, TargetAppId: targetAppId, Zone: zone, DeployEnv: deployEnv, Host: host},

		Consume: &Consume{
			KafkaEnable: false,
			RedisEnable: false,
		},
		Kafka: &Kafka{
			Topic:   "goim-push-topic",
			Group:   "goim-push-group-job",
			Brokers: []string{"10.0.41.145:9092"},
		},
		Comet: &Comet{RoutineChan: 1024, RoutineSize: 32},
		Room: &Room{
			Batch:  20,
			Signal: xtime.Duration(time.Second),
			Idle:   xtime.Duration(time.Minute * 15),
		},
	}
}

// Config is job config.
type Config struct {
	Env       *Env
	Kafka     *Kafka
	Consume   *Consume
	Discovery *Discovery
	Redis     *Redis
	Comet     *Comet
	Room      *Room
}

type Discovery struct {
	Nodes    string
	Username string
	Password string
}

type Consume struct {
	KafkaEnable bool
	RedisEnable bool
}

// Room is room config.
type Room struct {
	Batch  int
	Signal xtime.Duration
	Idle   xtime.Duration
}

// Comet is comet config.
type Comet struct {
	RoutineChan int
	RoutineSize int
}

// Kafka is kafka config.
type Kafka struct {
	Topic   string
	Group   string
	Brokers []string
}

// Env is env config.
type Env struct {
	AppId       string
	TargetAppId string
	Region      string
	Zone        string
	DeployEnv   string
	Host        string
}

// Redis .
type Redis struct {
	Network      string
	Addr         string
	Auth         string
	Active       int
	Idle         int
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
	IdleTimeout  xtime.Duration
	Expire       xtime.Duration
}
