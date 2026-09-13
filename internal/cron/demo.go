package cron

import (
	"context"
	"log/slog"
)

// registerDemo 注册示例定时任务，调度配置对应 cron.demo。
func registerDemo(manager *Manager, logger *slog.Logger) error {
	return manager.AddConfiguredFunc("demo", func(_ context.Context) {
		logger.Info("示例定时任务执行")
	})
}
