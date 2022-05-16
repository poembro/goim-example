package job

import (
	"context"

	pb "goim-demo/api/logic/grpc"
	"goim-demo/internal/job/conf"

	cluster "github.com/bsm/sarama-cluster"
	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
)

func newKafkaSub(c *conf.Kafka) *cluster.Consumer {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	consumer, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Topic}, config)
	if err != nil {
		panic(err)
	}
	return consumer
}

// Consume messages, watch signals
func (j *Job) ConsumeKafka() {
	for {
		select {
		case err := <-j.consumer.Errors():
			log.Errorf("consumer error(%v)", err)
		case n := <-j.consumer.Notifications():
			log.Infof("consumer rebalanced(%v)", n)
		case msg, ok := <-j.consumer.Messages():
			if !ok {
				return
			}
			j.consumer.MarkOffset(msg, "")
			// 推送消息过程
			pushMsg := new(pb.PushMsg)
			if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
				log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
				continue
			}
			//获取kafka消息后 protobuff格式解析
			if err := j.push(context.Background(), pushMsg); err != nil {
				log.Errorf("j.push(%v) error(%v)", pushMsg, err)
			}
			log.Infof("consume: %s/Partition: %d/Offset: %d\t%s\t%+v", msg.Topic, msg.Partition, msg.Offset, msg.Key, pushMsg)
		}
	}
}
