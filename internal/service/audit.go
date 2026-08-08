package service

import (
	"bytes"
	"coffe/internal/domain"
	"context"
	"encoding/csv"
	"fmt"
	"time"
)

type auditInput struct {
	category   domain.AuditCategory
	action     string
	resourceID string
	outcome    domain.AuditOutcome
	message    string
}

type AuditRecordInput struct {
	Category   domain.AuditCategory
	Action     string
	ResourceID string
	Outcome    domain.AuditOutcome
	Message    string
}

func (a *App) RecordAudit(ctx context.Context, input AuditRecordInput) error {
	return a.appendAudit(ctx, auditInput{
		category: input.Category, action: input.Action, resourceID: input.ResourceID,
		outcome: input.Outcome, message: input.Message,
	})
}

func (a *App) appendAudit(ctx context.Context, input auditInput) error {
	rawID, err := a.ids.New()
	if err != nil {
		return fmt.Errorf("new audit id: %w", err)
	}
	entry, err := domain.NewAuditEntry(domain.AuditValues{
		ID: domain.AuditID(rawID), Category: input.category, Action: input.action,
		ResourceID: input.resourceID, Outcome: input.outcome,
		Message: input.message, OccurredAt: a.clock.Now(),
	})
	if err != nil {
		return err
	}
	if err := a.repository.AppendAudit(ctx, entry); err != nil {
		return fmt.Errorf("append audit: %w", err)
	}
	return nil
}

func (a *App) ListAudit(ctx context.Context, limit int) ([]domain.AuditEntry, error) {
	if limit < 1 || limit > 1000 {
		return nil, fmt.Errorf("audit limit: %w", ErrInvalidInput)
	}
	entries, err := a.repository.ListAudit(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit: %w", err)
	}
	return entries, nil
}

func (a *App) PruneAudit(ctx context.Context) error {
	policy, err := a.repository.GetPolicy(ctx)
	if err != nil {
		return fmt.Errorf("get audit policy: %w", err)
	}
	before := a.clock.Now().Add(-time.Duration(policy.AuditRetentionDays()) * 24 * time.Hour)
	if err := a.repository.DeleteAuditBefore(ctx, before); err != nil {
		return fmt.Errorf("prune audit: %w", err)
	}
	return nil
}

func (a *App) ExportAuditCSV(ctx context.Context, limit int) ([]byte, error) {
	entries, err := a.ListAudit(ctx, limit)
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write([]string{
		"occurred_at", "category", "action", "resource_id", "outcome", "message",
	}); err != nil {
		return nil, fmt.Errorf("write audit CSV header: %w", err)
	}
	for _, entry := range entries {
		if err := writer.Write([]string{
			entry.OccurredAt().Format(time.RFC3339Nano), string(entry.Category()), entry.Action(),
			safeCSVCell(entry.ResourceID()), string(entry.Outcome()), safeCSVCell(entry.Message()),
		}); err != nil {
			return nil, fmt.Errorf("write audit CSV row %s: %w", entry.ID(), err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush audit CSV: %w", err)
	}
	return output.Bytes(), nil
}
