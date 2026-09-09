package cron

import "log/slog"

// Register 集中注册脚手架内置的定时任务。
func Register(manager *Manager, logger *slog.Logger) error {
	return registerDemo(manager, logger)
}
