package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

func registerRecordRoutes(rg *gin.RouterGroup, h *handlerv1.RecordHandler) {
	rg.POST("/records", h.Create)
	rg.GET("/records", h.List)
	rg.PUT("/records/:recordId", h.Update)
}
