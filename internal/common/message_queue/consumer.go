package message_queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type handlerFunc func(msg Message)

func defaultHandler(msg Message) {
	fmt.Println(msg)
}

type consumer struct {
	ctx      context.Context
	duration time.Duration
	ch       chan []string
	handler  handlerFunc
}

func NewConsumer(ctx context.Context, handler handlerFunc) *consumer {
	return &consumer{
		ctx:      ctx,
		duration: time.Second,
		ch:       make(chan []string, 1000),
		handler:  handler,
	}
}

func (c *consumer) listen(redisClient *redis.Client, topic string) {

	// 处理消息
	go func() {
		for {
			select {
			case ret := <-c.ch:
				key := topic + HashSuffix
				result, err := redisClient.HMGet(c.ctx, key, ret...).Result()
				if err != nil {
					fmt.Println(err)
				}
				if len(result) > 0 {
					redisClient.HDel(c.ctx, key, ret...)
				}
				msg := Message{}
				for _, v := range result {
					if v == nil {
						continue
					}
					str := v.(string)
					json.Unmarshal([]byte(str), &msg)

					go c.handler(msg)
				}
			}
		}
	}()

	// 监听消息
	ticker := time.NewTicker(c.duration)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			log.Println("consumer quit:", c.ctx.Err())
			return
		case <-ticker.C:
			start := int64(0)
			end := int64(time.Now().Unix())

			key := topic + SetSuffix
			result, err := redisClient.ZRange(c.ctx, key, start, end).Result()
			if err != nil {
				fmt.Println(err)
				return
			}

			if len(result) > 0 {
				redisClient.ZRemRangeByScore(c.ctx, key, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end)).Result()

				c.ch <- result
			}
		}

	}
}
