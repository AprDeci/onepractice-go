package cron

import (
	"context"
	"log/slog"

	"onepractice-golang/internal/service"
)

// registerPointsReserveSweep 注册积分预扣补偿任务：扫描超时未结算的作文扣费并退款。
// 调度配置对应 cron.points_reserve_expiry。
func registerPointsReserveSweep(manager *Manager, logger *slog.Logger, points *service.PointsService) error {
	if points == nil {
		return nil
	}
	return manager.AddConfiguredFunc("points_reserve_expiry", func(ctx context.Context) {
		refunded, err := points.SweepExpiredReserved(ctx)
		if err != nil {
			logger.Error("积分预扣补偿失败", slog.Any("err", err))
			return
		}
		if refunded > 0 {
			logger.Info("积分预扣补偿完成", slog.Int("refunded", refunded))
		}
	})
}
