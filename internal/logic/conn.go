package logic

import (
	"context"
	"encoding/json"
	"goim-example/api/protocol"
	"goim-example/internal/logic/model"
	"time"

	log "github.com/golang/glog"
	"github.com/google/uuid"
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
	log.Infof("conn connected params:%+v", params)

	mid = params.Mid
	roomID = params.RoomID
	accepts = params.Accepts
	hb = int64(l.c.Node.Heartbeat) * int64(l.c.Node.HeartbeatMax)
	if key = params.Key; key == "" {
		key = uuid.New().String()
	}
	if err = l.dao.AddMapping(c, mid, key, server); err != nil {
		log.Errorf("l.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
	}
	log.Infof("conn connected err:%+v", err)

	log.Infof("conn connected key:%s server:%s mid:%d token:%s", key, server, mid, token)
	return
}

// Receive receive a message. 框架之外,第三方业务 逻辑扩展
func (l *Logic) Receive(c context.Context, mid int64, proto *protocol.Proto) (err error) {
	switch proto.Op {
	case protocol.OpSync:
		// op, keys, msg, err := l.business.MsgSync(c, mid, proto.Body)
		// if op != 0 && err == nil {
		// 	l.PushKeys(c, op, keys, msg)
		// }
	case protocol.OpMessageAck:
		// l.business.MessageACK(c, mid, proto.Body)
	default:
		log.Infof("receive mid:%d message:%+v", mid, proto)
	}
	return
}

// Keys is online   框架之外,第三方业务 逻辑扩展
func (l *Logic) IsOnline(c context.Context, keys []string) bool {
	servers, err := l.dao.ServersByKeys(c, keys)
	if err != nil {
		return false
	}
	if len(servers) > 0 && servers[0] == "" {
		return false
	}

	return true
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

// Heartbeat heartbeat a conn.
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
