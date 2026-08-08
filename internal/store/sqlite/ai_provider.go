package sqlite

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *Store) SaveAIProvider(ctx context.Context, provider domain.AIProvider) error {
	const query = `INSERT INTO ai_providers (
		id, name, kind, base_url, model, secret_name, priority, enabled, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name, kind = excluded.kind, base_url = excluded.base_url,
		model = excluded.model, secret_name = excluded.secret_name,
		priority = excluded.priority, enabled = excluded.enabled`
	if _, err := s.db.ExecContext(ctx, query,
		provider.ID(), provider.Name().String(), provider.Kind(), provider.BaseURL(), provider.Model(),
		provider.SecretName(), provider.Priority(), provider.Enabled(), formatTime(provider.CreatedAt()),
	); err != nil {
		return fmt.Errorf("save AI provider: %w", err)
	}
	return nil
}

func (s *Store) GetAIProvider(ctx context.Context, id domain.AIProviderID) (domain.AIProvider, error) {
	const query = `SELECT id, name, kind, base_url, model, secret_name, priority, enabled, created_at
		FROM ai_providers WHERE id = ?`
	provider, err := scanAIProvider(s.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AIProvider{}, fmt.Errorf("AI provider %s: %w", id, service.ErrNotFound)
	}
	if err != nil {
		return domain.AIProvider{}, err
	}
	return provider, nil
}

func (s *Store) ListAIProviders(ctx context.Context) (providers []domain.AIProvider, err error) {
	const query = `SELECT id, name, kind, base_url, model, secret_name, priority, enabled, created_at
		FROM ai_providers ORDER BY priority, id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query AI providers: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		provider, err := scanAIProvider(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI providers: %w", err)
	}
	return providers, nil
}

func (s *Store) SaveAIUsage(ctx context.Context, usage domain.AIUsage) error {
	const query = `INSERT INTO ai_usage (
		id, provider_id, model, prompt_tokens, completion_tokens, total_tokens,
		latency_millis, success, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, query,
		usage.ID(), usage.ProviderID(), usage.Model(), usage.PromptTokens(), usage.CompletionTokens(),
		usage.TotalTokens(), usage.LatencyMillis(), usage.Success(), formatTime(usage.CreatedAt()),
	); err != nil {
		return fmt.Errorf("save AI usage: %w", err)
	}
	return nil
}

func (s *Store) ListAIUsage(ctx context.Context, limit int) (usage []domain.AIUsage, err error) {
	const query = `SELECT id, provider_id, model, prompt_tokens, completion_tokens,
		total_tokens, latency_millis, success, created_at
		FROM ai_usage ORDER BY created_at DESC, id DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query AI usage: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var id, providerID, model, rawCreatedAt string
		var promptTokens, completionTokens, totalTokens int
		var latencyMillis int64
		var success bool
		if err := rows.Scan(
			&id, &providerID, &model, &promptTokens, &completionTokens,
			&totalTokens, &latencyMillis, &success, &rawCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan AI usage: %w", err)
		}
		createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse AI usage created_at: %w", err)
		}
		record, err := domain.NewAIUsage(domain.AIUsageValues{
			ID: domain.AIUsageID(id), ProviderID: domain.AIProviderID(providerID), Model: model,
			PromptTokens: promptTokens, CompletionTokens: completionTokens, TotalTokens: totalTokens,
			LatencyMillis: latencyMillis, Success: success, CreatedAt: createdAt,
		})
		if err != nil {
			return nil, err
		}
		usage = append(usage, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI usage: %w", err)
	}
	return usage, nil
}

func scanAIProvider(row rowScanner) (domain.AIProvider, error) {
	var id, rawName, rawKind, baseURL, model, secretName, rawCreatedAt string
	var priority int
	var enabled bool
	if err := row.Scan(
		&id, &rawName, &rawKind, &baseURL, &model, &secretName, &priority, &enabled, &rawCreatedAt,
	); err != nil {
		return domain.AIProvider{}, fmt.Errorf("scan AI provider: %w", err)
	}
	name, err := domain.NewName(rawName)
	if err != nil {
		return domain.AIProvider{}, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return domain.AIProvider{}, fmt.Errorf("parse AI provider created_at: %w", err)
	}
	provider, err := domain.NewAIProvider(domain.AIProviderValues{
		ID: domain.AIProviderID(id), Name: name, Kind: domain.AIProviderKind(rawKind),
		BaseURL: baseURL, Model: model, SecretName: secretName,
		Priority: priority, Enabled: enabled, CreatedAt: createdAt,
	})
	if err != nil {
		return domain.AIProvider{}, fmt.Errorf("restore AI provider: %w", err)
	}
	return provider, nil
}
