package cron

import (
	"log/slog"

	"onepractice-golang/internal/service"
)

// Register 集中注册脚手架内置的定时任务。
func Register(manager *Manager, logger *slog.Logger, points *service.PointsService) error {
	if err := registerDemo(manager, logger); err != nil {
		return err
	}
	return registerPointsReserveSweep(manager, logger, points)
}
