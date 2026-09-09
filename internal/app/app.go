package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"onepractice-golang/internal/auth"
	"onepractice-golang/internal/common/logger"
	"onepractice-golang/internal/common/mail"
	"onepractice-golang/internal/config"
	"onepractice-golang/internal/cron"
	"onepractice-golang/internal/router"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const shutdownTimeout = 5 * time.Second

// App owns the HTTP server and the resources it needs to serve requests.
type App struct {
	Config    config.Config
	Server    *http.Server
	DB        *gorm.DB
	Redis     *redis.Client
	Logger    *slog.Logger
	logCloser io.Closer
	Mail      *mail.Module
	Cron      *cron.Manager

	lifecycleMu sync.Mutex
	running     bool
	closed      bool

	closeOnce sync.Once
	closeErr  error
}

func New() (*App, error) {
	return NewWithConfig(config.Load())
}

func NewWithConfig(cfg config.Config) (*App, error) {
	database, err := openDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	redisClient, err := openRedis(cfg.Redis)
	if err != nil {
		if closeErr := closeDatabase(database); closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("open redis: %w", err),
				fmt.Errorf("close database: %w", closeErr),
			)
		}
		return nil, fmt.Errorf("open redis: %w", err)
	}

	logger, logCloser, err := logger.New(&cfg)
	if err != nil {
		return nil, err
	}
	slog.SetDefault(logger)

	auth.Init(cfg.Auth, redisClient)
	mailModule := mail.NewModule(context.Background(), cfg.Mail, redisClient)
	engine := router.New(cfg, database, redisClient, mailModule.Sender, logger)

	cronManager := cron.NewManager(cfg.Cron, logger)
	if err := cron.Register(cronManager, logger); err != nil {
		return nil, fmt.Errorf("注册定时任务失败: %w", err)
	}

	return &App{
		Config: cfg,
		Server: &http.Server{
			Addr:    ":" + cfg.Server.Port,
			Handler: engine,
		},
		DB:        database,
		Redis:     redisClient,
		Logger:    logger,
		logCloser: logCloser,
		Mail:      mailModule,
		Cron:      cronManager,
	}, nil
}

// Run starts the HTTP server and waits for an interrupt or termination signal.
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return a.run(ctx)
}

// RunContext is useful for callers and tests that need to control the run
// lifetime without sending an operating-system signal.
func (a *App) RunContext(ctx context.Context) error {
	return a.run(ctx)
}

func (a *App) run(ctx context.Context) error {
	if a == nil || a.Server == nil {
		return errors.New("app server is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	a.lifecycleMu.Lock()
	if a.closed {
		a.lifecycleMu.Unlock()
		return errors.New("应用已关闭")
	}
	if a.running {
		a.lifecycleMu.Unlock()
		return errors.New("应用已在运行")
	}

	// 预绑定端口，确保监听失败能同步返回，而不是在 goroutine 中异步丢失。
	listener, err := net.Listen("tcp", a.Server.Addr)
	if err != nil {
		a.lifecycleMu.Unlock()
		return fmt.Errorf("监听 HTTP 服务失败: %w", err)
	}
	a.running = true

	log := a.Logger
	if log == nil {
		log = slog.Default()
	}
	if a.Cron != nil {
		a.Cron.Start()
	}
	log.Info("HTTP 服务启动成功", slog.String("addr", a.Server.Addr))
	a.lifecycleMu.Unlock()

	serverErr := make(chan error, 1)
	go func() {
		err := a.Server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serverErr <- err
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return errors.Join(err, a.Close())
		}
		return a.Close()
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		shutdownErr := a.Shutdown(shutdownCtx)
		if serveErr := <-serverErr; serveErr != nil {
			return errors.Join(shutdownErr, serveErr)
		}
		return shutdownErr
	}
}

// Shutdown gracefully stops the HTTP server and then closes its resources.
func (a *App) Shutdown(ctx context.Context) error {
	if a == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var shutdownErr error
	if a.Server != nil {
		shutdownErr = a.Server.Shutdown(ctx)
	}
	return errors.Join(shutdownErr, a.Close())
}

// Close releases the database and Redis resources. It is safe to call more
// than once and also handles disabled resources returned as nil.
func (a *App) Close() error {
	if a == nil {
		return nil
	}

	a.closeOnce.Do(func() {
		a.lifecycleMu.Lock()
		a.closed = true
		a.running = false
		a.lifecycleMu.Unlock()

		var errs []error
		if a.Mail != nil {
			a.Mail.Close()
		}
		if err := closeDatabase(a.DB); err != nil {
			errs = append(errs, fmt.Errorf("close database: %w", err))
		}
		if a.logCloser != nil {
			if err := a.logCloser.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close logger: %w", err))
			}
		}
		if err := closeRedis(a.Redis); err != nil {
			errs = append(errs, fmt.Errorf("close redis: %w", err))
		}
		if a.Cron != nil {
			a.Cron.Stop()
		}
		a.closeErr = errors.Join(errs...)
	})

	return a.closeErr
}
