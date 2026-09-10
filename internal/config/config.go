package config

type Config struct {
	Server   ServerConfig
	Log      LogConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Mail     MailConfig
	LLM      LLMConfig
	Cron     CronConfig `mapstructure:"cron"`
}

// LLMConfig 描述 chat 模型列表与默认使用的模型；GlmKey 仅用于 OCR。
type LLMConfig struct {
	// GlmKey 供 OCR（layout_parsing）使用，与 chat 模型列表相互独立。
	GlmKey  string                    `mapstructure:"glm_key"`
	Default string                    `mapstructure:"default"`
	Models  map[string]LLMModelConfig `mapstructure:"models"`
}

// LLMModelConfig 描述一个 OpenAI 兼容 chat 模型的接入参数。
type LLMModelConfig struct {
	BaseURL     string   `mapstructure:"base_url"`
	Model       string   `mapstructure:"model"`
	APIKey      string   `mapstructure:"api_key"`
	Temperature *float32 `mapstructure:"temperature"`
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
	Disabled bool
}

type AuthConfig struct {
	TokenName string
	Timeout   int64
}

type MailConfig struct {
	APIKey   string
	From     string
	Disabled bool
}

// LogConfig 描述日志目录、级别和输出方式。
type LogConfig struct {
	Dir           string `mapstructure:"dir"`
	Level         string `mapstructure:"level"`
	Console       bool   `mapstructure:"console"`
	AddSource     bool   `mapstructure:"add_source"`
	RetentionDays int    `mapstructure:"retention_days"`
}

// CORSConfig 描述允许跨域访问的来源和凭证策略。
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

type CronTaskConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Expression string `mapstructure:"expression"`
}

type CronConfig map[string]CronTaskConfig
