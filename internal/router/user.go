package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(rg *gin.RouterGroup, h *handlerv1.UserHandler) {
	rg.POST("/auth/registrations", h.Register)
	rg.POST("/auth/sessions", h.Login)
	rg.POST("/auth/password-resets", h.ResetPassword)
}

func registerUserProtectedRoutes(rg *gin.RouterGroup, h *handlerv1.UserHandler) {
	rg.GET("/users/me", h.Info)
	rg.DELETE("/auth/sessions/current", h.Logout)
}
