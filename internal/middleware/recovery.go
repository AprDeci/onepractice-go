package middleware

import (
	"log/slog"

	"onepractice-golang/internal/response"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		slog.Error("panic recovered", "error", recovered)
		response.Error(c, 500, "server error")
	})
}
