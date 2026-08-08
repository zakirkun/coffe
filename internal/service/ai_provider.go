package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"strings"
)

type CreateAIProviderInput struct {
	Name       string
	Kind       string
	BaseURL    string
	Model      string
	SecretName string
	Priority   int
	Enabled    bool
}

func (a *App) CreateAIProvider(
	ctx context.Context,
	input CreateAIProviderInput,
) (domain.AIProvider, error) {
	name, err := domain.NewName(input.Name)
	if err != nil {
		return domain.AIProvider{}, err
	}
	rawID, err := a.ids.New()
	if err != nil {
		return domain.AIProvider{}, fmt.Errorf("new AI provider id: %w", err)
	}
	provider, err := domain.NewAIProvider(domain.AIProviderValues{
		ID: domain.AIProviderID(rawID), Name: name,
		Kind:    domain.AIProviderKind(strings.TrimSpace(input.Kind)),
		BaseURL: input.BaseURL, Model: input.Model, SecretName: input.SecretName,
		Priority: input.Priority, Enabled: input.Enabled, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return domain.AIProvider{}, err
	}
	if err := a.repository.SaveAIProvider(ctx, provider); err != nil {
		return domain.AIProvider{}, fmt.Errorf("save AI provider: %w", err)
	}
	if err := a.appendAudit(ctx, auditInput{
		category: domain.AuditAI, action: "provider_create", resourceID: string(provider.ID()),
		outcome: domain.AuditSuccess, message: "AI provider registered",
	}); err != nil {
		return domain.AIProvider{}, err
	}
	return provider, nil
}

func (a *App) ListAIProviders(ctx context.Context) ([]domain.AIProvider, error) {
	providers, err := a.repository.ListAIProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("list AI providers: %w", err)
	}
	return providers, nil
}

func (a *App) ListAIModels(ctx context.Context, rawID string) ([]string, error) {
	id := domain.AIProviderID(strings.TrimSpace(rawID))
	if id == "" || string(id) != rawID {
		return nil, fmt.Errorf("AI provider id: %w", ErrInvalidInput)
	}
	provider, err := a.repository.GetAIProvider(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get AI provider: %w", err)
	}
	models, err := a.aiModels.Models(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("list AI models: %w", err)
	}
	return models, nil
}

func (a *App) ListAIUsage(ctx context.Context, limit int) ([]domain.AIUsage, error) {
	if limit < 1 || limit > 500 {
		return nil, fmt.Errorf("AI usage limit: %w", ErrInvalidInput)
	}
	usage, err := a.repository.ListAIUsage(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list AI usage: %w", err)
	}
	return usage, nil
}
