# goim-demo
一个goim的demo 改服务发现为etcd 改kafka为redis 

## 特性 
 * 高性能
 * 纯Golang实现
 * 支持单个用户拥有多个设备
 * 支持推送至指定设备(该设备可以订阅多房间)
 * 支持广播推送至所有指定房间
 * 心跳支持（应用心跳和tcp、keepalive）
 * 支持安全验证（未授权用户不能订阅）
 * 多协议支持（websocket，tcp）
 * 可拓扑的架构（comet、job、logic模块可动态无限扩展）
 * 基于redis 、 kafka 发布订阅做异步消息推送(目前demo代码通过配置文件中consume.kafkaEnable、consume.redisEnable设置)


## 更改介绍

### 服务发现
- 改用etcd 
- 服务注册 规则按 /环境/服务名/地区/节点编号
- 服务发现 规则按 /环境/服务名/地区 (即 每个地区可以发现n个节点)

### kafka 改redis
- 目前demo代码支持redis/kafka 做中间件


## 部署
1.安装redis etcd golang1.17+ 略;
```sh 
$ git clone git@github.com:poembro/goim-demo.git
$ cd goim-demo
$ go mod tidy
$ make build
$ make runjob     ##运行 job 服务
$ make runlogic   ##运行 logic 服务
$ make runcomet   ##运行 comet 服务

$ cd examples/javascript/ && go run main.go   ## 运行http静态页面
$ cd test/ && go run tcp_client_testing.go 9999 100 192.168.84.168:3101   ## 运行100个并发测试脚本 

```


## 看源码过程中 各个结构图
[![](./goim.png)](https://github.com/poembro/goim-demo)


### 问题
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


### 文档
```

#### 推送至user_id它订阅的房间
> POST /goim/push/mids?operation=1001&mids=123 HTTP/1.1
> Host: 127.0.0.1:3111
> User-Agent: insomnia/2021.6.0-alpha.7
> Accept: */*
> Content-Length: 34

| 推送至user_id它订阅的房间

* upload completely sent off: 34 out of 34 bytes
* Mark bundle as not supporting multiuse

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Tue, 17 May 2022 06:51:47 GMT
< Content-Length: 23


#### 推送至device_id它订阅的房间
> POST /goim/push/keys?operation=1001&keys=123456123 HTTP/1.1
> Host: 127.0.0.1:3111
> User-Agent: insomnia/2021.6.0-alpha.7
> Accept: */*
> Content-Length: 23

| 11111111111111111111111

* upload completely sent off: 23 out of 23 bytes
* Mark bundle as not supporting multiuse

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Tue, 17 May 2022 06:52:22 GMT
< Content-Length: 23



#### 推送至room_id它订阅的房间

> POST /goim/push/room?operation=1001&type=live&room=1000 HTTP/1.1
> Host: 127.0.0.1:3111
> User-Agent: insomnia/2021.6.0-alpha.7
> Accept: */*
> Content-Length: 37

| 推送至room_id它订阅的房间444

* upload completely sent off: 37 out of 37 bytes
* Mark bundle as not supporting multiuse

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Tue, 17 May 2022 06:52:55 GMT
< Content-Length: 23




#### 推送至房间
> POST /goim/push/all?operation=1001&speed=1 HTTP/1.1
> Host: 127.0.0.1:3111
> User-Agent: insomnia/2021.6.0-alpha.7
> Accept: */*
> Content-Length: 22

|  1116666

* upload completely sent off: 22 out of 22 bytes
* Mark bundle as not supporting multiuse

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Tue, 17 May 2022 06:54:27 GMT
< Content-Length: 23




#### 统计多少房间多少人
> GET /goim/online/top?type=live&limit=10 HTTP/1.1
> Host: 127.0.0.1:3111
> User-Agent: insomnia/2021.6.0-alpha.7
> Accept: */*
> Content-Length: 12

| hello world 

* upload completely sent off: 12 out of 12 bytes
* Mark bundle as not supporting multiuse

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Tue, 17 May 2022 06:55:09 GMT
< Content-Length: 33
.



#### 某单个房间多少人 
> GET /goim/online/room?type=live&rooms=1000 HTTP/1.1
> Host: 127.0.0.1:3111
> User-Agent: insomnia/2021.6.0-alpha.7
> Accept: */*
> Content-Length: 12

| hello world 

* upload completely sent off: 12 out of 12 bytes
* Mark bundle as not supporting multiuse

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Tue, 17 May 2022 06:55:49 GMT
< Content-Length: 41

```
