package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
)

func (r *memoryRepository) SaveWorker(_ context.Context, worker domain.Worker) error {
	r.workers[worker.ID()] = worker
	return nil
}

func (r *memoryRepository) GetWorker(_ context.Context, id domain.WorkerID) (domain.Worker, error) {
	worker, ok := r.workers[id]
	if !ok {
		return domain.Worker{}, service.ErrNotFound
	}
	return worker, nil
}

func (r *memoryRepository) GetWorkerByTarget(
	_ context.Context,
	targetID domain.TargetID,
) (domain.Worker, error) {
	for _, worker := range r.workers {
		if worker.TargetID() == targetID {
			return worker, nil
		}
	}
	return domain.Worker{}, service.ErrNotFound
}

func (r *memoryRepository) ListWorkers(_ context.Context) ([]domain.Worker, error) {
	workers := make([]domain.Worker, 0, len(r.workers))
	for _, worker := range r.workers {
		workers = append(workers, worker)
	}
	return workers, nil
}
