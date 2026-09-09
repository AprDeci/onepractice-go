package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

func registerCaptchaRoutes(rg *gin.RouterGroup, h *handlerv1.CaptchaHandler) {
	rg.POST("/auth/email-verifications", h.SendEmail)
	rg.POST("/auth/email-verifications/verification", h.VerifyEmail)
}
