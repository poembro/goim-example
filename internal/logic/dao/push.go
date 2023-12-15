package dao

import (
	"context"
	pb "goim-example/api/logic"
	"strconv"

	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"

	"github.com/Shopify/sarama"
)

func (d *Dao) Send(pushMsg *pb.PushMsg) (err error) {
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}

	if d.c.Consume.KafkaEnable {
		var key string
		if len(pushMsg.Keys) > 0 {
			key = pushMsg.Keys[0]
		} else if pushMsg.Room != "" {
			key = pushMsg.Room
		} else if pushMsg.Operation != 0 {
			key = strconv.FormatInt(int64(pushMsg.Operation), 10)
		}
		m := &sarama.ProducerMessage{
			Key:   sarama.StringEncoder(key),
			Topic: d.c.Kafka.Topic,
			Value: sarama.ByteEncoder(b),
		}
		if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
			log.Errorf("PushMsg.send(push pushMsg:%v) error(%v)", pushMsg, err)
		}
	}

	if d.c.Consume.RedisEnable {
		conn := d.redis.Get()
		defer conn.Close()
		_, err = conn.Do("PUBLISH", d.c.Kafka.Topic, string(b))
		if err != nil {
			log.Errorf("PushMsg.send(broadcast_room pushMsg:%v) error(%v)", pushMsg, err)
		}
	}
	return nil
}

// PushMsg  针对单个人的推送
func (d *Dao) PushMsg(c context.Context, op int32, server string, keys []string, msg []byte) error {
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

	return d.Send(pushMsg)
}

// BroadcastRoomMsg 针对房间的推送
func (d *Dao) BroadcastRoomMsg(c context.Context, op int32, room string, msg []byte) error {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_ROOM,
		Operation: op,
		Room:      room,
		Msg:       msg,
	}
	return d.Send(pushMsg)
}

// BroadcastMsg 针对所有房间的推送
func (d *Dao) BroadcastMsg(c context.Context, op, speed int32, msg []byte) error {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_BROADCAST,
		Operation: op,
		Speed:     speed,
		Msg:       msg,
	}

	return d.Send(pushMsg)
}
