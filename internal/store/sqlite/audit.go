package sqlite

import (
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *Store) AppendAudit(ctx context.Context, entry domain.AuditEntry) error {
	const query = `INSERT INTO audit_log (
		id, category, action, resource_id, outcome, message, occurred_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, query,
		entry.ID(), entry.Category(), entry.Action(), entry.ResourceID(),
		entry.Outcome(), entry.Message(), formatTime(entry.OccurredAt()),
	); err != nil {
		return fmt.Errorf("append audit: %w", err)
	}
	return nil
}

func (s *Store) ListAudit(ctx context.Context, limit int) (entries []domain.AuditEntry, err error) {
	const query = `SELECT id, category, action, resource_id, outcome, message, occurred_at
		FROM audit_log ORDER BY occurred_at DESC, id DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query audit: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var id, rawCategory, action, resourceID, rawOutcome, message, rawOccurredAt string
		if err := rows.Scan(
			&id, &rawCategory, &action, &resourceID, &rawOutcome, &message, &rawOccurredAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit: %w", err)
		}
		occurredAt, err := time.Parse(time.RFC3339Nano, rawOccurredAt)
		if err != nil {
			return nil, fmt.Errorf("parse audit occurred_at: %w", err)
		}
		entry, err := domain.NewAuditEntry(domain.AuditValues{
			ID: domain.AuditID(id), Category: domain.AuditCategory(rawCategory),
			Action: action, ResourceID: resourceID, Outcome: domain.AuditOutcome(rawOutcome),
			Message: message, OccurredAt: occurredAt,
		})
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit: %w", err)
	}
	return entries, nil
}

func (s *Store) DeleteAuditBefore(ctx context.Context, before time.Time) error {
	const query = `DELETE FROM audit_log WHERE occurred_at < ?`
	if _, err := s.db.ExecContext(ctx, query, formatTime(before)); err != nil {
		return fmt.Errorf("delete expired audit: %w", err)
	}
	return nil
}
