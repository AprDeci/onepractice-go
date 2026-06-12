package cache

import (
	"context"
	"time"

	"onepractice-golang/internal/config"

	"github.com/redis/go-redis/v9"
)

func Open(cfg config.RedisConfig) (*redis.Client, error) {
	if cfg.Disabled {
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
