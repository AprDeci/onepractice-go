package handler

import (
	"onepractice-golang/internal/common/response"
	"onepractice-golang/internal/dto"

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
// @Success 200 {object} response.Body{data=dto.HealthResponse}
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	data := dto.HealthResponse{Status: "ok", Database: h.db != nil}
	response.Success(c, data)
}
