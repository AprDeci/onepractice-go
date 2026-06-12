package handler

import (
	"onepractice-golang/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check 健康检查。
// @Summary 健康检查
// @Description 返回服务状态和数据库是否启用。
// @Tags health
// @Produce json
// @Success 200 {object} response.Body
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	data := gin.H{"status": "ok", "database": h.db != nil}
	response.Success(c, data)
}
