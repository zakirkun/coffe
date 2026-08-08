package domain_test

import (
	"coffe/internal/domain"
	"testing"
	"time"
)

func Test_NewDeploymentRecord_preserves_revisions_for_rollback(t *testing.T) {
	// Given
	before, err := domain.NewGitRevision("0123456789abcdef0123456789abcdef01234567")
	if err != nil {
		t.Fatalf("NewGitRevision(before) error = %v", err)
	}
	after, err := domain.NewGitRevision("89abcdef0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewGitRevision(after) error = %v", err)
	}

	// When
	record, err := domain.NewDeploymentRecord(domain.DeploymentValues{
		PlanID: "plan-1", TargetID: "target-1", Kind: domain.PlanComposeDeploy,
		RevisionBefore: before, RevisionAfter: after, Health: domain.HealthHealthy,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	// Then
	if err != nil {
		t.Fatalf("NewDeploymentRecord() error = %v", err)
	}
	if record.RevisionBefore() != before || record.RevisionAfter() != after {
		t.Fatalf("record revisions = %q/%q", record.RevisionBefore().String(), record.RevisionAfter().String())
	}
}
