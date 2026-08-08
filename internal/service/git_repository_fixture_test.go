package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"slices"
	"time"
)

func (r *memoryRepository) SaveGitOperation(_ context.Context, operation domain.GitOperation) error {
	r.gitOperations[operation.ID()] = operation
	return nil
}

func (r *memoryRepository) GetGitOperation(
	_ context.Context,
	id domain.GitOperationID,
) (domain.GitOperation, error) {
	operation, ok := r.gitOperations[id]
	if !ok {
		return domain.GitOperation{}, service.ErrNotFound
	}
	return operation, nil
}

func (r *memoryRepository) ListGitOperations(_ context.Context, limit int) ([]domain.GitOperation, error) {
	operations := make([]domain.GitOperation, 0, min(limit, len(r.gitOperations)))
	for _, operation := range r.gitOperations {
		if len(operations) == limit {
			break
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func (r *memoryRepository) AppendGitEvent(
	_ context.Context,
	id domain.GitOperationID,
	message string,
	occurredAt time.Time,
) error {
	event, err := domain.NewGitEvent(id, message, occurredAt)
	if err != nil {
		return err
	}
	r.gitEvents[id] = append(r.gitEvents[id], event)
	return nil
}

func (r *memoryRepository) ListGitEvents(
	_ context.Context,
	id domain.GitOperationID,
) ([]domain.GitEvent, error) {
	return slices.Clone(r.gitEvents[id]), nil
}
