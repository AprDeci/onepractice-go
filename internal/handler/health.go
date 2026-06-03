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

func (h *HealthHandler) Check(c *gin.Context) {
	data := gin.H{"status": "ok", "database": h.db != nil}
	response.Success(c, data)
}
