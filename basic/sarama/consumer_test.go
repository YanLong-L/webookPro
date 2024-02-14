package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	// 正常来说，一个消费者都是归属于一个消费者的组的
	// 消费者组就是你的业务
	consumer, err := sarama.NewConsumerGroup(addrs,
		"test_group", cfg)
	require.NoError(t, err)
	// 带超时的 context
	start := time.Now()
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	time.AfterFunc(time.Minute*10, func() {
		cancel()
	})
	err = consumer.Consume(ctx,
		[]string{"test_topic"}, testConsumerGroupHandler{})
	// 你消费结束，就会到这里
	t.Log(err, time.Since(start).String())
}

type testConsumerGroupHandler struct {
}

func (t testConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	println("setup")
	return nil
}

func (t testConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	println("cleanup")
	return nil
}

func (t testConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		m1 := msg
		go func() {
			// 消费msg
			log.Println(string(m1.Value))
			session.MarkMessage(m1, "")
		}()
	}
	return nil
}

// ConsumeClaimV1 批量消费版本
func (t testConsumerGroupHandler) ConsumeClaimV1(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	const batchSize = 10 // 每次批量消费10个
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var eg errgroup.Group
		var last *sarama.ConsumerMessage
		done := false
		//
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true // 代表context已经超时了
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					// 代表消费者被关闭了
					return nil
				}
				last = msg
				eg.Go(func() error {
					// 我就在这里消费
					time.Sleep(time.Second)
					// 你在这里重试
					log.Println(string(msg.Value))
					return nil
				})
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			// 这边能怎么办？
			// 记录日志
			continue
		}
		// 就这样
		if last != nil {
			session.MarkMessage(last, "")
		}
	}

}
