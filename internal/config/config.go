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

// LLMConfig 描述大模型服务的访问凭证。
type LLMConfig struct {
	GlmKey      string `mapstructure:"glm_key"`
	DeepseekKey string `mapstructure:"deepseek_key"`
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
