package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

func registerPaperRoutes(rg *gin.RouterGroup, h *handlerv1.PaperHandler) {
	rg.GET("/papers", h.List)
	rg.GET("/papers/:paperId/intro", h.Intro)
	rg.GET("/paper-types", h.Types)
}
