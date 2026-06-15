package router

import (
	"onepractice-golang/internal/config"
	"onepractice-golang/internal/handler"
	"onepractice-golang/internal/middleware"
	"onepractice-golang/internal/service"

	"github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"gorm.io/gorm"
)

func New(cfg config.Config, database *gorm.DB, redisClient *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), middleware.Recovery())

	health := handler.NewHealthHandler(database)
	r.GET("/health", health.Check)

	plugin := sagin.NewPlugin(sagin.GetManager())
	api := r.Group("/api")
	api.Use(plugin.TokenInterceptor())

	captchaService := service.NewCaptchaService(database, cfg.Mail, redisClient)
	userHandler := handler.NewUserHandler(service.NewUserService(database, captchaService))
	users := api.Group("/user")
	users.POST("/register", userHandler.Register)
	users.POST("/login", userHandler.Login)
	users.POST("/resetpassword", userHandler.ResetPassword)

	captchaHandler := handler.NewCaptchaHandler(captchaService)
	captcha := api.Group("/captcha")
	captcha.GET("/email", captchaHandler.Email)
	captcha.POST("/email/verify", captchaHandler.VerifyEmail)

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

	paperService := service.NewPaperService(database)
	recordHandler := handler.NewRecordHandler(service.NewRecordService(redisClient, paperService))

	protected := api.Group("")
	protected.Use(plugin.AuthMiddleware())
	protected.GET("/user/info", userHandler.Info)
	protected.POST("/user/logout", userHandler.Logout)
	protected.GET("/auth/check", health.Check)
	protected.POST("/record/save", recordHandler.Save)
	protected.GET("/record/list", recordHandler.List)
	protected.POST("/record/update", recordHandler.Update)
	// OpenAPI
	r.GET("/openapi/*any", openapiui.WrapHandler(openapiui.Config{
		SpecURL:      "/openapi/openapi.json",
		SpecFilePath: "./openapi/swagger.json",
		Title:        "Example API",
		Theme:        "light", // or "dark"
	}))

	return r
}
