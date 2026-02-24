// Package skeleton provides a minimal service skeleton with
// startup, shutdown and internal HTTP server management helpers.
//
// It is intended as a small framework to demonstrate graceful
// startup and shutdown of long-running services.
package skeleton

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	eventStartupInitiated  = "service.startup.initiated"
	eventStartupCompleted  = "service.startup.completed"
	eventStartupFailed     = "service.startup.failed"
	eventShutdownInitiated = "service.shutdown.initiated"
	eventShutdownCompleted = "service.shutdown.completed"
	eventShutdownFailed    = "service.shutdown.failed"
	eventFatalError        = "service.fatal_error"
	eventExited            = "service.exited"

	reasonNormal          = "normal"
	reasonUnexpectedError = "unexpected_error"
	reasonStartupFailed   = "startup_failed"
	reasonShutdownFailed  = "shutdown_failed"

	statusSuccess = "success"
	statusFailure = "failure"
)

// Config holds configuration values for a Service.
//
// Fields include the service name, the internal server address
// and the timeout used during shutdown.
type Config struct {
	ServiceName string

	InternalAddr string

	ShutdownTimeout time.Duration
}

// Service represents the running service and its managed resources.
//
// It contains the configured values, the internal HTTP server
// and a logger used for emitting lifecycle events.
type Service struct {
	config Config

	internalServer *http.Server

	logger *slog.Logger
}

// New returns a pointer to a Service initialized with the
// provided Config. The created Service is not started; call
// `Run` to start and manage the service lifecycle.
func New(config Config) *Service {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.String("service", config.ServiceName),
	)

	return &Service{
		config: config,
		logger: logger,
	}
}

// Run starts the service, listens for shutdown signals and
// orchestrates graceful shutdown. It returns an error when
// startup or shutdown fails.
func (s *Service) Run() (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	status, reason := statusSuccess, reasonNormal
	defer func() {
		s.exit(status, reason)
	}()

	// startup service
	errCh := make(chan error)

	if err := s.startup(ctx, errCh); err != nil {
		status, reason = statusFailure, reasonStartupFailed
		return err
	}

	select {
	case <-ctx.Done():
	case err := <-errCh:
		s.logger.Error(eventFatalError, slog.String("error", err.Error()))
		status, reason = statusFailure, reasonUnexpectedError
	}

	// shutdown service
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.shutdown(shutdownCtx); err != nil {
		status = statusFailure
		if reason != reasonNormal {
			reason = reason + "; " + reasonShutdownFailed
		} else {
			reason = reasonShutdownFailed
		}

		return err
	}

	return nil
}

// startup performs the service startup sequence. It initializes
// required resources such as the internal HTTP server and reports
// any startup error on the provided channel.
func (s *Service) startup(ctx context.Context, errCh chan<- error) (err error) {
	s.logger.Info(eventStartupInitiated)

	defer func() {
		if err != nil {
			s.logger.Error(eventStartupFailed, slog.String("error", err.Error()))
		}
	}()

	go func() {
		if err := s.initInternalServer(ctx); err != nil {
			errCh <- err
		}
	}()

	s.logger.Info(eventStartupCompleted)
	return
}

// shutdown attempts to gracefully stop service resources within
// the provided context deadline. It returns an error if shutdown
// fails for any managed component.
func (s *Service) shutdown(ctx context.Context) (err error) {
	startTime := time.Now()
	s.logger.Info(eventShutdownInitiated,
		slog.String("timeout_duration", s.config.ShutdownTimeout.String()))

	defer func() {
		if err != nil {
			s.logger.Error(eventShutdownFailed, slog.String("error", err.Error()))
		}
	}()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.internalServer.Shutdown(ctx)
	})

	if err = g.Wait(); err != nil {
		return err
	}

	s.logger.Info(eventShutdownCompleted,
		slog.String("duration", time.Since(startTime).String()))

	return nil
}

// exit performs final cleanup and logs the service exit `status`
// and `reason`. It closes the internal server if present.
func (s *Service) exit(status string, reason string) {

	// close internal server
	_ = s.internalServer.Close()

	s.logger.Info(eventExited,
		slog.String("status", status),
		slog.String("reason", reason))
}

// initInternalServer initializes and runs the internal HTTP server
// that exposes administrative endpoints (for example, health checks).
// It blocks while the server is running and returns an error if the
// server fails unexpectedly.
func (s *Service) initInternalServer(_ context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthzHandler)

	s.internalServer = &http.Server{
		Addr:    s.config.InternalAddr,
		Handler: mux,
	}

	if err := s.internalServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
