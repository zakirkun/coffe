package worker_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"coffe/internal/worker"
	"context"
	"testing"
)

func Test_HybridExecutor_falls_back_when_target_has_no_worker(t *testing.T) {
	// Given
	fallback := &fallbackExecutor{}
	executor := worker.NewHybridExecutor(worker.HybridDependencies{
		Registry: emptyRegistry{}, Secrets: emptySecrets{}, Client: worker.NewClient(worker.ClientConfig{}),
		Fallback: fallback,
	})

	// When
	_, err := executor.Execute(context.Background(), service.ExecutionStep{})
	// Then
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if fallback.calls != 1 {
		t.Fatalf("fallback calls = %d, want 1", fallback.calls)
	}
}

type emptyRegistry struct{}

func (emptyRegistry) GetWorkerByTarget(context.Context, domain.TargetID) (domain.Worker, error) {
	return domain.Worker{}, service.ErrNotFound
}

type emptySecrets struct{}

func (emptySecrets) Get(context.Context, string) (string, error) { return "", nil }

type fallbackExecutor struct{ calls int }

func (e *fallbackExecutor) Execute(
	context.Context,
	service.ExecutionStep,
) (service.StepResult, error) {
	e.calls++
	return service.StepResult{}, nil
}
