package db

import (
	"log/slog"

	"onepractice-golang/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	if cfg.DSN == "" {
		slog.Warn("MYSQL_DSN is empty; database is disabled")
		return nil, nil
	}

	return gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
}
