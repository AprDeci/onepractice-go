package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

func registerWordFavoriteRoutes(rg *gin.RouterGroup, h *handlerv1.WordFavoriteHandler) {
	rg.GET("/users/me/favorite-words", h.List)
	rg.POST("/users/me/favorite-words", h.Add)
	rg.GET("/users/me/favorite-words/:wordId", h.Check)
	rg.DELETE("/users/me/favorite-words/:wordId", h.Remove)
}
