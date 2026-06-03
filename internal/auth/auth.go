package auth

import (
	"onepractice-golang/internal/config"

	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"github.com/sa-tokens/sa-token-go/storage/memory"
)

func Init(cfg config.AuthConfig) {
	storage := memory.NewStorage()
	saConfig := sagin.DefaultConfig()
	saConfig.TokenName = cfg.TokenName
	saConfig.Timeout = cfg.Timeout
	saConfig.IsPrintBanner = false

	manager := sagin.NewManager(storage, saConfig)
	sagin.SetManager(manager)
}
