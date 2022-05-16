goim v2.0
==============
`Terry-Mao/goim` 是一个支持集群的im及实时推送服务。


```
遇到问题 一: 不同包路径下的 同名 api.proto 报错提示已经引入

luoyuxiangdeMacBook-Pro:goim-demo luoyuxiang$ make runjob
export GODEBUG=http2debug=2 && export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn && ./target/job -conf=target/job.toml -region=sh -zone=sh001 -deploy.env=prod -host=192.168.84.168 -log_dir=./target 
WARNING: proto: file "api.proto" is already registered
See https://developers.google.com/protocol-buffers/docs/reference/go/faq#namespace-conflict



解决
https://stackoverflow.com/questions/67693170/proto-file-is-already-registered-with-different-packages
 
方案 一:  go.mod 中 使用指定protobuf包版本
google.golang.org/protobuf v1.26.1-0.20210525005349-febffdd88e85 // indirect

方案二: 
编译后 启动时加参数 
export GODEBUG=http2debug=2 && export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn && ./target/job

```



---------------------------------------
  * [特性](#特性)
  * [安装](#安装)
  * [配置](#配置)
  * [例子](#例子)
  * [文档](#文档)
  * [集群](#集群)
  * [更多](#更多)

---------------------------------------

## 特性
 * 轻量级
 * 高性能
 * 纯Golang实现
 * 支持单个、多个、单房间以及广播消息推送
 * 支持单个Key多个订阅者（可限制订阅者最大人数）
 * 心跳支持（应用心跳和tcp、keepalive）
 * 支持安全验证（未授权用户不能订阅）
 * 多协议支持（websocket，tcp）
 * 可拓扑的架构（job、logic模块可动态无限扩展）
 * 基于Kafka做异步消息推送

## 安装
### 一、安装依赖
```sh
$ yum -y install java-1.7.0-openjdk
```

### 二、安装Kafka消息队列服务

kafka在官网已经描述的非常详细，在这里就不过多说明，安装、启动请查看[这里](http://kafka.apache.org/documentation.html#quickstart).

### 三、搭建golang环境
1.下载源码(根据自己的系统下载对应的[安装包](http://golang.org/dl/))
```sh
$ cd /data/programfiles
$ wget -c --no-check-certificate https://storage.googleapis.com/golang/go1.5.2.linux-amd64.tar.gz


wget  -c --no-check-certificate https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz
tar -zxvf go1.13.5.linux-amd64.tar.gz
mkdir /usr/local/golang
mv go /usr/local/golang/go-go1.13
rm -rf go1.13.5.linux-amd64.tar.gz
mkdir /data/web/main/golang/item -p
```

### 四、部署goim
1.下载goim及依赖包
```sh
$ yum install hg
$ go get -u goim-demo
$ mv $GOPATH/src/goim-demo $GOPATH/src/goim
$ cd $GOPATH/src/goim
$ go get ./...
```
 

### 五、启动goim
```sh 
vi /etc/profile
export GOROOT=/usr/local/golang/go-go1.13
export GOPATH=/data/web/main/golang/item
export PATH=$PATH:$GOROOT/bin
 
export REGION=sh  ##  地区, 例如, 中国区, 南美区, 北美区...
export ZONE=sh001 ##机器编号 可用区域, 例如中国区下的 gd 广东地区, sh 上海地区, 一般是指骨干 IDC 机房, 或者跨地区的逻辑区域, 这是同区内调度的主要划分点. 一般是同区内调度, 不会跨区调度
export DEPLOY_ENV=dev ##环境   再划分小一点的运行环境划分, 比如 Env = dev 开发环境, Env = trial 试商用
export ADDRS=192.168.3.222  ##当前机器ip
export WEIGHT=10     ##权重
export OFFLINE=false  ##
export DEBUG=true    ##debug 开关

source /etc/profile

iptables -F
nohup /data/app/kafka/bin/zookeeper-server-start.sh /data/app/kafka/config/zookeeper.properties &
nohup /data/app/kafka/bin/kafka-server-start.sh /data/app/kafka/config/server.properties &

cd /data/web/main/golang/item/src/github.com/bilibili/discovery
./cmd/discovery/discovery -conf cmd/discovery/discovery-example.toml  -log.dir="/data/web/main/golang/item/src/github.com/bilibili/discovery/tmp"

/usr/local/redis/src/redis-server /usr/local/redis/redis.conf

cd /data/web/main/golang/item/src/goim
./target/logic -conf=target/logic.toml
./target/comet -conf=target/comet.toml 
 ./target/job -conf=target/job.toml

cd examples/javascript/
go run main.go

```
如果启动失败，默认配置可通过查看panic-xxx.log日志文件来排查各个模块问题.

### 六、测试

推送协议可查看[push http协议文档](./docs/push.md)

## 配置

TODO

## 例子

Websocket: [Websocket Client Demo](https://goim-demo/tree/master/examples/javascript)

Android: [Android](https://github.com/roamdy/goim-sdk)

iOS: [iOS](https://github.com/roamdy/goim-oc-sdk)

## 文档
[push http协议文档](./docs/push.md)推送接口

## 集群

### comet

comet 属于接入层，非常容易扩展，直接开启多个comet节点，修改配置文件中的base节点下的server.id修改成不同值（注意一定要保证不同的comet进程值唯一），前端接入可以使用LVS 或者 DNS来转发

### logic

logic 属于无状态的逻辑层，可以随意增加节点，使用nginx upstream来扩展http接口，内部rpc部分，可以使用LVS四层转发

### kafka

kafka 可以使用多broker，或者多partition来扩展队列

### router

router 属于有状态节点，logic可以使用一致性hash配置节点，增加多个router节点（目前还不支持动态扩容），提前预估好在线和压力情况

### job

job 根据kafka的partition来扩展多job工作方式，具体可以参考下kafka的partition负载

##更多
TODO

 

 一个用户上线后 redis 分别写入如下
 1577077537.899076 [0 192.168.3.222:56839] "HSET" "mid_123" "6d2d1fd6-7e0a-44b8-8b02-5fe109e0326e" "192.168.3.222"
1577077537.899437 [0 192.168.3.222:56839] "EXPIRE" "mid_123" "1800"

1577077537.899723 [0 192.168.3.222:56839] "SET" "key_6d2d1fd6-7e0a-44b8-8b02-5fe109e0326e" "192.168.3.222"
1577077537.899949 [0 192.168.3.222:56839] "EXPIRE" "key_6d2d1fd6-7e0a-44b8-8b02-5fe109e0326e" "1800"

1577077540.519876 [0 192.168.3.222:56839] "HSET" "ol_192.168.3.222" "43" "{\"server\":\"192.168.3.222\",\"room_count\":{\"live://1000\":1},\"updated\":1577077540}"
1577077540.519965 [0 192.168.3.222:56839] "EXPIRE" "ol_192.168.3.222" "1800"

1577077630.543396 [0 192.168.3.222:56839] "HSET" "ol_192.168.3.222" "43" "{\"server\":\"192.168.3.222\",\"room_count\":{\"live://1000\":1},\"updated\":1577077630}"
1577077630.543426 [0 192.168.3.222:56839] "EXPIRE" "ol_192.168.3.222" "1800"

1577077630.543396 [0 192.168.3.222:56839] "HSET" "ol_192.168.3.222" "43" "{\"server\":\"192.168.3.222\",\"room_count\":{\"live://1000\":1},\"updated\":1577077630}"
1577077630.543426 [0 192.168.3.222:56839] "EXPIRE" "ol_192.168.3.222" "1800"

...

 一个用户离线后 redis 分别写入如下
1577077780.454623 [0 192.168.3.222:56839] "HDEL" "mid_123" "6d2d1fd6-7e0a-44b8-8b02-5fe109e0326e"
1577077780.454670 [0 192.168.3.222:56839] "DEL" "key_6d2d1fd6-7e0a-44b8-8b02-5fe109e0326e"

 


问题：
  分区部署是comet节点和job节点分区就近部署，logic节点布在中心区域这样部署的吗？你们在实践过程中推荐怎么分区部署？谢谢

回答：
  我们现在都是使用cloud方案，comet可以部署在全球边缘节点，然后利用cloud vpc专线打通；
  比如：
  核心部署（logic、job、disocvery、kafka/zk）：
  region=sh
  zone=sh001

  边缘节点（comet）：
  国内节点：
  北京、上海、广州、四川
  国外节点：
  香港、日本、硅谷、法兰克福（分别覆盖日韩、欧洲、北美、澳大利亚等地区）


看过代码后我的理解:
  -->>这里说的是 comet 部署到任意地方 但是要注意 -zone=sh001 (每台机器都要有 机器标识 或者 服务标识)
	nohup target/comet -conf=target/comet.toml -region=sh -zone=sh001 deploy.env=dev weight=10 addrs=192.168.3.222 debug=true 2>&1 > target/comet.log &
  
  -->>与之关联  每个comet服务或者机器标识,都必须对应一个job -zone=sh001 (并且每台机器的消费组 "goim-push-group-job-sh001" 不一样)
	nohup target/job -conf=target/job.toml -region=sh -zone=sh001 deploy.env=dev 2>&1 > target/job.log &  //注意这里没有Addrs参数 只需要-zone标识对应上就ok
 
  -->> logic 服务不与任何服务关联 (但是grpc kafka 两块要能通)


  #####案例######
  nohup target/logic -conf=target/logic.toml -region=sh -zone=sh001 deploy.env=dev weight=10 addrs=192.168.3.222 2>&1 > target/logic.log &
  nohup target/comet -conf=target/comet.toml -region=sh -zone=sh001 deploy.env=dev weight=10 addrs=192.168.3.222 debug=true 2>&1 > target/comet.log &

  nohup target/job -conf=target/job.toml -region=sh -zone=sh001 deploy.env=dev 2>&1 > target/job.log &
  nohup target/job -conf=target/job111.toml -region=sh -zone=sh002 deploy.env=dev debug=true 2>&1 > target/job111.log &  //注意这里改了"消费组"参数 和命令行上的标识// 于是可以实现job comet部署在任意地方  不相关联
  nohup target/job -conf=target/job100.toml -region=sh -zone=sh003 deploy.env=dev debug=true 2>&1 > target/job100.log &  //注意这里改了"消费组"参数 和命令行上的标识// 于是可以实现job comet部署在任意地方  不相关联
  -->> 注意开一台job为啥不能供应所有comet，答案在 job.go 第115行  ins := insMap[j.c.Env.Zone]



kafka topic 建议16个partition

