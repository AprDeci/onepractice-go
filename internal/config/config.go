package config

import "fmt"

type Config struct {
	Server   ServerConfig
	Log      LogConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Mail     MailConfig
	LLM      LLMConfig
	Cron     CronConfig `mapstructure:"cron"`
	Points   PointsConfig
}

// PointsConfig 描述积分系统的默认单价与预留订单超时时间。
type PointsConfig struct {
	Defaults           PointsDefaultsConfig `mapstructure:"defaults"`
	ReserveTimeoutMins int                  `mapstructure:"reserve_timeout_minutes"`
}

// PointsDefaultsConfig 是未配置 point_rules 时各动作的积分单价默认值。
type PointsDefaultsConfig struct {
	OCRCost          int64 `mapstructure:"ocr_cost"`
	EssayCost        int64 `mapstructure:"essay_cost"`
	DailyLoginReward int64 `mapstructure:"daily_login_reward"`
}

// Validate 校验积分默认单价必须为正数，避免出现零价或负价扣费。
func (c PointsConfig) Validate() error {
	if c.Defaults.OCRCost <= 0 {
		return fmt.Errorf("points.defaults.ocr_cost must be > 0")
	}
	if c.Defaults.EssayCost <= 0 {
		return fmt.Errorf("points.defaults.essay_cost must be > 0")
	}
	if c.Defaults.DailyLoginReward <= 0 {
		return fmt.Errorf("points.defaults.daily_login_reward must be > 0")
	}
	return nil
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
