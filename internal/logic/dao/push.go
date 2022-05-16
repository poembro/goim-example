package dao

import (
	"context"
	pb "goim-demo/api/logic/grpc"

	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
)

// PushMsg  针对单个人的推送
func (d *Dao) PushMsg(c context.Context, op int32, server string, keys []string, msg []byte) (err error) {
	conn := d.redis.Get()
	defer conn.Close()

	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_PUSH,
		Operation: op,
		Server:    server,
		Keys:      keys,
		Msg:       msg,
	}

	// 即时消息存储扩展 HOOKS:
	// 在这里增加即时消息存储扩展
	// 如果需要只存储离线消息, 可以先检查当前用户是否在线, 依据用户在线情况处理存储

	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}

	_, err = conn.Do("PUBLISH", d.c.Kafka.Topic, string(b))
	if err != nil {
		log.Errorf("PushMsg.send(broadcast_room pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}

// BroadcastRoomMsg 针对房间的推送
func (d *Dao) BroadcastRoomMsg(c context.Context, op int32, room string, msg []byte) (err error) {
	conn := d.redis.Get()
	defer conn.Close()

	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_ROOM,
		Operation: op,
		Room:      room,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}

	_, err = conn.Do("PUBLISH", d.c.Kafka.Topic, string(b))
	if err != nil {
		log.Errorf("PushMsg.send(broadcast_room pushMsg:%v) error(%v)", pushMsg, err)
	}

	return
}

// BroadcastMsg 针对所有房间的推送
func (d *Dao) BroadcastMsg(c context.Context, op, speed int32, msg []byte) (err error) {
	conn := d.redis.Get()
	defer conn.Close()

	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_BROADCAST,
		Operation: op,
		Speed:     speed,
		Msg:       msg,
	}

	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}

	_, err = conn.Do("PUBLISH", d.c.Kafka.Topic, string(b))
	if err != nil {
		log.Errorf("PushMsg.send(broadcast_room pushMsg:%v) error(%v)", pushMsg, err)
	}

	return
}
