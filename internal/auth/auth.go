package auth

import (
	"onepractice-golang/internal/config"

	"github.com/redis/go-redis/v9"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	redisstore "github.com/sa-tokens/sa-token-go/storage/redis"
)

// Satoken-go 初始化
func Init(cfg config.AuthConfig, redisClient *redis.Client) {
	storage := memory.NewStorage()
	if redisClient != nil {
		storage = redisstore.NewStorageFromClient(redisClient)
	}
	saConfig := sagin.DefaultConfig()
	saConfig.TokenName = cfg.TokenName
	saConfig.Timeout = cfg.Timeout
	saConfig.IsPrintBanner = false

	manager := sagin.NewManager(storage, saConfig)
	sagin.SetManager(manager)
}
