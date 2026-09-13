package middleware

import (
	"log/slog"
	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Get(response.RequestIDKey)
				attrs := []slog.Attr{
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
					slog.String("request_id", stringValue(requestID)),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("ip", c.ClientIP()),
				}
				logger.LogAttrs(c.Request.Context(), slog.LevelError, "HTTP 请求发生 panic", attrs...)
				response.Error(c, apperror.New(apperror.CodeInternal, "系统异常"))
				c.Abort()
			}
		}()
		c.Next()
	}
}
