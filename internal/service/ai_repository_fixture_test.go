package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"slices"
)

func (r *memoryRepository) SaveAIProvider(_ context.Context, provider domain.AIProvider) error {
	r.aiProviders[provider.ID()] = provider
	return nil
}

func (r *memoryRepository) GetAIProvider(
	_ context.Context,
	id domain.AIProviderID,
) (domain.AIProvider, error) {
	provider, ok := r.aiProviders[id]
	if !ok {
		return domain.AIProvider{}, service.ErrNotFound
	}
	return provider, nil
}

func (r *memoryRepository) ListAIProviders(_ context.Context) ([]domain.AIProvider, error) {
	providers := make([]domain.AIProvider, 0, len(r.aiProviders))
	for _, provider := range r.aiProviders {
		providers = append(providers, provider)
	}
	return providers, nil
}

func (r *memoryRepository) SaveAIUsage(_ context.Context, usage domain.AIUsage) error {
	r.aiUsage = append(r.aiUsage, usage)
	return nil
}

func (r *memoryRepository) ListAIUsage(_ context.Context, limit int) ([]domain.AIUsage, error) {
	return slices.Clone(r.aiUsage[:min(limit, len(r.aiUsage))]), nil
}
