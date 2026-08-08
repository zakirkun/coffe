package ai

import (
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"
)

type ProviderStore interface {
	ListAIProviders(context.Context) ([]domain.AIProvider, error)
	SaveAIUsage(context.Context, domain.AIUsage) error
}

type SecretReader interface {
	Get(context.Context, string) (string, error)
}

type IDSource interface {
	New() (string, error)
}

type Clock interface {
	Now() time.Time
}

type RouterDependencies struct {
	Store    ProviderStore
	Secrets  SecretReader
	IDs      IDSource
	Clock    Clock
	Fallback *OpenAIPlanner
}

type Router struct {
	store    ProviderStore
	secrets  SecretReader
	ids      IDSource
	clock    Clock
	fallback *OpenAIPlanner
}

func NewRouter(dependencies RouterDependencies) *Router {
	return &Router{
		store: dependencies.Store, secrets: dependencies.Secrets,
		ids: dependencies.IDs, clock: dependencies.Clock, fallback: dependencies.Fallback,
	}
}

func (r *Router) Classify(ctx context.Context, prompt string) (domain.Intent, error) {
	providers, err := r.store.ListAIProviders(ctx)
	if err != nil {
		return "", fmt.Errorf("list AI providers: %w", err)
	}
	var attempts []error
	for _, provider := range providers {
		if !provider.Enabled() {
			continue
		}
		intent, err := r.classifyProvider(ctx, provider, prompt)
		if err == nil {
			return intent, nil
		}
		attempts = append(attempts, fmt.Errorf("provider %s: %w", provider.ID(), err))
	}
	if len(attempts) == 0 && r.fallback != nil {
		return r.fallback.Classify(ctx, prompt)
	}
	if len(attempts) == 0 {
		return "", fmt.Errorf("no enabled AI provider")
	}
	return "", errors.Join(attempts...)
}

func (r *Router) classifyProvider(
	ctx context.Context,
	provider domain.AIProvider,
	prompt string,
) (domain.Intent, error) {
	apiKey, err := r.providerKey(ctx, provider)
	if err != nil {
		return "", err
	}
	client, err := NewOpenAIPlanner(OpenAIConfig{
		BaseURL: provider.BaseURL(), Model: provider.Model(), APIKey: apiKey, Timeout: 45 * time.Second,
	})
	if err != nil {
		return "", err
	}
	startedAt := time.Now()
	intent, usage, callErr := client.ClassifyDetailed(ctx, prompt)
	if err := r.recordUsage(ctx, usageAttempt{
		provider: provider, usage: usage, latency: time.Since(startedAt), success: callErr == nil,
	}); err != nil {
		return "", errors.Join(callErr, err)
	}
	return intent, callErr
}

func (r *Router) Models(ctx context.Context, provider domain.AIProvider) ([]string, error) {
	apiKey, err := r.providerKey(ctx, provider)
	if err != nil {
		return nil, err
	}
	client, err := NewOpenAIPlanner(OpenAIConfig{
		BaseURL: provider.BaseURL(), Model: provider.Model(), APIKey: apiKey, Timeout: 15 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return client.Models(ctx)
}

func (r *Router) providerKey(ctx context.Context, provider domain.AIProvider) (string, error) {
	if provider.SecretName() == "" {
		return "", nil
	}
	key, err := r.secrets.Get(ctx, provider.SecretName())
	if err != nil {
		return "", fmt.Errorf("get provider key: %w", err)
	}
	return key, nil
}

type usageAttempt struct {
	provider domain.AIProvider
	usage    CompletionUsage
	latency  time.Duration
	success  bool
}

func (r *Router) recordUsage(ctx context.Context, attempt usageAttempt) error {
	rawID, err := r.ids.New()
	if err != nil {
		return fmt.Errorf("new AI usage id: %w", err)
	}
	record, err := domain.NewAIUsage(domain.AIUsageValues{
		ID: domain.AIUsageID(rawID), ProviderID: attempt.provider.ID(), Model: attempt.provider.Model(),
		PromptTokens: attempt.usage.PromptTokens, CompletionTokens: attempt.usage.CompletionTokens,
		TotalTokens: attempt.usage.TotalTokens, LatencyMillis: attempt.latency.Milliseconds(),
		Success: attempt.success, CreatedAt: r.clock.Now(),
	})
	if err != nil {
		return err
	}
	if err := r.store.SaveAIUsage(ctx, record); err != nil {
		return fmt.Errorf("save AI usage: %w", err)
	}
	return nil
}
