package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

func registerDictionaryRoutes(rg *gin.RouterGroup, h *handlerv1.DictionaryHandler) {
	rg.GET("/dictionary/definitions", h.Lookup)
	rg.GET("/dictionary/words", h.ListWords)
	rg.GET("/dictionary/words/by-spelling/:spelling", h.GetWordBySpelling)
	rg.GET("/dictionary/words/:wordId", h.GetWord)
	rg.GET("/dictionary/books", h.ListBooks)
	rg.GET("/dictionary/books/:bookId/words", h.ListBookWords)
}
