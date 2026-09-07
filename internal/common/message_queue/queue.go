package message_queue

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

const (
	HashSuffix = ":hash"
	SetSuffix  = ":set"
)

var once sync.Once

type Queue struct {
	ctx context.Context

	redis *redis.Client
	topic string

	producer *producer
	consumer *consumer
}

func NewQueue(ctx context.Context, redis *redis.Client, opts ...Option) *Queue {
	var queue *Queue
	var defaultOptions Options

	once.Do(func() {
		defaultOptions = Options{
			topic:   "topic",
			handler: defaultHandler,
		}
	})

	for _, apply := range opts {
		apply(&defaultOptions)
	}

	// 创建Queue实例
	queue = &Queue{
		ctx:      ctx,
		redis:    redis,
		topic:    defaultOptions.topic,
		producer: NewProducer(ctx),                         // 创建生产者
		consumer: NewConsumer(ctx, defaultOptions.handler), // 创建消费者，使用处理函数
	}

	return queue
}

func (q *Queue) Start() {
	go q.consumer.listen(q.redis, q.topic)
}

func (q *Queue) publish(msg *Message) (int64, error) {
	return q.producer.publish(q.redis, q.topic, msg)
}
