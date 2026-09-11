package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

func registerEssayRoutes(rg *gin.RouterGroup, h *handlerv1.EssayHandler) {
	rg.POST("/essay/tasks", h.CreateTask)
	rg.GET("/essay/tasks/:taskId", h.GetTask)
	rg.GET("/essay/records/:recordId/results", h.GetResultsByRecord)
}
