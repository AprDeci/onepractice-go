package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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

	closeOnce sync.Once
	closeErr  error
}

// New loads configuration and initializes all application dependencies.
func New() (*App, error) {
	return NewWithConfig(config.Load())
}

// NewWithConfig initializes an application with the supplied configuration.
// Keeping configuration injection separate makes application startup testable.
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

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- a.Server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return a.Close()
		}
		return errors.Join(err, a.Close())
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		shutdownErr := a.Shutdown(shutdownCtx)
		serverErr := <-serverErr
		if serverErr != nil && !errors.Is(serverErr, http.ErrServerClosed) {
			return errors.Join(shutdownErr, serverErr)
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
		a.closeErr = errors.Join(errs...)
	})

	return a.closeErr
}
