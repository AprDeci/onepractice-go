package router

import (
	"onepractice-golang/internal/config"
	"onepractice-golang/internal/handler"
	"onepractice-golang/internal/middleware"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"gorm.io/gorm"
)

func New(_ config.Config, database *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), middleware.Recovery())

	health := handler.NewHealthHandler(database)
	r.GET("/health", health.Check)

	plugin := sagin.NewPlugin(sagin.GetManager())
	api := r.Group("/api")
	api.Use(plugin.TokenInterceptor())

	paperHandler := handler.NewPaperHandler(service.NewPaperService(database))
	papers := api.Group("/paper")
	papers.GET("/all", paperHandler.All)
	papers.POST("/getPaperwithQuerys", paperHandler.Page)
	papers.POST("/getPaperandRatingWithQuerys", paperHandler.PageWithRating)
	papers.GET("/type", paperHandler.ByType)
	papers.GET("/types", paperHandler.Types)
	papers.GET("/intro", paperHandler.Intro)

	questionHandler := handler.NewQuestionHandler(service.NewQuestionService(database))
	questions := api.Group("/question")
	questions.GET("/getById", questionHandler.ByPaperID)
	questions.GET("/getByType", questionHandler.ByPaperIDAndType)
	questions.GET("/getAllByIdSplitByPart", questionHandler.SplitByPart)
	questions.GET("/getAnswersByPaperId", questionHandler.Answers)

	protected := api.Group("")
	protected.Use(plugin.AuthMiddleware())
	protected.GET("/auth/check", health.Check)

	return r
}
