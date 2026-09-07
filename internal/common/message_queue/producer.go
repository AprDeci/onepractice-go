package message_queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type producer struct {
	ctx context.Context
}

func NewProducer(ctx context.Context) *producer {
	return &producer{
		ctx: ctx,
	}
}

func (p *producer) publish(redisClient *redis.Client, topic string, msg *Message) (int64, error) {
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 5
	}
	payload, err := msg.MarshalBinary()
	if err != nil {
		return 0, err
	}
	return redisClient.Eval(p.ctx, `redis.call('HSET', KEYS[2], ARGV[1], ARGV[3]); return redis.call('ZADD', KEYS[1], ARGV[2], ARGV[1])`, []string{topic + SetSuffix, topic + HashSuffix}, msg.GetId(), msg.GetScore(), payload).Int64()

}
