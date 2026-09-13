package router

import (
	"onepractice-golang/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerHealthRoutes(r *gin.Engine, h *handler.HealthHandler) {
	r.GET("/health", h.Check)
}
