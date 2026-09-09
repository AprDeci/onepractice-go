package router

import (
	handlerv1 "onepractice-golang/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

func registerQuestionRoutes(rg *gin.RouterGroup, h *handlerv1.QuestionHandler) {
	rg.GET("/papers/:paperId/questions", h.List)
	rg.GET("/papers/:paperId/answers", h.Answers)
	rg.POST("/questions/practice", h.Practice)
}
