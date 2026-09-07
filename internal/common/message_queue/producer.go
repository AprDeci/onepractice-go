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
	z := redis.Z{
		Score:  msg.GetScore(),
		Member: msg.GetId(),
	}

	setkey := topic + SetSuffix
	n, err := redisClient.ZAdd(p.ctx, setkey, z).Result()
	if err != nil {
		return n, err
	}

	hashkey := topic + HashSuffix
	return redisClient.HSet(p.ctx, hashkey, msg.GetId(), msg).Result()

}
