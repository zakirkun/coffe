package domain_test

import (
	"coffe/internal/domain"
	"testing"
	"time"
)

func Test_NewAuditEntry_accepts_structured_redacted_metadata(t *testing.T) {
	// Given
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)

	// When
	entry, err := domain.NewAuditEntry(domain.AuditValues{
		ID: "audit-1", Category: domain.AuditGit, Action: "approve",
		ResourceID: "git-1", Outcome: domain.AuditSuccess,
		Message: "git approval recorded", OccurredAt: now,
	})
	// Then
	if err != nil {
		t.Fatalf("NewAuditEntry() error = %v", err)
	}
	if entry.Category() != domain.AuditGit || entry.Outcome() != domain.AuditSuccess {
		t.Fatalf("entry category/outcome = %q/%q", entry.Category(), entry.Outcome())
	}
}
