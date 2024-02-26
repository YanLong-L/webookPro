package startup

import (
	"github.com/IBM/sarama"
	events2 "webookpro/interactive/events"
	"webookpro/internal/events"
)

func InitKafka() sarama.Client {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9092"}, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
func NewConsumers(c1 *events2.InteractiveReadEventBatchConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
