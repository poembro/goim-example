package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"goim-demo/api/comet/grpc"
	"goim-demo/internal/logic/model"
	"time"

	log "github.com/golang/glog"

	//"github.com/google/uuid"
	"crypto/md5"
)

// Connect connected a conn.
func (l *Logic) Connect(c context.Context, server, cookie string, token []byte) (mid int64, key, roomID string, accepts []int32, hb int64, err error) {
	var params struct {
		Mid      int64   `json:"mid"`
		Key      string  `json:"key"`
		RoomID   string  `json:"room_id"`
		Platform string  `json:"platform"`
		Accepts  []int32 `json:"accepts"`
	}
	if err = json.Unmarshal(token, &params); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return
	}
	mid = params.Mid
	roomID = params.RoomID
	accepts = params.Accepts
	key = params.Key
	hb = int64(l.c.Node.Heartbeat) * int64(l.c.Node.HeartbeatMax)

	//自行处理验证 start
	//验证 key 值
	if mid == 0 {
		err = fmt.Errorf("mid is err !!")
		return
	}
	if key == "" {
		err = fmt.Errorf("key is err !!")
		return
	}

	fomatKey := fmt.Sprintf("%d-%s-%s", mid, params.Platform, roomID)
	byteFomatKey := []byte(fomatKey)
	has := md5.Sum(byteFomatKey)
	md5key := fmt.Sprintf("%x", has) //将[]byte转成16进制
	//fmt.Printf("fomatKey:%s  ==>   key:%s    md5key:%s \r\n",fomatKey, key, md5key)
	if key != md5key {
		//err = fmt.Errorf("auth key is err !!")
		//return
	}

	//验证是否已经在线
	/*
		keyOnlines, _ := l.dao.ServersByKeys(c, []string{key})
		for k, val := range keyOnlines {
			fmt.Printf("key: %v  val: %v \r\n", k, val)
			if val != "" {
				err = fmt.Errorf("already has a login !!")
				return
			}
		}
	*/
	//自行处理验证 end

	//如果验证通过, 则生成会话数据, 存入 redis 中; 否则返回认证失败
	if err = l.dao.AddMapping(c, mid, key, server); err != nil {
		log.Errorf("l.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
	}

	log.Infof("conn connected key:%s server:%s mid:%d token:%s", key, server, mid, token)
	return
}

// Disconnect disconnect a conn.
func (l *Logic) Disconnect(c context.Context, mid int64, key, server string) (has bool, err error) {
	if has, err = l.dao.DelMapping(c, mid, key, server); err != nil {
		log.Errorf("l.dao.DelMapping(%d,%s) error(%v)", mid, key, server)
		return
	}
	log.Infof("conn disconnected key:%s server:%s mid:%d", key, server, mid)
	return
}

// Heartbeat heartbeat a conn. 对应 goim/internal/logic/dao/redis.go
func (l *Logic) Heartbeat(c context.Context, mid int64, key, server string) (err error) {
	has, err := l.dao.ExpireMapping(c, mid, key)
	if err != nil {
		log.Errorf("l.dao.ExpireMapping(%d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if !has {
		if err = l.dao.AddMapping(c, mid, key, server); err != nil {
			log.Errorf("l.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	log.Infof("conn heartbeat key:%s server:%s mid:%d", key, server, mid)
	return
}

// RenewOnline 将对应comet服务写入redis "HSET" "ol_192.168.3.222" "43" "{\"server\":\"192.168.3.222\",\"room_count\":{\"live://1000\":1},\"updated\":1577077540}"
func (l *Logic) RenewOnline(c context.Context, server string, roomCount map[string]int32) (map[string]int32, error) {
	online := &model.Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := l.dao.AddServerOnline(context.Background(), server, online); err != nil {
		return nil, err
	}
	return l.roomCount, nil
}

// Receive receive a message.
func (l *Logic) Receive(c context.Context, mid int64, proto *grpc.Proto) (err error) {
	log.Infof("receive mid:%d message:%+v", mid, proto)
	return
}
