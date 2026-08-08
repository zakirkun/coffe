package service_test

import (
	"coffe/internal/domain"
	"context"
	"slices"
)

func (r *memoryRepository) SaveQueryHistory(_ context.Context, history domain.QueryHistory) error {
	r.queryHistory[history.DatabaseID()] = append(r.queryHistory[history.DatabaseID()], history)
	return nil
}

func (r *memoryRepository) ListQueryHistory(
	_ context.Context,
	databaseID domain.DatabaseID,
	limit int,
) ([]domain.QueryHistory, error) {
	history := r.queryHistory[databaseID]
	return slices.Clone(history[:min(limit, len(history))]), nil
}
