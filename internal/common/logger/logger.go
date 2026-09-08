package logger

import (
	"fmt"
	"io"
	"log/slog"
	"onepractice-golang/internal/config"
	"os"
)

func New(cfg *config.Config) (*slog.Logger, io.Closer, error) {
	level, err := parseLevel(cfg.Log.Level)
	if err != nil {
		return nil, nil, err
	}

	handler, err := NewDailyHandler(DailyHandlerOptions{
		ProjectName:   "onepractice",
		BaseDir:       ".",
		Dir:           cfg.Log.Dir,
		Level:         level,
		AddSource:     cfg.Log.AddSource,
		Console:       cfg.Log.Console,
		ConsoleOut:    os.Stdout,
		RetentionDays: cfg.Log.RetentionDays,
	})
	if err != nil {
		return nil, nil, err
	}
	return slog.New(handler), handler, nil
}

// parseLevel 将配置文件中的日志级别字符串解析为 slog.Level。
// 支持的值为 "debug"、"info"、"warn"、"error"。
// 不区分大小写，默认返回 slog.LevelInfo 并报错。
func parseLevel(raw string) (slog.Level, error) {
	switch raw {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("配置错误: log.level 仅支持 debug、info、warn、error")
	}
}
