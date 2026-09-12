package router

import (
	"context"
	"fmt"
	"onepractice-golang/internal/agent"
	"onepractice-golang/internal/agent/llm"
	"onepractice-golang/internal/config"
	"onepractice-golang/internal/handler"
	handlerv1 "onepractice-golang/internal/handler/v1"
	"onepractice-golang/internal/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Deps struct {
	LegacyUser         *handler.UserHandler
	LegacyCaptcha      *handler.CaptchaHandler
	LegacyPaper        *handler.PaperHandler
	LegacyQuestion     *handler.QuestionHandler
	LegacyDictionary   *handler.DictionaryHandler
	LegacyRecord       *handler.RecordHandler
	LegacyWordFavorite *handler.WordFavoriteHandler
	LegacyHealth       *handler.HealthHandler
	V1User             *handlerv1.UserHandler
	V1Captcha          *handlerv1.CaptchaHandler
	V1Paper            *handlerv1.PaperHandler
	V1Question         *handlerv1.QuestionHandler
	V1Dictionary       *handlerv1.DictionaryHandler
	V1Record           *handlerv1.RecordHandler
	V1WordFavorite     *handlerv1.WordFavoriteHandler
	V1Essay            *handlerv1.EssayHandler
	V1Agent            *handlerv1.AgentHandler
	V1Points           *handlerv1.PointsHandler

	// Points 供进程级后台任务（如预扣补偿 cron）复用。
	Points *service.PointsService
}

func newDeps(cfg config.Config, db *gorm.DB, redisClient *redis.Client, mailSender service.MailSender, essayService *service.EssayService) Deps {
	captchaSvc := service.NewCaptchaServiceWithSender(db, redisClient, mailSender)
	paperSvc := service.NewPaperService(db)
	userSvc := service.NewUserService(db, captchaSvc)
	questionSvc := service.NewQuestionService(db)
	dictionarySvc := service.NewDictionaryService(db)
	recordSvc := service.NewRecordService(paperSvc, db)
	favoriteSvc := service.NewWordFavoriteService(db)
	pointsSvc := service.NewPointsService(db, redisClient, cfg.Points)
	if essayService != nil {
		essayService.SetPoints(pointsSvc)
	}

	return Deps{
		LegacyUser:         handler.NewUserHandler(userSvc),
		LegacyCaptcha:      handler.NewCaptchaHandler(captchaSvc),
		LegacyPaper:        handler.NewPaperHandler(paperSvc),
		LegacyQuestion:     handler.NewQuestionHandler(questionSvc),
		LegacyDictionary:   handler.NewDictionaryHandler(dictionarySvc),
		LegacyRecord:       handler.NewRecordHandler(recordSvc),
		LegacyWordFavorite: handler.NewWordFavoriteHandler(favoriteSvc),
		LegacyHealth:       handler.NewHealthHandler(db),
		V1User:             handlerv1.NewUserHandler(userSvc),
		V1Captcha:          handlerv1.NewCaptchaHandler(captchaSvc),
		V1Paper:            handlerv1.NewPaperHandler(paperSvc),
		V1Question:         handlerv1.NewQuestionHandler(questionSvc),
		V1Dictionary:       handlerv1.NewDictionaryHandler(dictionarySvc),
		V1Record:           handlerv1.NewRecordHandler(recordSvc),
		V1WordFavorite:     handlerv1.NewWordFavoriteHandler(favoriteSvc),
		V1Essay:            handlerv1.NewEssayHandler(essayService),
		V1Agent:            handlerv1.NewAgentHandler(llm.NewGlmClient(cfg.LLM.GlmKey), pointsSvc),
		V1Points:           handlerv1.NewPointsHandler(pointsSvc),
		Points:             pointsSvc,
	}
}

// newEssayService 依据 cfg.LLM.Default 选定模型并创建作文批改服务。
func newEssayService(cfg config.Config, redisClient *redis.Client, db *gorm.DB) (*service.EssayService, error) {
	modelCfg, ok := cfg.LLM.Models[cfg.LLM.Default]
	if !ok {
		return nil, fmt.Errorf("llm.default %q 未在 llm.models 中定义", cfg.LLM.Default)
	}

	cm, err := agent.NewEssayChatModel(context.Background(), llm.ModelConfig{
		BaseURL:     modelCfg.BaseURL,
		Model:       modelCfg.Model,
		APIKey:      modelCfg.APIKey,
		Temperature: modelCfg.Temperature,
	})
	if err != nil {
		return nil, err
	}
	return service.NewEssayService(redisClient, cm, db), nil
}
