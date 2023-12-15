package conf

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	xtime "goim-example/pkg/time"

	"github.com/BurntSushi/toml"
)

const (
	appId       = "goim.comet"
	targetAppId = "goim.logic"
)

var (
	confPath  string
	region    string
	zone      string
	deployEnv string
	host      string
	addrs     string
	weight    int64
	offline   bool
	debug     bool

	// Conf config
	Conf *Config
)

func init() {
	var (
		defHost, _    = os.Hostname()
		defAddrs      = os.Getenv("ADDRS")
		defWeight, _  = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
		defOffline, _ = strconv.ParseBool(os.Getenv("OFFLINE"))
		defDebug, _   = strconv.ParseBool(os.Getenv("DEBUG"))
	)
	flag.StringVar(&confPath, "conf", "comet-example.toml", "default config path.")
	flag.StringVar(&region, "region", os.Getenv("REGION"), "avaliable region. or use REGION env variable, value: sh etc.")
	flag.StringVar(&zone, "zone", os.Getenv("ZONE"), "avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.")
	flag.StringVar(&deployEnv, "deploy.env", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "host", defHost, "machine hostname. or use default machine hostname.")
	flag.StringVar(&addrs, "addrs", defAddrs, "server public ip addrs. or use ADDRS env variable, value: 127.0.0.1 etc.")
	flag.Int64Var(&weight, "weight", defWeight, "load balancing weight, or use WEIGHT env variable, value: 10 etc.")
	flag.BoolVar(&offline, "offline", defOffline, "server offline. or use OFFLINE env variable, value: true/false etc.")
	flag.BoolVar(&debug, "debug", defDebug, "server debug. or use DEBUG env variable, value: true/false etc.")
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
		Debug: debug,
		Env: &Env{
			DeployEnv:   deployEnv, // 环境 如:dev/fat1/uat/pre/prod
			TargetAppId: targetAppId,
			AppId:       appId,
			Region:      region,                    // 地区 如:sh
			Zone:        zone,                      // 空间 如:sh001
			Host:        host,                      // 主机名 如:localhost / sf
			Weight:      weight,                    // 权重 如:10
			Addrs:       strings.Split(addrs, ","), // 公网ip 如:192.168.84.168,192.168.84.169
			Offline:     offline,                   // 在线状态 如:true/false
		},
		//Discovery: &Discovery{
		//	Nodes: "http://10.0.41.145:2379,http://10.0.41.145:2479,http://10.0.41.145:2579",
		//},
		RPCClient: &RPCClient{
			Dial:    xtime.Duration(time.Second),
			Timeout: xtime.Duration(time.Second),
		},
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              ":3109",
			Timeout:           xtime.Duration(time.Second),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
		TCP: &TCP{
			Bind:         []string{":3101"},
			Sndbuf:       4096,
			Rcvbuf:       4096,
			KeepAlive:    false,
			Reader:       32,
			ReadBuf:      1024,
			ReadBufSize:  8192,
			Writer:       32,
			WriteBuf:     1024,
			WriteBufSize: 8192,
		},
		Websocket: &Websocket{
			Bind: []string{":3102"},
		},
		Protocol: &Protocol{
			Timer:            32,
			TimerSize:        2048,
			CliProto:         5,
			SvrProto:         10,
			HandshakeTimeout: xtime.Duration(time.Second * 5),
		},
		Bucket: &Bucket{
			Size:          32,   //b 表示创建32个桶  Bucket切片长度  即：每个 设备id 哈希32 选择1个桶
			Channel:       1024, //b.chs map[deviceId]*Channel 容量为 1024  即: 每个设备id 连接后创建的 Channel结构 会放入这个对应map
			Room:          1024, //b.rooms map[roomId]*Room 容量为1024  即: 每个设备id 所在房间号 会放入这个对应map
			RoutineAmount: 32,   //开启32个groutine去等 广播至房间的 消息
			RoutineSize:   1024, //b.routines[RoutineAmount] 广播至房间的 消息通道是有缓存通道 1024
		},
	}
}

// Config is comet config.
type Config struct {
	Debug     bool
	Env       *Env
	Discovery *Discovery
	TCP       *TCP
	Websocket *Websocket
	Protocol  *Protocol
	Bucket    *Bucket
	RPCClient *RPCClient
	RPCServer *RPCServer
	Whitelist *Whitelist
}

// Env is env config.
type Env struct {
	DeployEnv   string
	AppId       string
	TargetAppId string
	Region      string
	Zone        string
	Host        string
	Weight      int64
	Offline     bool
	Addrs       []string
}

type Discovery struct {
	Nodes    string
	Username string
	Password string
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    xtime.Duration
	Timeout xtime.Duration
}

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           xtime.Duration
	IdleTimeout       xtime.Duration
	MaxLifeTime       xtime.Duration
	ForceCloseWait    xtime.Duration
	KeepAliveInterval xtime.Duration
	KeepAliveTimeout  xtime.Duration
}

// TCP is tcp config.
type TCP struct {
	Bind         []string
	Sndbuf       int
	Rcvbuf       int
	KeepAlive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// Websocket is websocket config.
type Websocket struct {
	Bind        []string
	TLSOpen     bool
	TLSBind     []string
	CertFile    string
	PrivateFile string
}

// Protocol is protocol config.
type Protocol struct {
	Timer            int
	TimerSize        int
	SvrProto         int
	CliProto         int
	HandshakeTimeout xtime.Duration
}

// Bucket is bucket config.
type Bucket struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

// Whitelist is white list config.
type Whitelist struct {
	Whitelist []int64 // 在配置文件中写 mid 则该mid连接后会打日志
	WhiteLog  string  // 且日志被打印在该文件中
}
