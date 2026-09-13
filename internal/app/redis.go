package app

import (
	"onepractice-golang/internal/cache"
	"onepractice-golang/internal/config"

	"github.com/redis/go-redis/v9"
)

func openRedis(cfg config.RedisConfig) (*redis.Client, error) {
	return cache.Open(cfg)
}

func closeRedis(client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}
