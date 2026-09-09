package router

import (
	"log/slog"
	"onepractice-golang/internal/config"
	"onepractice-golang/internal/middleware"
	"onepractice-golang/internal/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"gorm.io/gorm"
)

func New(cfg config.Config, database *gorm.DB, redisClient *redis.Client, mailSender service.MailSender, logger *slog.Logger) *gin.Engine {
	r := gin.New()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.RequestID(), middleware.AccessLog(logger), middleware.Recovery(logger))

	deps := newDeps(cfg, database, redisClient, mailSender)
	registerHealthRoutes(r, deps.LegacyHealth)
	registerDocsRoutes(r)
	plugin := sagin.NewPlugin(sagin.GetManager())

	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.TimeoutMiddleware(5 * time.Second))
	apiV1.Use(plugin.TokenInterceptor())
	registerUserRoutes(apiV1, deps.V1User)
	registerCaptchaRoutes(apiV1, deps.V1Captcha)
	registerPaperRoutes(apiV1, deps.V1Paper)
	registerQuestionRoutes(apiV1, deps.V1Question)
	registerDictionaryRoutes(apiV1, deps.V1Dictionary)
	v1Protected := apiV1.Group("")
	v1Protected.Use(middleware.Auth())
	registerUserProtectedRoutes(v1Protected, deps.V1User)
	registerRecordRoutes(v1Protected, deps.V1Record)
	registerWordFavoriteRoutes(v1Protected, deps.V1WordFavorite)

	legacy := r.Group("/api")
	legacy.Use(middleware.TimeoutMiddleware(5 * time.Second))
	legacy.Use(plugin.TokenInterceptor())
	legacy.Use(middleware.Deprecated("Wed, 30 Sep 2026 00:00:00 GMT"))
	registerLegacyPublicRoutes(legacy, deps)
	legacyProtected := legacy.Group("")
	legacyProtected.Use(middleware.Auth())
	registerLegacyProtectedRoutes(legacyProtected, deps)
	registerAgentRoutes(legacyProtected, deps.LegacyAgent)

	return r
}
