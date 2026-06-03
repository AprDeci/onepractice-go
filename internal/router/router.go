package router

import (
	"onepractice-golang/internal/config"
	"onepractice-golang/internal/handler"
	"onepractice-golang/internal/middleware"

	"github.com/gin-gonic/gin"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"gorm.io/gorm"
)

func New(_ config.Config, database *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), middleware.Recovery())

	health := handler.NewHealthHandler(database)
	r.GET("/health", health.Check)

	plugin := sagin.NewPlugin(sagin.GetManager())
	api := r.Group("/api")
	api.Use(plugin.TokenInterceptor())

	protected := api.Group("")
	protected.Use(plugin.AuthMiddleware())
	protected.GET("/auth/check", health.Check)

	return r
}
