package skeleton

import (
	"context"
	"testing"
)

type mockTask struct{}

func (t *mockTask) Start() error {
	return nil
}

func (t *mockTask) Shutdown(ctx context.Context) error {
	return nil
}

func TestSupervisor_Register_TaskAdded(t *testing.T) {
	mockSupervisor := supervisor{
		tasks: make(map[string]Task),
	}

	err := mockSupervisor.Register("task", &mockTask{})
	if err != nil {
		t.Errorf("Unexpected error: [%v]", err)
	}

	if _, exists := mockSupervisor.tasks["task"]; !exists {
		t.Errorf("Expected task to be added, but it doesn't exist")
	}
}

func TestSupervisor_Register_SupervisorAlreadyRunning(t *testing.T) {
	mockSupervisor := supervisor{
		tasks:     make(map[string]Task),
		isRunning: true,
	}

	err := mockSupervisor.Register("task", &mockTask{})
	if err != ErrSupervisorAlreadyRunning {
		t.Errorf("Expected error: [%v], Got: [%v]", ErrSupervisorAlreadyRunning, err)
	}
}

func TestSupervisor_Register_DuplicatedTask(t *testing.T) {
	mockSupervisor := supervisor{
		tasks: make(map[string]Task),
	}

	_ = mockSupervisor.Register("task", &mockTask{})
	err := mockSupervisor.Register("task", &mockTask{})

	if err != ErrDuplicatedTask {
		t.Errorf("Expected error: [%v], Got: [%v]", ErrDuplicatedTask, err)
	}
}
