package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"strings"
)

func (a *App) ListGitOperations(ctx context.Context, limit int) ([]domain.GitOperation, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("git operation limit: %w", ErrInvalidInput)
	}
	operations, err := a.repository.ListGitOperations(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list git operations: %w", err)
	}
	return operations, nil
}

func (a *App) ListGitEvents(ctx context.Context, rawID string) ([]domain.GitEvent, error) {
	id := domain.GitOperationID(strings.TrimSpace(rawID))
	if id == "" || string(id) != rawID {
		return nil, fmt.Errorf("git operation id: %w", ErrInvalidInput)
	}
	events, err := a.repository.ListGitEvents(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list git events: %w", err)
	}
	return events, nil
}

func (a *App) getGitOperation(ctx context.Context, rawID string) (domain.GitOperation, error) {
	id := domain.GitOperationID(strings.TrimSpace(rawID))
	if id == "" || string(id) != rawID {
		return domain.GitOperation{}, fmt.Errorf("git operation id: %w", ErrInvalidInput)
	}
	operation, err := a.repository.GetGitOperation(ctx, id)
	if err != nil {
		return domain.GitOperation{}, fmt.Errorf("get git operation: %w", err)
	}
	return operation, nil
}

func (a *App) appendGitEvent(ctx context.Context, id domain.GitOperationID, message string) error {
	event, err := domain.NewGitEvent(id, message, a.clock.Now())
	if err != nil {
		return err
	}
	if err := a.repository.AppendGitEvent(ctx, id, event.Message(), event.OccurredAt()); err != nil {
		return fmt.Errorf("append git event: %w", err)
	}
	return nil
}
