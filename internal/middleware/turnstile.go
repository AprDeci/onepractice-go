package middleware

import (
	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	"onepractice-golang/internal/common/turnstile"

	"github.com/gin-gonic/gin"
)

// TurnstileTokenHeader 是携带一次性人机校验 token 的请求头。
const TurnstileTokenHeader = "X-Turnstile-Token"

// VerifyTurnstile 校验请求头中的 Turnstile token；校验器为 nil 或未启用时放行。
func VerifyTurnstile(v *turnstile.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := v.Verify(c.Request.Context(), c.GetHeader(TurnstileTokenHeader), c.ClientIP()); err != nil {
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "人机校验失败"))
			c.Abort()
			return
		}
		c.Next()
	}
}
