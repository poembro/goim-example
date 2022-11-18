package main

// Start Commond eg: ./client 1 1000 localhost:3101
// first parameter：beginning userId
// second parameter: amount of clients
// third parameter: comet server ip

import (
	"goim-demo/internal/logic/business/util"
	"goim-demo/pkg/bufio"
	"sync"

	"flag"
	"fmt"
	"goim-demo/api/protocol"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	log "github.com/golang/glog"
)

const (
	opHeartbeat      = int32(2)
	opHeartbeatReply = int32(3)
	opAuth           = int32(7)
	opAuthReply      = int32(8)

	rawHeaderLen = uint16(16)
	heart        = 30 * time.Second
)

var (
	countDown  int64
	aliveCount int64
)

var FdMutex sync.Mutex

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	begin, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	go result()
	for i := begin; i < begin+num; i++ {
		n := int64(i)
		go func(mid int64) {
			for {
				startClient(mid)
				fmt.Println("有报错重连....")
			}
		}(n)
	}
	// signal
	var exit chan bool
	<-exit
}

func result() {
	var (
		lastTimes int64
		interval  = int64(30)
	)
	for {
		nowCount := atomic.LoadInt64(&countDown)
		nowAlive := atomic.LoadInt64(&aliveCount)
		diff := nowCount - lastTimes
		lastTimes = nowCount
		fmt.Println(fmt.Sprintf("%s 活跃连接:%d down:%d down/s:%d", time.Now().Format("2006-01-02 15:04:05"), nowAlive, nowCount, diff/interval))
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func startClient(mid int64) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	atomic.AddInt64(&aliveCount, 1)
	quit := make(chan bool, 1)
	defer func() {
		close(quit)
		atomic.AddInt64(&aliveCount, -1)
	}()
	// connnect to server
	conn, err := net.Dial("tcp", os.Args[3])
	if err != nil {
		log.Errorf("net.Dial(%s) error(%v)", os.Args[3], err)
		return
	}

	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)

	deviceId := util.Md5(fmt.Sprintf("%s_%d", "web", mid))
	f := fmt.Sprintf(`{"mid":"%d","key":"%s", "room_id":"live://1000", "platform":"web", "accepts":[1000,1001,1002]}`, mid, deviceId)

	proto := new(protocol.Proto)
	proto.Ver = 1
	proto.Op = 7
	proto.Seq = 222
	proto.Body = []byte(f)
	if err = tcpWriteProto(wr, proto); err != nil {
		log.Errorf("tcpWriteProto() error(%v)", err)
		return
	}
	if err = proto.ReadTCP(rd); err != nil {
		log.Errorf("tcpReadProto() error(%v)", err)
		return
	}
	fmt.Printf("mid:%d auth ok, proto: %v \r\n", mid, proto)

	// writer
	go func() {
		for {
			p := new(protocol.Proto)
			p.Ver = 1
			p.Op = 2
			p.Seq = 111
			if err = tcpWriteProto(wr, p); err != nil {
				log.Errorf("mid:%d tcpWriteProto() error(%v)", mid, err)
				return
			}
			fmt.Printf("mid:%d Write heartbeat \r\n", mid)
			time.Sleep(heart)

			select {
			case <-quit:
				return
			default:
			}
		}
	}()

	// reader
	for {
		pr := new(protocol.Proto)
		if err = pr.ReadTCP(rd); err != nil {
			log.Errorf("mid:%d tcpReadProto() error(%v)", mid, err)
			quit <- true
			return
		}
		if pr.Op == opAuthReply {
			fmt.Printf("mid:%d auth success \r\n", mid)
		} else if pr.Op == opHeartbeatReply {
			fmt.Printf("mid:%d receive heartbeat \r\n", mid)
			// 设置读取超时
			//golang的标准网络库是最后期限方式  (平常linux 是空闲超时)
			if err = conn.SetReadDeadline(time.Now().Add(heart + 60*time.Second)); err != nil {
				log.Errorf("conn.SetReadDeadline() error(%v)", err)
				quit <- true
				return
			}
		} else {
			fmt.Printf("mid:%d op:%d msg: %s \r\n", mid, pr.Op, string(pr.Body))
			atomic.AddInt64(&countDown, 1)
		}
	}
}

func tcpWriteProto(wr *bufio.Writer, proto *protocol.Proto) (err error) {
	FdMutex.Lock()
	defer FdMutex.Unlock()

	// write
	err = proto.WriteTCP(wr)

	//fmt.Printf("发送协议包: %#v 缓冲中已使用的字节数 %d \r\n", proto.Op, wr.Buffered())
	//fmt.Println(p)
	//fmt.Println("缓冲中还有多少字节未使用。:", wr.Available())         //3827

	err = wr.Flush()
	return
}
