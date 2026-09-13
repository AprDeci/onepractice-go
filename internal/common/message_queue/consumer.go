package message_queue

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Handler func(context.Context, Message) error

func defaultHandler(_ context.Context, msg Message) error {
	slog.Info("message handled", "id", msg.Id)
	return nil
}

const consumeScript = `local ids=redis.call('ZRANGEBYSCORE',KEYS[1],'-inf',ARGV[1],'LIMIT',0,ARGV[2]);local ret={};for _,id in ipairs(ids) do local p=redis.call('HGET',KEYS[2],id);redis.call('ZREM',KEYS[1],id);if p then redis.call('HDEL',KEYS[2],id);table.insert(ret,p) end end;return ret`

type consumer struct {
	ctx                context.Context
	handler            Handler
	producer           *producer
	workers, batchSize int
}

func NewConsumer(ctx context.Context, handler Handler, producer *producer, workers, batchSize int) *consumer {
	return &consumer{ctx: ctx, handler: handler, producer: producer, workers: workers, batchSize: batchSize}
}
func (c *consumer) listen(r *redis.Client, topic string) {
	jobs := make(chan Message, c.workers)
	for i := 0; i < c.workers; i++ {
		go c.worker(r, topic, jobs)
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			close(jobs)
			return
		case <-ticker.C:
			payloads, err := r.Eval(c.ctx, consumeScript, []string{topic + SetSuffix, topic + HashSuffix}, time.Now().Unix(), c.batchSize).StringSlice()
			if err != nil {
				slog.Error("consume messages", "error", err, "topic", topic)
				continue
			}
			for _, p := range payloads {
				var m Message
				if err := json.Unmarshal([]byte(p), &m); err != nil {
					slog.Error("decode message", "error", err)
					continue
				}
				select {
				case jobs <- m:
				case <-c.ctx.Done():
					close(jobs)
					return
				}
			}
		}
	}
}
func (c *consumer) worker(r *redis.Client, topic string, jobs <-chan Message) {
	for m := range jobs {
		if err := c.handler(c.ctx, m); err != nil {
			c.retry(r, topic, &m, err)
		}
	}
}
func (c *consumer) retry(r *redis.Client, topic string, m *Message, err error) {
	m.RetryCount++
	m.LastError = err.Error()
	if m.RetryCount > m.MaxRetries {
		p, e := m.MarshalBinary()
		if e == nil {
			_ = r.HSet(c.ctx, topic+":dead", m.Id, p).Err()
		}
		return
	}
	delay := retryDelay(m.RetryCount)
	m.ConsumeTime = time.Now().Add(delay)
	if _, e := c.producer.publish(r, topic, m); e != nil {
		slog.Error("retry message", "error", e, "id", m.Id)
	}
}

func retryDelay(retryCount int) time.Duration {
	if retryCount < 1 {
		return 0
	}
	delay := 2 * time.Second << uint(retryCount-1)
	if delay > 30*time.Second {
		return 30 * time.Second
	}
	return delay
}
