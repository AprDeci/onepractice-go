package middleware

import (
	"onepractice-golang/internal/common/response"
	"onepractice-golang/internal/common/util"

	"github.com/gin-gonic/gin"
)

// HeaderRequestID 是请求链路标识使用的 HTTP 头名称。
const HeaderRequestID = "X-Request-ID"

// RequestID 读取或生成请求 ID，并写入 Gin 上下文和响应头。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = util.NewRequestID()
		}
		c.Set(response.RequestIDKey, requestID)
		c.Writer.Header().Set(HeaderRequestID, requestID)
		c.Next()
	}
}
