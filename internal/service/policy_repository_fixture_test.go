package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
)

func (r *memoryRepository) SavePolicy(_ context.Context, policy domain.Policy) error {
	r.policy = policy
	r.hasPolicy = true
	return nil
}

func (r *memoryRepository) GetPolicy(_ context.Context) (domain.Policy, error) {
	if !r.hasPolicy {
		return domain.Policy{}, service.ErrNotFound
	}
	return r.policy, nil
}
