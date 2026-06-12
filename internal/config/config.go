package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Mail     MailConfig
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
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Name     string
	Disabled bool
}

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
		Mail: MailConfig{
			Host:     v.GetString("mail.host"),
			Port:     v.GetInt("mail.port"),
			Username: v.GetString("mail.username"),
			Password: v.GetString("mail.password"),
			From:     v.GetString("mail.from"),
			Name:     v.GetString("mail.name"),
			Disabled: v.GetBool("mail.disabled"),
		},
	}
	fmt.Println(cfg)
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
	bindEnv(v, "mail.host", "SMTP_HOST")
	bindEnv(v, "mail.port", "SMTP_PORT")
	bindEnv(v, "mail.username", "SMTP_USERNAME")
	bindEnv(v, "mail.password", "SMTP_PASSWORD")
	bindEnv(v, "mail.from", "SMTP_FROM")
	bindEnv(v, "mail.name", "SMTP_NAME")
	bindEnv(v, "mail.disabled", "SMTP_DISABLED")
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
	v.SetDefault("auth.token_name", "Authorization")
	v.SetDefault("auth.timeout", int64(15*24*60*60))
	v.SetDefault("mail.port", 465)
	v.SetDefault("mail.name", "onepractice")
	v.SetDefault("mail.disabled", true)
	v.SetDefault("mail.from", "")
	v.SetDefault("mail.host", "")
	v.SetDefault("mail.username", "")
	v.SetDefault("mail.password", "")
}
