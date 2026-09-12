package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

// registerPointsPublicRoutes 注册无需鉴权的积分接口（价格/消耗信息）。
func registerPointsPublicRoutes(rg *gin.RouterGroup, h *handlerv1.PointsHandler) {
	rg.GET("/points/costs", h.Costs)
}

// registerPointsRoutes 注册需要鉴权的积分接口。
func registerPointsRoutes(rg *gin.RouterGroup, h *handlerv1.PointsHandler) {
	rg.GET("/points/balance", h.Balance)
	rg.GET("/points/transactions", h.ListTransactions)
	rg.POST("/points/daily-checkin", h.DailyCheckin)
	rg.GET("/points/checkin-status", h.CheckinStatus)
}
