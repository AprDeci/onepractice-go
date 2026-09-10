package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// defaultTemperature 未显式配置 temperature 时使用的默认值。
const defaultTemperature float32 = 0.2

// ModelConfig 描述一个 OpenAI 兼容 chat 模型的接入参数。
type ModelConfig struct {
	BaseURL     string
	Model       string
	APIKey      string
	Temperature *float32
}

// ChatModelOption 在模型参数之上对 chat model 做额外定制。
type ChatModelOption func(*openai.ChatModelConfig)

// WithJSONOutput 开启 JSON 输出模式（response_format=json_object）。
func WithJSONOutput() ChatModelOption {
	return func(cfg *openai.ChatModelConfig) {
		cfg.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}
}

// WithMaxTokens 限制单次生成的最大 token 数，避免长 JSON 被截断。
func WithMaxTokens(n int) ChatModelOption {
	return func(cfg *openai.ChatModelConfig) {
		cfg.MaxTokens = &n
	}
}

// NewChatModel 根据配置构建 OpenAI 兼容的 chat model。
// base_url、model 必填；temperature 省略时使用 0.2。
func NewChatModel(ctx context.Context, cfg ModelConfig, opts ...ChatModelOption) (model.BaseChatModel, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, fmt.Errorf("base_url 未配置")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, fmt.Errorf("model 未配置")
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("api_key 未配置")
	}

	temp := defaultTemperature
	if cfg.Temperature != nil {
		temp = *cfg.Temperature
	}

	chatCfg := &openai.ChatModelConfig{
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		Temperature: &temp,
	}
	for _, opt := range opts {
		opt(chatCfg)
	}

	return openai.NewChatModel(ctx, chatCfg)
}
