package service_test

import (
	"coffe/internal/domain"
	"context"
	"testing"
	"time"
)

func Test_App_RetryPlan_creates_new_draft_after_failure(t *testing.T) {
	// Given
	fixture := newFixture(t)
	fixture.executor.failAt = domain.StepComposeBuild
	plan := fixture.preview(t)
	approved, err := fixture.app.ApprovePlan(context.Background(), string(plan.ID()))
	if err != nil {
		t.Fatalf("ApprovePlan() error = %v", err)
	}
	if _, err := fixture.app.ExecutePlan(context.Background(), string(approved.ID())); err == nil {
		t.Fatal("ExecutePlan() error = nil, want failure")
	}

	// When
	retry, err := fixture.app.RetryPlan(context.Background(), string(approved.ID()))
	// Then
	if err != nil {
		t.Fatalf("RetryPlan() error = %v", err)
	}
	if retry.Status() != domain.PlanDraft || retry.SourcePlanID() != approved.ID() {
		t.Fatalf("retry status/source = %q/%q", retry.Status(), retry.SourcePlanID())
	}
}

func Test_App_PreviewRollback_uses_previous_revision(t *testing.T) {
	// Given
	fixture := newFixture(t)
	plan := fixture.preview(t)
	approved, err := fixture.app.ApprovePlan(context.Background(), string(plan.ID()))
	if err != nil {
		t.Fatalf("ApprovePlan() error = %v", err)
	}
	if _, err := fixture.app.ExecutePlan(context.Background(), string(approved.ID())); err != nil {
		t.Fatalf("ExecutePlan() error = %v", err)
	}
	record, err := fixture.repo.LatestDeployment(context.Background(), fixture.target.ID())
	if err != nil {
		t.Fatalf("LatestDeployment() error = %v", err)
	}

	// When
	rollback, err := fixture.app.PreviewRollback(context.Background(), string(fixture.target.ID()))
	// Then
	if err != nil {
		t.Fatalf("PreviewRollback() error = %v", err)
	}
	revision, ok := rollback.Revision()
	if rollback.Kind() != domain.PlanComposeRollback || !ok || revision != record.RevisionBefore() {
		t.Fatalf("rollback kind/revision = %q/%q", rollback.Kind(), revision.String())
	}
}

func Test_App_CancelPlan_stops_running_execution(t *testing.T) {
	// Given
	fixture := newFixture(t)
	fixture.executor.blockAt = domain.StepInspectRepository
	plan := fixture.preview(t)
	approved, err := fixture.app.ApprovePlan(context.Background(), string(plan.ID()))
	if err != nil {
		t.Fatalf("ApprovePlan() error = %v", err)
	}
	if _, err := fixture.app.StartPlan(context.Background(), string(approved.ID())); err != nil {
		t.Fatalf("StartPlan() error = %v", err)
	}
	select {
	case <-fixture.executor.started:
	case <-time.After(time.Second):
		t.Fatal("executor did not start")
	}

	// When
	if err := fixture.app.CancelPlan(context.Background(), string(approved.ID())); err != nil {
		t.Fatalf("CancelPlan() error = %v", err)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := fixture.app.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	stored, err := fixture.repo.GetPlan(context.Background(), approved.ID())
	// Then
	if err != nil {
		t.Fatalf("GetPlan() error = %v", err)
	}
	if stored.Status() != domain.PlanCanceled {
		t.Fatalf("stored status = %q, want canceled", stored.Status())
	}
}
