package skeleton

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	ErrAddTaskFailed = errors.New("error adding task")

	ErrSupervisorReturnedError = errors.New("supervisor returned an error")
)

const (
	eventAppLoaded   = "app.loaded"
	eventAppStarting = "app.starting"
	eventAppStarted  = "app.started"
	eventAppFailed   = "app.failed"
	eventAppExited   = "app.exited"

	eventAppShutdownInitiated = "app.shutdown.initiated"
	eventAppShutdownCompleted = "app.shutdown.completed"
	eventAppShutdownFailed    = "app.shutdown.failed"
)

type App struct {
	config Config

	supervisor *supervisor
}

func New(cfg Config) *App {
	// use global logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.String("app", cfg.AppName),
	)

	slog.SetDefault(logger)

	slog.Info(eventAppLoaded, slog.Group("config",
		slog.String("app_name", cfg.AppName),
		slog.String("version", cfg.Version),
		slog.String("shutdown_timeout", cfg.ShutdownTimeout.String()),
	))

	return &App{
		config:     cfg,
		supervisor: newSupervisor(),
	}
}

func (a *App) Register(name string, task Task) error {
	if err := a.supervisor.Register(name, task); err != nil {
		return ErrAddTaskFailed
	}

	return nil
}

func (a *App) Run() (returnErr error) {

	startTime := time.Now()

	defer func(startTime time.Time) {
		duration := time.Since(startTime)

		if returnErr != nil {
			slog.Error(eventAppExited,
				slog.Bool("success", false),
				slog.String("error", returnErr.Error()),
				slog.Duration("duration_ns", duration))
		} else {
			slog.Info(eventAppExited,
				slog.Bool("success", true),
				slog.Duration("duration_ns", duration))
		}
	}(startTime)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	slog.Info(eventAppStarting)

	errCh := make(chan error)
	go func() {
		if err := a.supervisor.StartTasks(); err != nil {
			errCh <- err
			cancel()
		}
	}()

	slog.Info(eventAppStarted)

	select {
	case err := <-errCh:
		slog.Error(eventAppFailed,
			slog.String("error", err.Error()))
		returnErr = err
		cancel()
	case <-ctx.Done():
		cancel()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
	defer cancel()

	shutdownStartTime := time.Now()
	slog.Info(eventAppShutdownInitiated)

	if err := a.supervisor.ShutdownTasks(shutdownCtx); err != nil {
		slog.Error(eventAppShutdownFailed,
			slog.String("error", err.Error()))

		if returnErr == nil {
			returnErr = err
		}

		return
	}

	slog.Info(eventAppShutdownCompleted,
		slog.Duration("duration_ns", time.Since(shutdownStartTime)))

	return
}
