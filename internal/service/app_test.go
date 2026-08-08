package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"slices"
	"strings"
	"testing"
)

func Test_App_PreviewDeploy_creates_draft_from_allowlisted_steps(t *testing.T) {
	// Given
	fixture := newFixture(t)

	// When
	plan, err := fixture.app.PreviewDeploy(context.Background(), service.PreviewInput{
		ProjectID: string(fixture.project.ID()),
		TargetID:  string(fixture.target.ID()),
		Prompt:    "deploy payments to office",
	})
	// Then
	if err != nil {
		t.Fatalf("PreviewDeploy() error = %v", err)
	}
	want := []domain.StepKind{
		domain.StepInspectRepository,
		domain.StepGitPull,
		domain.StepCaptureRevision,
		domain.StepComposeConfig,
		domain.StepComposeBuild,
		domain.StepComposeUp,
		domain.StepComposeHealth,
	}
	if !slices.Equal(plan.Steps(), want) {
		t.Fatalf("PreviewDeploy() steps = %v, want %v", plan.Steps(), want)
	}
	if plan.Status() != domain.PlanDraft {
		t.Fatalf("PreviewDeploy() status = %q, want draft", plan.Status())
	}
}

func Test_App_ApprovePlan_requires_explicit_transition(t *testing.T) {
	// Given
	fixture := newFixture(t)
	plan := fixture.preview(t)

	// When
	approved, err := fixture.app.ApprovePlan(context.Background(), string(plan.ID()))
	// Then
	if err != nil {
		t.Fatalf("ApprovePlan() error = %v", err)
	}
	if approved.Status() != domain.PlanApproved {
		t.Fatalf("ApprovePlan() status = %q, want approved", approved.Status())
	}
}

func Test_App_PlanPreview_returns_rendered_allowlisted_commands(t *testing.T) {
	// Given
	fixture := newFixture(t)
	plan := fixture.preview(t)

	// When
	preview, err := fixture.app.PlanPreview(context.Background(), string(plan.ID()))
	// Then
	if err != nil {
		t.Fatalf("PlanPreview() error = %v", err)
	}
	if len(preview.Commands) != len(plan.Steps()) {
		t.Fatalf("PlanPreview() commands = %d, want %d", len(preview.Commands), len(plan.Steps()))
	}
	if preview.Commands[0] != "rendered:inspect_repository" {
		t.Fatalf("PlanPreview() first command = %q", preview.Commands[0])
	}
}

func Test_App_ExecutePlan_runs_each_step_after_approval(t *testing.T) {
	// Given
	fixture := newFixture(t)
	plan := fixture.preview(t)
	approved, err := fixture.app.ApprovePlan(context.Background(), string(plan.ID()))
	if err != nil {
		t.Fatalf("ApprovePlan() error = %v", err)
	}

	// When
	result, err := fixture.app.ExecutePlan(context.Background(), string(approved.ID()))
	// Then
	if err != nil {
		t.Fatalf("ExecutePlan() error = %v", err)
	}
	if result.Status() != domain.PlanSucceeded {
		t.Fatalf("ExecutePlan() status = %q, want succeeded", result.Status())
	}
	if !slices.Equal(fixture.executor.steps, approved.Steps()) {
		t.Fatalf("executed steps = %v, want %v", fixture.executor.steps, approved.Steps())
	}
	record, recordErr := fixture.repo.LatestDeployment(context.Background(), fixture.target.ID())
	if recordErr != nil {
		t.Fatalf("LatestDeployment() error = %v", recordErr)
	}
	if record.PlanID() != approved.ID() {
		t.Fatalf("deployment plan = %q, want %q", record.PlanID(), approved.ID())
	}
	events, eventsErr := fixture.repo.ListEvents(context.Background(), approved.ID())
	if eventsErr != nil {
		t.Fatalf("ListEvents() error = %v", eventsErr)
	}
	for _, event := range events {
		if strings.Contains(event.Message(), "secret") {
			t.Fatalf("event leaked secret: %q", event.Message())
		}
	}
}

func Test_App_ExecutePlan_records_failed_status_when_step_fails(t *testing.T) {
	// Given
	fixture := newFixture(t)
	fixture.executor.failAt = domain.StepComposeBuild
	plan := fixture.preview(t)
	approved, err := fixture.app.ApprovePlan(context.Background(), string(plan.ID()))
	if err != nil {
		t.Fatalf("ApprovePlan() error = %v", err)
	}

	// When
	_, err = fixture.app.ExecutePlan(context.Background(), string(approved.ID()))

	// Then
	if err == nil {
		t.Fatal("ExecutePlan() error = nil, want failure")
	}
	stored, getErr := fixture.repo.GetPlan(context.Background(), approved.ID())
	if getErr != nil {
		t.Fatalf("GetPlan() error = %v", getErr)
	}
	if stored.Status() != domain.PlanFailed {
		t.Fatalf("stored status = %q, want failed", stored.Status())
	}
}
