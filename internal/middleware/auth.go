package middleware

import (
	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"

	"github.com/gin-gonic/gin"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := sagin.GetTokenFromCtx(c)
		if token == "" {
			// response.ErrorEnum(c, response.ErrTokenInvalid)
			response.Error(c, apperror.New(apperror.CodeUnauthorized, "token无效"))
			c.Abort()
			return
		}
		if _, err := sagin.GetLoginID(token); err != nil {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, "token无效"))
			c.Abort()
			return
		}
		c.Next()
	}
}
