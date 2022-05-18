package comet

import (
	"sync"

	"goim-demo/api/protocol"
	"goim-demo/internal/comet/errors"
	"goim-demo/pkg/bufio"
)

// Channel used by message pusher send msg to write goroutine.
type Channel struct {
	Room     *Room
	CliProto Ring
	signal   chan *protocol.Proto
	Writer   bufio.Writer
	Reader   bufio.Reader
	Next     *Channel
	Prev     *Channel

	Mid      int64  //mid 就是用户id memberID
	Key      string //uuid 用户与comet建立长连接key做标识  比如im信息要发给的某用户的多个端
	IP       string
	watchOps map[int32]struct{} //其中的 int32 是房间号. map 多个房间号, map 结构是用来查询房间号是否在 map 中存在. watchOps 是当前长连接用户用来监听当前客户端接收哪个房间的 im 消息推送, 换个方式说, 一个 goim 终端可以接收多个房间发送来的 im 消息
	mutex    sync.RWMutex
}

// NewChannel 初始化是在 tcp / websocket ServeWebsocket方法 进行首次连接时处理的,
func NewChannel(cli, svr int) *Channel {
	c := new(Channel)
	c.CliProto.Init(cli)                       //cli为 5
	c.signal = make(chan *protocol.Proto, svr) //svr 为10
	c.watchOps = make(map[int32]struct{})
	return c
}

// Watch 在auth完后被执行过 表示这个人可以接受哪些房间的消息
func (c *Channel) Watch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		c.watchOps[op] = struct{}{}
	}
	c.mutex.Unlock()
}

// UnWatch unwatch an operation
func (c *Channel) UnWatch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		delete(c.watchOps, op)
	}
	c.mutex.Unlock()
}

// NeedPush verify if in watch. 验证这个人是否可以接受对应房间的消息
func (c *Channel) NeedPush(op int32) bool {
	c.mutex.RLock()
	if _, ok := c.watchOps[op]; ok {
		c.mutex.RUnlock()
		return true
	}
	c.mutex.RUnlock()
	return false
}

// Push server push message.
func (c *Channel) Push(p *protocol.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
		err = errors.ErrSignalFullMsgDropped
	}
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() *protocol.Proto {
	return <-c.signal
}

// Signal 对应server_tcp.go
func (c *Channel) Signal() {
	c.signal <- protocol.ProtoReady
}

// Close close the channel.
func (c *Channel) Close() {
	c.signal <- protocol.ProtoFinish
}
