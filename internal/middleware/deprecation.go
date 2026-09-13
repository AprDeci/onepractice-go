package middleware

import "github.com/gin-gonic/gin"

func Deprecated(sunset string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Deprecation", "true")
		c.Header("Sunset", sunset)
		c.Next()
	}
}
