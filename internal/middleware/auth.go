package middleware

import (
	"onepractice-golang/internal/response"

	"github.com/gin-gonic/gin"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := sagin.GetTokenFromCtx(c)
		if token == "" {
			response.ErrorEnum(c, response.ErrTokenInvalid)
			c.Abort()
			return
		}
		if _, err := sagin.GetLoginID(token); err != nil {
			response.ErrorEnum(c, response.ErrTokenInvalid)
			c.Abort()
			return
		}
		c.Next()
	}
}
