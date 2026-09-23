package app

import (
	"log/slog"
	"onepractice-golang/internal/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	if cfg.DSN == "" {
		slog.Warn("MYSQL_DSN is empty; database is disabled")
		return nil, nil
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)                  // 空闲连接数
	sqlDB.SetMaxOpenConns(100)                 // 最大连接数（< MySQL max_connections）
	sqlDB.SetConnMaxLifetime(time.Hour)        // 1小时强制重建，避开 wait_timeout
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // 30分钟回收空闲连接
	return db, nil
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
