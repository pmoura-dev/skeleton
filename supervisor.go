package skeleton

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
)

var (
	ErrSupervisorAlreadyRunning = errors.New("supervisor already running")
	ErrTaskFailed               = errors.New("a task has failed")
	ErrTaskShutdownFailed       = errors.New("task shutdown has failed")
	ErrDuplicatedTask           = errors.New("task already exists")
	ErrShutdownDeadlineExceeded = errors.New("shutdown deadline exceeded")
)

const (
	eventTaskAdded   = "task.added"
	eventTaskStarted = "task.started"
	eventTaskFailed  = "task.failed"

	eventTaskShutdownInitiated = "task.shutdown.initiated"
	eventTaskShutdownCompleted = "task.shutdown.completed"
	eventTaskShutdownFailed    = "task.shutdown.failed"
)

type Task interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type Supervisor interface {
	AddTask(name string, task Task) error
	StartTasks() error
	ShutdownTasks(ctx context.Context) error
}

type supervisor struct {
	tasks map[string]Task

	isRunning bool
}

func newSupervisor() *supervisor {
	return &supervisor{
		tasks: make(map[string]Task),
	}
}

func (s *supervisor) Register(name string, task Task) error {
	if s.isRunning {
		return ErrSupervisorAlreadyRunning
	}

	if _, exists := s.tasks[name]; exists {
		return ErrDuplicatedTask
	}

	s.tasks[name] = task
	slog.Info(eventTaskAdded, slog.String("task", name))
	return nil
}

func (s *supervisor) StartTasks() error {
	errCh := make(chan error)

	for name, task := range s.tasks {
		go func() {
			if err := task.Start(); err != nil {
				slog.Error(eventTaskFailed,
					slog.String("task", name),
					slog.String("error", err.Error()),
				)
				errCh <- ErrTaskFailed
				return
			}
		}()

		slog.Info(eventTaskStarted, slog.String("task", name))
	}

	return <-errCh
}

func (s *supervisor) ShutdownTasks(ctx context.Context) error {
	var (
		wg           sync.WaitGroup
		failedStatus atomic.Bool
	)

	for name, task := range s.tasks {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := task.Shutdown(ctx); err != nil {
				slog.Error(eventTaskShutdownFailed,
					slog.String("task", name),
					slog.String("error", err.Error()))

				failedStatus.Store(true)
				return
			}

			slog.Info(eventTaskShutdownCompleted, slog.String("task", name))
		}()

		slog.Info(eventTaskShutdownInitiated, slog.String("task", name))
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if failedStatus.Load() {
			return ErrTaskShutdownFailed
		}
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ErrShutdownDeadlineExceeded
		}

		return ctx.Err()
	}

	return nil
}
