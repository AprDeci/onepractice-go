package cron

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync"

	"onepractice-golang/internal/config"

	robfigcron "github.com/robfig/cron/v3"
)

// Manager 管理定时任务的注册、启动、停止和运行上下文。
type Manager struct {
	cron    *robfigcron.Cron
	cfg     config.CronConfig
	logger  *slog.Logger
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
}

// NewManager 创建使用标准五段式表达式的定时任务管理器。
func NewManager(cfg config.CronConfig, logger *slog.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		cron:   robfigcron.New(),
		cfg:    cfg,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddFunc 注册具名任务，并统一记录执行日志和恢复任务 panic。
func (m *Manager) AddFunc(spec string, name string, fn func(context.Context)) error {
	_, err := m.cron.AddFunc(spec, func() {
		m.logger.Info("定时任务开始", slog.String("job", name))
		defer func() {
			if recovered := recover(); recovered != nil {
				m.logger.Error("定时任务发生 panic",
					slog.String("job", name),
					slog.Any("panic", recovered),
					slog.String("stack", string(debug.Stack())),
				)
				return
			}
			m.logger.Info("定时任务结束", slog.String("job", name))
		}()
		fn(m.runContext())
	})
	return err
}

// AddConfiguredFunc 按任务名称读取配置，仅注册已启用且表达式有效的任务。
func (m *Manager) AddConfiguredFunc(name string, fn func(context.Context)) error {
	task, exists := m.cfg[name]
	if !exists {
		return fmt.Errorf("配置缺失: cron.%s", name)
	}
	if !task.Enabled {
		return nil
	}

	expression := strings.TrimSpace(task.Expression)
	if expression == "" {
		return fmt.Errorf("配置缺失: cron.%s.expression", name)
	}
	if err := m.AddFunc(expression, name, fn); err != nil {
		return fmt.Errorf("配置错误: cron.%s.expression: %w", name, err)
	}
	return nil
}

// Start 在存在已注册任务且尚未启动时开始调度。
func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.cron.Entries()) == 0 || m.started {
		return
	}
	if m.ctx.Err() != nil {
		m.ctx, m.cancel = context.WithCancel(context.Background())
	}
	m.cron.Start()
	m.started = true
	m.logger.Info("定时任务管理器启动成功")
}

// Stop 取消任务运行上下文并停止调度，返回可等待在途任务结束的上下文。
func (m *Manager) Stop() context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		m.logger.Info("定时任务管理器停止中")
	}
	m.cancel()
	ctx := m.cron.Stop()
	m.started = false
	return ctx
}

func (m *Manager) runContext() context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ctx
}
