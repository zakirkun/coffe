package service_test

import (
	"coffe/internal/domain"
	"context"
	"slices"
	"time"
)

func (r *memoryRepository) AppendAudit(_ context.Context, entry domain.AuditEntry) error {
	r.audit = append(r.audit, entry)
	return nil
}

func (r *memoryRepository) ListAudit(_ context.Context, limit int) ([]domain.AuditEntry, error) {
	return slices.Clone(r.audit[:min(limit, len(r.audit))]), nil
}

func (r *memoryRepository) DeleteAuditBefore(_ context.Context, before time.Time) error {
	r.auditDeleteAt = before
	kept := r.audit[:0]
	for _, entry := range r.audit {
		if !entry.OccurredAt().Before(before) {
			kept = append(kept, entry)
		}
	}
	r.audit = kept
	return nil
}
