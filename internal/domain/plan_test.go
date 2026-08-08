package domain_test

import (
	"coffe/internal/domain"
	"errors"
	"testing"
	"time"
)

func Test_ActionPlan_Approve_returns_approved_copy(t *testing.T) {
	// Given
	createdAt := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	approvedAt := createdAt.Add(time.Minute)
	plan, err := domain.NewActionPlan(domain.PlanDraftValues{
		ID: "plan-1", ProjectID: "project-1", TargetID: "target-1", Kind: domain.PlanComposeDeploy,
		Steps: []domain.StepKind{domain.StepInspectRepository, domain.StepGitPull}, CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("NewActionPlan() error = %v", err)
	}

	// When
	approved, err := plan.Approve(approvedAt)
	// Then
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if approved.Status() != domain.PlanApproved {
		t.Fatalf("approved.Status() = %q, want %q", approved.Status(), domain.PlanApproved)
	}
	if plan.Status() != domain.PlanDraft {
		t.Fatalf("original plan mutated: status = %q", plan.Status())
	}
	if got, ok := approved.ApprovedAt(); !ok || !got.Equal(approvedAt) {
		t.Fatalf("ApprovedAt() = %v, %v; want %v, true", got, ok, approvedAt)
	}
}

func Test_ActionPlan_Start_rejects_draft(t *testing.T) {
	// Given
	plan, err := domain.NewActionPlan(domain.PlanDraftValues{
		ID: "plan-1", ProjectID: "project-1", TargetID: "target-1", Kind: domain.PlanComposeDeploy,
		Steps:     []domain.StepKind{domain.StepInspectRepository},
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewActionPlan() error = %v", err)
	}

	// When
	_, err = plan.Start()

	// Then
	if !errors.Is(err, domain.ErrPlanTransition) {
		t.Fatalf("Start() error = %v, want ErrPlanTransition", err)
	}
}

func Test_ActionPlan_steps_are_copied(t *testing.T) {
	// Given
	steps := []domain.StepKind{domain.StepInspectRepository}
	plan, err := domain.NewActionPlan(domain.PlanDraftValues{
		ID: "plan-1", ProjectID: "project-1", TargetID: "target-1", Kind: domain.PlanComposeDeploy, Steps: steps,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewActionPlan() error = %v", err)
	}

	// When
	steps[0] = domain.StepComposeUp
	got := plan.Steps()
	got[0] = domain.StepComposeBuild

	// Then
	if plan.Steps()[0] != domain.StepInspectRepository {
		t.Fatalf("plan steps mutated: got %q", plan.Steps()[0])
	}
}

func Test_ActionPlan_rejects_unknown_step(t *testing.T) {
	// Given
	steps := []domain.StepKind{domain.StepKind("arbitrary_shell")}

	// When
	_, err := domain.NewActionPlan(domain.PlanDraftValues{
		ID: "plan-1", ProjectID: "project-1", TargetID: "target-1", Kind: domain.PlanComposeDeploy, Steps: steps,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})

	// Then
	if !errors.Is(err, domain.ErrInvalidPlan) {
		t.Fatalf("NewActionPlan() error = %v, want ErrInvalidPlan", err)
	}
}

func Test_ActionPlan_Cancel_transitions_running_plan(t *testing.T) {
	// Given
	createdAt := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	plan := newApprovedPlan(t, createdAt)
	running, err := plan.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// When
	canceled, err := running.Cancel()
	// Then
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if canceled.Status() != domain.PlanCanceled {
		t.Fatalf("Cancel() status = %q, want canceled", canceled.Status())
	}
}

func Test_ActionPlan_Retry_returns_new_draft_linked_to_failed_plan(t *testing.T) {
	// Given
	createdAt := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	plan := newApprovedPlan(t, createdAt)
	running, err := plan.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	failed, err := running.Finish(false)
	if err != nil {
		t.Fatalf("Finish() error = %v", err)
	}

	// When
	retry, err := failed.Retry("plan-2", createdAt.Add(time.Minute))
	// Then
	if err != nil {
		t.Fatalf("Retry() error = %v", err)
	}
	if retry.Status() != domain.PlanDraft || retry.SourcePlanID() != failed.ID() {
		t.Fatalf("Retry() status/source = %q/%q", retry.Status(), retry.SourcePlanID())
	}
}

func newApprovedPlan(t *testing.T, createdAt time.Time) domain.ActionPlan {
	t.Helper()
	plan, err := domain.NewActionPlan(domain.PlanDraftValues{
		ID: "plan-1", ProjectID: "project-1", TargetID: "target-1", Kind: domain.PlanComposeDeploy,
		Steps: []domain.StepKind{domain.StepInspectRepository}, CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("NewActionPlan() error = %v", err)
	}
	approved, err := plan.Approve(createdAt.Add(time.Second))
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	return approved
}
