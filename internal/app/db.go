package app

import (
	"onepractice-golang/internal/config"
	"onepractice-golang/internal/db"

	"gorm.io/gorm"
)

func openDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	return db.Open(cfg)
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
