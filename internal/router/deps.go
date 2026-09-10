package router

import (
	"context"
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
	LegacyAgent        *handler.AgentHandler
	LegacyHealth       *handler.HealthHandler
	V1User             *handlerv1.UserHandler
	V1Captcha          *handlerv1.CaptchaHandler
	V1Paper            *handlerv1.PaperHandler
	V1Question         *handlerv1.QuestionHandler
	V1Dictionary       *handlerv1.DictionaryHandler
	V1Record           *handlerv1.RecordHandler
	V1WordFavorite     *handlerv1.WordFavoriteHandler
	V1Essay            *handlerv1.EssayHandler
}

func newDeps(cfg config.Config, db *gorm.DB, redisClient *redis.Client, mailSender service.MailSender, essayService *service.EssayService) Deps {
	captchaSvc := service.NewCaptchaServiceWithSender(db, redisClient, mailSender)
	paperSvc := service.NewPaperService(db)
	userSvc := service.NewUserService(db, captchaSvc)
	questionSvc := service.NewQuestionService(db)
	dictionarySvc := service.NewDictionaryService(db)
	recordSvc := service.NewRecordService(redisClient, paperSvc)
	favoriteSvc := service.NewWordFavoriteService(db)

	return Deps{
		LegacyUser:         handler.NewUserHandler(userSvc),
		LegacyCaptcha:      handler.NewCaptchaHandler(captchaSvc),
		LegacyPaper:        handler.NewPaperHandler(paperSvc),
		LegacyQuestion:     handler.NewQuestionHandler(questionSvc),
		LegacyDictionary:   handler.NewDictionaryHandler(dictionarySvc),
		LegacyRecord:       handler.NewRecordHandler(recordSvc),
		LegacyWordFavorite: handler.NewWordFavoriteHandler(favoriteSvc),
		LegacyAgent:        handler.NewAgentHandler(llm.NewGlmClient(cfg.LLM.GlmKey)),
		LegacyHealth:       handler.NewHealthHandler(db),
		V1User:             handlerv1.NewUserHandler(userSvc),
		V1Captcha:          handlerv1.NewCaptchaHandler(captchaSvc),
		V1Paper:            handlerv1.NewPaperHandler(paperSvc),
		V1Question:         handlerv1.NewQuestionHandler(questionSvc),
		V1Dictionary:       handlerv1.NewDictionaryHandler(dictionarySvc),
		V1Record:           handlerv1.NewRecordHandler(recordSvc),
		V1WordFavorite:     handlerv1.NewWordFavoriteHandler(favoriteSvc),
		V1Essay:            handlerv1.NewEssayHandler(essayService),
	}
}

// essayAPIKey 选取当前 provider 对应的 API Key。
func essayAPIKey(cfg config.LLMConfig) string {
	if llm.Provider(cfg.Provider) == llm.ProviderGLM {
		return cfg.GlmKey
	}
	return cfg.DeepseekKey
}

// newEssayService 构建作文批改 chat model 并创建服务；key 缺失或 provider 非法时返回错误。
func newEssayService(cfg config.Config, redisClient *redis.Client) (*service.EssayService, error) {
	cm, err := agent.NewEssayChatModel(context.Background(), llm.Provider(cfg.LLM.Provider), essayAPIKey(cfg.LLM))
	if err != nil {
		return nil, err
	}
	return service.NewEssayService(redisClient, cm), nil
}
