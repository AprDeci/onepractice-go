package router

import (
	"context"
	"fmt"
	"log/slog"
	"onepractice-golang/internal/common/mail"
	"onepractice-golang/internal/common/message_queue"
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

// New 组装依赖并返回 HTTP 引擎，以及用于关闭后台模块（mail、essay 等）的 cleanup 函数。
func New(cfg config.Config, database *gorm.DB, redisClient *redis.Client, logger *slog.Logger) (*gin.Engine, func(), error) {
	queueCtx, cancelQueues := context.WithCancel(context.Background())

	mailModule := mail.NewModule(cfg.Mail)
	if redisClient != nil {
		mailQueue := message_queue.NewQueue(queueCtx, redisClient,
			message_queue.WithTopic(mail.Topic),
			message_queue.WithHandler(mail.Consume(mailModule.Sender)),
		)
		mailQueue.Start()
		mailModule.Sender = mail.NewQueueSender(mailQueue)
	}

	essayService, err := newEssayService(cfg, redisClient, database)
	if err != nil {
		cancelQueues()
		return nil, nil, fmt.Errorf("初始化作文批改服务失败: %w", err)
	}
	if redisClient != nil {
		essayQueue := message_queue.NewQueue(queueCtx, redisClient,
			message_queue.WithTopic(service.EssayTopic),
			message_queue.WithWorkers(5),
			message_queue.WithHandler(essayService.Handle),
		)
		essayQueue.Start()
		essayService.SetQueue(essayQueue)
	}

	cleanup := func() {
		cancelQueues()
	}

	r := gin.New()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.RequestID(), middleware.AccessLog(logger), middleware.Recovery(logger))

	deps := newDeps(cfg, database, redisClient, mailModule.Sender, essayService)
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
	registerEssayRoutes(v1Protected, deps.V1Essay)

	legacy := r.Group("/api")
	legacy.Use(middleware.TimeoutMiddleware(5 * time.Second))
	legacy.Use(plugin.TokenInterceptor())
	legacy.Use(middleware.Deprecated("Wed, 30 Sep 2026 00:00:00 GMT"))
	registerLegacyPublicRoutes(legacy, deps)
	legacyProtected := legacy.Group("")
	legacyProtected.Use(middleware.Auth())
	registerLegacyProtectedRoutes(legacyProtected, deps)
	registerAgentRoutes(legacyProtected, deps.LegacyAgent)

	return r, cleanup, nil
}
