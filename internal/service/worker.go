package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"strings"
)

type RegisterWorkerInput struct {
	TargetID   string
	Name       string
	BaseURL    string
	SecretName string
	Enabled    bool
}

func (a *App) RegisterWorker(ctx context.Context, input RegisterWorkerInput) (domain.Worker, error) {
	targetID, err := parseTargetID(input.TargetID)
	if err != nil {
		return domain.Worker{}, err
	}
	if _, err := a.repository.GetTarget(ctx, targetID); err != nil {
		return domain.Worker{}, fmt.Errorf("get worker target: %w", err)
	}
	name, err := domain.NewName(input.Name)
	if err != nil {
		return domain.Worker{}, err
	}
	rawID, err := a.ids.New()
	if err != nil {
		return domain.Worker{}, fmt.Errorf("new worker id: %w", err)
	}
	worker, err := domain.NewWorker(domain.WorkerValues{
		ID: domain.WorkerID(rawID), TargetID: targetID, Name: name,
		BaseURL: input.BaseURL, SecretName: input.SecretName,
		Enabled: input.Enabled, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return domain.Worker{}, err
	}
	if err := a.repository.SaveWorker(ctx, worker); err != nil {
		return domain.Worker{}, fmt.Errorf("save worker: %w", err)
	}
	if err := a.appendAudit(ctx, auditInput{
		category: domain.AuditWorker, action: "register", resourceID: string(worker.ID()),
		outcome: domain.AuditSuccess, message: "remote worker registered",
	}); err != nil {
		return domain.Worker{}, err
	}
	return worker, nil
}

func (a *App) HeartbeatWorker(ctx context.Context, rawID string) (domain.Worker, error) {
	id := domain.WorkerID(strings.TrimSpace(rawID))
	if id == "" || string(id) != rawID {
		return domain.Worker{}, fmt.Errorf("worker id: %w", ErrInvalidInput)
	}
	worker, err := a.repository.GetWorker(ctx, id)
	if err != nil {
		return domain.Worker{}, fmt.Errorf("get worker: %w", err)
	}
	token, err := a.secrets.Get(ctx, worker.SecretName())
	if err != nil {
		return domain.Worker{}, fmt.Errorf("get worker token: %w", err)
	}
	heartbeat, err := a.workers.Ping(ctx, worker, token)
	if err != nil {
		return domain.Worker{}, fmt.Errorf("probe worker: %w", err)
	}
	updated, err := worker.WithHeartbeat(heartbeat.Version, a.clock.Now())
	if err != nil {
		return domain.Worker{}, err
	}
	if err := a.repository.SaveWorker(ctx, updated); err != nil {
		return domain.Worker{}, fmt.Errorf("save worker heartbeat: %w", err)
	}
	return updated, nil
}

func (a *App) ListWorkers(ctx context.Context) ([]domain.Worker, error) {
	workers, err := a.repository.ListWorkers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list workers: %w", err)
	}
	return workers, nil
}
