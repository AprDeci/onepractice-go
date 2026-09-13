package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

func registerAgentRoutes(rg *gin.RouterGroup, h *handlerv1.AgentHandler) {
	rg.POST("/ocr", h.OCR)
}
