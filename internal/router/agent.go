package router

import (
	"onepractice-golang/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerAgentRoutes(rg *gin.RouterGroup, h *handler.AgentHandler) {
	rg.POST("/ocr", h.OCR)
}
