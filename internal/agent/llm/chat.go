package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// Provider 表示 OpenAI 兼容的大模型服务商。
type Provider string

const (
	ProviderDeepSeek Provider = "deepseek"
	ProviderGLM      Provider = "glm"
)

// providerPreset 各服务商的接入预设，由工厂统一管理，调用方无需关心。
type providerPreset struct {
	baseURL     string
	model       string
	temperature float32
}

var providerPresets = map[Provider]providerPreset{
	ProviderDeepSeek: {baseURL: "https://api.deepseek.com/v1", model: "deepseek-chat", temperature: 0.2},
	ProviderGLM:      {baseURL: "https://open.bigmodel.cn/api/paas/v4", model: "glm-4-flash", temperature: 0.2},
}

// ChatModelOption 在服务商预设之上对 chat model 做额外定制。
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

// NewChatModel 按服务商预设构建 OpenAI 兼容的 chat model。
// base_url、model、temperature 均由工厂管理，调用方只需提供服务商、API Key 及可选定制。
func NewChatModel(ctx context.Context, provider Provider, apiKey string, opts ...ChatModelOption) (model.BaseChatModel, error) {
	preset, ok := providerPresets[provider]
	if !ok {
		return nil, fmt.Errorf("不支持的 LLM 服务商: %s", provider)
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("%s 的 API Key 未配置", provider)
	}

	temp := preset.temperature
	cfg := &openai.ChatModelConfig{
		BaseURL:     preset.baseURL,
		APIKey:      apiKey,
		Model:       preset.model,
		Temperature: &temp,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return openai.NewChatModel(ctx, cfg)
}
