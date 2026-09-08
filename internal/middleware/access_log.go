package middleware

import (
	"log/slog"
	"onepractice-golang/internal/common/response"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(response.RequestIDKey)
		attrs := []slog.Attr{
			slog.String("request_id", stringValue(requestID)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int("size", c.Writer.Size()),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", time.Since(start)),
		}
		if c.Writer.Status() >= 500 {
			if contextErr := c.Errors.Last(); contextErr != nil {
				attrs = append(attrs, slog.Any("error", contextErr.Err))
			}
			logger.LogAttrs(c.Request.Context(), slog.LevelError, "HTTP 请求处理失败", attrs...)
			return
		}
		logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "HTTP 请求完成", attrs...)
	}
}

func stringValue(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}
