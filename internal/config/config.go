package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Mail     MailConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
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
	return Config{
		Server: ServerConfig{
			Port: getenv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			DSN: getenv("MYSQL_DSN", "root:Luchen1122@tcp(fn.aprdec.top)/onepractice?charset=utf8&parseTime=True&loc=Local"),
		},
		Auth: AuthConfig{
			TokenName: getenv("SA_TOKEN_NAME", "Authorization"),
			Timeout:   getenvInt64("SA_TOKEN_TIMEOUT", 15*24*60*60),
		},
		Mail: MailConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     int(getenvInt64("SMTP_PORT", 465)),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     getenv("SMTP_FROM", os.Getenv("SMTP_USERNAME")),
			Name:     getenv("SMTP_NAME", "onepractice"),
			Disabled: getenv("SMTP_DISABLED", "true") == "true",
		},
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
