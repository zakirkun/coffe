package worker

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"errors"
	"fmt"
)

type Registry interface {
	GetWorkerByTarget(context.Context, domain.TargetID) (domain.Worker, error)
}

type HybridExecutor struct {
	registry Registry
	secrets  service.SecretReader
	client   *Client
	fallback service.Executor
}

type HybridDependencies struct {
	Registry Registry
	Secrets  service.SecretReader
	Client   *Client
	Fallback service.Executor
}

func NewHybridExecutor(dependencies HybridDependencies) *HybridExecutor {
	return &HybridExecutor{
		registry: dependencies.Registry, secrets: dependencies.Secrets,
		client: dependencies.Client, fallback: dependencies.Fallback,
	}
}

func (e *HybridExecutor) Execute(
	ctx context.Context,
	execution service.ExecutionStep,
) (service.StepResult, error) {
	worker, err := e.registry.GetWorkerByTarget(ctx, execution.Target.ID())
	if errors.Is(err, service.ErrNotFound) || (!worker.Enabled() && err == nil) {
		return e.fallback.Execute(ctx, execution)
	}
	if err != nil {
		return service.StepResult{}, fmt.Errorf("resolve target worker: %w", err)
	}
	token, err := e.secrets.Get(ctx, worker.SecretName())
	if err != nil {
		return service.StepResult{}, fmt.Errorf("get worker token: %w", err)
	}
	return e.client.Execute(ctx, worker, token, execution)
}
