package app

import (
	"log/slog"
	"onepractice-golang/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	if cfg.DSN == "" {
		slog.Warn("MYSQL_DSN is empty; database is disabled")
		return nil, nil
	}

	return gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
}

func closeDatabase(database *gorm.DB) error {
	if database == nil {
		return nil
	}

	sqlDB, err := database.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
