package message_queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const (
	HashSuffix = ":hash"
	SetSuffix  = ":set"
)

type Queue struct {
	ctx context.Context

	redis *redis.Client
	topic string

	producer *producer
	consumer *consumer
}

func NewQueue(ctx context.Context, redis *redis.Client, opts ...Option) *Queue {
	defaultOptions := Options{topic: "topic", handler: defaultHandler, workers: 4, batchSize: 8}

	for _, apply := range opts {
		apply(&defaultOptions)
	}

	producer := NewProducer(ctx)
	// 创建Queue实例
	queue := &Queue{
		ctx:      ctx,
		redis:    redis,
		topic:    defaultOptions.topic,
		producer: producer,
		consumer: NewConsumer(ctx, defaultOptions.handler, producer, defaultOptions.workers, defaultOptions.batchSize),
	}

	return queue
}

func (q *Queue) Start() {
	if q == nil || q.redis == nil {
		return
	}
	go q.consumer.listen(q.redis, q.topic)
}

func (q *Queue) Publish(msg *Message) (int64, error) {
	if q == nil || q.redis == nil {
		return 0, nil
	}
	return q.producer.publish(q.redis, q.topic, msg)
}
