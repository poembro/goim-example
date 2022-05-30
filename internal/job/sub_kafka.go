package job

import (
	"context"
	"fmt"
	pb "goim-demo/api/logic"
	"goim-demo/internal/job/conf"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"

	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
)

type AppConsumer struct {
	j     *Job
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *AppConsumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *AppConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *AppConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	//不要将下面的代码移动到goroutine
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for msg := range claim.Messages() {
		// MarkMessage将消息标记为已使用
		session.MarkMessage(msg, "")

		// 推送消息过程
		pushMsg := new(pb.PushMsg)
		if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
			log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
			continue
		}
		//获取kafka消息后 protobuff格式解析
		if err := consumer.j.Push(context.Background(), pushMsg); err != nil {
			log.Errorf("j.Push(%v) error(%v)", pushMsg, err)
		}

		log.Infof("consume: %s/Partition: %d/Offset: %d\t%s\t%+v", msg.Topic, msg.Partition, msg.Offset, msg.Key, pushMsg)
	}

	return nil
}

func newKafkaSub(c *conf.Kafka) sarama.ConsumerGroup {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_1_0
	config.ClientID = fmt.Sprintf("%s_%s", c.Group, uuid.New().String())
	config.ChannelBufferSize = 256 // channel长度默认256

	//一开始是哪个worker在处理就一直是它，后面加进来的worker不起作用 除非第一个挂了
	//config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	//config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin //轮流
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange //默认  分区分配策略

	//多个partition 自动计算, 设置选择分区的策略为Hash
	// 生产消息时记得 &sarama.ProducerMessage{ Key: sarama.StringEncoder(strconv.Itoa(RecvID)),)
	// Kafka客户端会根据Key进行Hash，我们通过把接收用户ID作为Key，这样就能让所有发给某个人的消息落到同一个分区了，也就有序了。
	//p.config.Producer.Partitioner = sarama.NewHashPartitioner  // 默认 hash

	config.Consumer.Return.Errors = true

	//config.Consumer.Offsets.Initial = sarama.OffsetOldest //从最老的位置开始消费   默认是从最新的位置开始消费
	//config.Consumer.Offsets.Initial = -2                    // 未找到组消费位移的时候从哪边开始消费

	client, err := sarama.NewConsumerGroup(c.Brokers, c.Group, config)
	if err != nil {
		panic(err)
	}
	return client
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Infof("Resuming consumption")
	} else {
		client.PauseAll()
		log.Infof("Pausing consumption")
	}

	*isPaused = !*isPaused
}

// Consume messages, watch signals
func (j *Job) ConsumeKafka() {
	keepRunning := true
	consumptionIsPaused := false

	app := &AppConsumer{
		j:     j,
		ready: make(chan bool),
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := j.consumer.Consume(ctx, strings.Split(j.c.Kafka.Topic, ","), app); err != nil {
				log.Infof("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			app.ready = make(chan bool)
		}
	}()

	<-app.ready // Await till the consumer has been set up

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	log.Infof("Sarama consumer up and running!...")
	for keepRunning {
		select {
		case <-ctx.Done():
			log.Infof("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Infof("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(j.consumer, &consumptionIsPaused)
		}
	}
	cancel()
	wg.Wait()
	if err := j.consumer.Close(); err != nil {
		log.Infof("Error closing client: %v", err)
	}
}
