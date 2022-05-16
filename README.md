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
 * 基于redis 发布订阅做异步消息推送(目前demo代码为了阅读简洁暂采用redis，生产环境可更换为 kafka 等)


## 更改介绍

### 服务发现
- 改用etcd 
- 服务注册 规则按 /环境/服务名/地区/节点编号
- 服务发现 规则按 /环境/服务名/地区 (即 每个地区可以发现n个节点)

### kafka 改redis
- 目前demo代码为了阅读简洁暂采用redis发布订阅


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

```





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