package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
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

func Load() Config {
	return Config{
		Server: ServerConfig{
			Port: getenv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			DSN: os.Getenv("MYSQL_DSN"),
		},
		Auth: AuthConfig{
			TokenName: getenv("SA_TOKEN_NAME", "Authorization"),
			Timeout:   getenvInt64("SA_TOKEN_TIMEOUT", 15*24*60*60),
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
