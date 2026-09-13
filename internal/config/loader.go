package config

import (
	"strings"

	"github.com/spf13/viper"
)

func Load() Config {
	v := viper.New()
	setDefaults(v)
	bindEnvs(v)
	loadConfigFile(v)
	v.SetEnvPrefix("ONEPRACTICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg = Config{
		Server:   ServerConfig{Port: v.GetString("server.port")},
		Database: DatabaseConfig{DSN: v.GetString("database.dsn")},
		Redis: RedisConfig{
			Addr:     v.GetString("redis.addr"),
			Username: v.GetString("redis.username"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
			Disabled: v.GetBool("redis.disabled"),
		},
		Auth: AuthConfig{
			TokenName: v.GetString("auth.token_name"),
			Timeout:   v.GetInt64("auth.timeout"),
		},
		Log: LogConfig{
			Dir:           v.GetString("log.dir"),
			Level:         v.GetString("log.level"),
			Console:       v.GetBool("log.console"),
			AddSource:     v.GetBool("log.add_source"),
			RetentionDays: v.GetInt("log.retention_days"),
		},
		Mail: MailConfig{
			APIKey:   v.GetString("mail.api_key"),
			From:     v.GetString("mail.from"),
			Disabled: v.GetBool("mail.disabled"),
		},
		LLM: LLMConfig{
			GlmKey:  v.GetString("llm.glm_key"),
			Default: v.GetString("llm.default"),
		},
		Points: PointsConfig{
			Defaults: PointsDefaultsConfig{
				OCRCost:          v.GetInt64("points.defaults.ocr_cost"),
				EssayCost:        v.GetInt64("points.defaults.essay_cost"),
				DailyLoginReward: v.GetInt64("points.defaults.daily_login_reward"),
			},
			ReserveTimeoutMins: v.GetInt("points.reserve_timeout_minutes"),
		},
		Turnstile: TurnstileConfig{
			Enabled:   v.GetBool("turnstile.enabled"),
			SecretKey: v.GetString("turnstile.secret_key"),
		},
	}
	// chat 模型列表为动态 map，逐字段读取无法覆盖，需整体反序列化。
	_ = v.UnmarshalKey("llm.models", &cfg.LLM.Models)
	// UnmarshalKey 不识别环境变量绑定，这里对每个模型重新解析 api_key：
	// 绑定 <NAME>_API_KEY 环境变量（如 DEEPSEEK_API_KEY / GLM_API_KEY），未设置时回退配置文件。
	for name, model := range cfg.LLM.Models {
		key := "llm.models." + name + ".api_key"
		bindEnv(v, key, strings.ToUpper(strings.ReplaceAll(name, "-", "_"))+"_API_KEY")
		model.APIKey = v.GetString(key)
		cfg.LLM.Models[name] = model
	}
	// 定时任务为动态 map，逐字段读取无法覆盖，需整体反序列化。
	_ = v.UnmarshalKey("cron", &cfg.Cron)
	return cfg
}

func bindEnvs(v *viper.Viper) {
	bindEnv(v, "server.port", "SERVER_PORT")
	bindEnv(v, "database.dsn", "MYSQL_DSN")
	bindEnv(v, "redis.addr", "REDIS_ADDR")
	bindEnv(v, "redis.username", "REDIS_USERNAME")
	bindEnv(v, "redis.password", "REDIS_PASSWORD")
	bindEnv(v, "redis.db", "REDIS_DB")
	bindEnv(v, "redis.disabled", "REDIS_DISABLED")
	bindEnv(v, "auth.token_name", "SA_TOKEN_NAME")
	bindEnv(v, "auth.timeout", "SA_TOKEN_TIMEOUT")
	bindEnv(v, "mail.api_key", "SENDFLARE_API_KEY")
	bindEnv(v, "mail.from", "SENDFLARE_FROM")
	bindEnv(v, "mail.disabled", "SENDFLARE_DISABLED")
	bindEnv(v, "llm.glm_key", "GLM_API_KEY")
	bindEnv(v, "points.defaults.ocr_cost", "POINTS_OCR_COST")
	bindEnv(v, "points.defaults.essay_cost", "POINTS_ESSAY_COST")
	bindEnv(v, "points.defaults.daily_login_reward", "POINTS_DAILY_LOGIN_REWARD")
	bindEnv(v, "points.reserve_timeout_minutes", "POINTS_RESERVE_TIMEOUT_MINUTES")
	bindEnv(v, "turnstile.enabled", "TURNSTILE_ENABLED")
	bindEnv(v, "turnstile.secret_key", "TURNSTILE_SECRET_KEY")
}

func bindEnv(v *viper.Viper, key string, envNames ...string) {
	args := append([]string{key}, envNames...)
	args = append(args, "ONEPRACTICE_"+strings.ToUpper(strings.ReplaceAll(key, ".", "_")))
	_ = v.BindEnv(args...)
}

func loadConfigFile(v *viper.Viper) {
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./golang")
	v.AddConfigPath("./golang/config")

	if path := v.GetString("config.file"); path != "" {
		v.SetConfigFile(path)
	}

	_ = v.ReadInConfig()
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("config.file", "")
	v.SetDefault("server.port", "8080")
	v.SetDefault("database.dsn", "root:Luchen1122@tcp(fn.aprdec.top)/onepractice?charset=utf8&parseTime=True&loc=Local")
	v.SetDefault("redis.addr", "fn.aprdec.top:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.disabled", false)
	v.SetDefault("auth.token_name", "token")
	v.SetDefault("auth.timeout", int64(15*24*60*60))
	v.SetDefault("log.dir", "./log")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.console", true)
	v.SetDefault("log.add_source", false)
	v.SetDefault("log.retention_days", 7)
	v.SetDefault("mail.disabled", true)
	v.SetDefault("mail.api_key", "")
	v.SetDefault("mail.from", "")
	v.SetDefault("llm.glm_key", "")
	v.SetDefault("llm.default", "")
	v.SetDefault("points.defaults.ocr_cost", 1)
	v.SetDefault("points.defaults.essay_cost", 2)
	v.SetDefault("points.defaults.daily_login_reward", 6)
	v.SetDefault("points.reserve_timeout_minutes", 10)
	v.SetDefault("turnstile.enabled", false)
	v.SetDefault("turnstile.secret_key", "")
}
