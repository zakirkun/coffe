package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"errors"
	"testing"
)

func Test_App_ExecutePlan_blocks_swarm_mutation_when_policy_revoked_after_preview(t *testing.T) {
	// Given
	stack, err := domain.NewStackName("payments")
	if err != nil {
		t.Fatalf("NewStackName() error = %v", err)
	}
	fixture := newFixtureWithWorkflow(t, domain.NewSwarmWorkflow(mustComposeFile(t), stack))
	fixture.updatePolicy(t, true, true)
	plan := fixture.preview(t)
	approved, err := fixture.app.ApprovePlan(context.Background(), string(plan.ID()))
	if err != nil {
		t.Fatalf("ApprovePlan() error = %v", err)
	}
	fixture.updatePolicy(t, true, false)

	// When
	_, err = fixture.app.ExecutePlan(context.Background(), string(approved.ID()))

	// Then
	if !errors.Is(err, service.ErrPolicyDenied) {
		t.Fatalf("ExecutePlan() error = %v, want ErrPolicyDenied", err)
	}
	if len(fixture.executor.steps) != 0 {
		t.Fatalf("executed steps = %v, want none", fixture.executor.steps)
	}
}

func Test_App_ExecuteGit_blocks_push_when_policy_revoked_after_preview(t *testing.T) {
	// Given
	fixture := newFixture(t)
	fixture.updatePolicy(t, true, true)
	preview, err := fixture.app.PreviewGit(context.Background(), service.CreateGitOperationInput{
		ProjectID: string(fixture.project.ID()), Kind: string(domain.GitPush), Remote: "origin", Branch: "main",
	})
	if err != nil {
		t.Fatalf("PreviewGit() error = %v", err)
	}
	approved, err := fixture.app.ApproveGit(context.Background(), string(preview.Operation.ID()))
	if err != nil {
		t.Fatalf("ApproveGit() error = %v", err)
	}
	fixture.updatePolicy(t, false, true)

	// When
	finished, err := fixture.app.ExecuteGit(context.Background(), string(approved.ID()))

	// Then
	if !errors.Is(err, service.ErrPolicyDenied) {
		t.Fatalf("ExecuteGit() error = %v, want ErrPolicyDenied", err)
	}
	if fixture.git.executions != 0 {
		t.Fatalf("Git Execute() calls = %d, want 0", fixture.git.executions)
	}
	if finished.Status() != domain.GitFailed {
		t.Fatalf("ExecuteGit() status = %q, want failed", finished.Status())
	}
}

func Test_App_ExecuteGit_persists_failed_state_when_policy_lookup_cancels_request(t *testing.T) {
	// Given
	var policyRepo *cancelingPolicyRepository
	fixture := newFixtureWithRepository(
		t,
		domain.NewWorkflow(mustComposeFile(t), true),
		func(repo *memoryRepository) service.Repository {
			policyRepo = &cancelingPolicyRepository{Repository: repo}
			return policyRepo
		},
	)
	fixture.updatePolicy(t, true, true)
	preview, err := fixture.app.PreviewGit(context.Background(), service.CreateGitOperationInput{
		ProjectID: string(fixture.project.ID()), Kind: string(domain.GitPush), Remote: "origin", Branch: "main",
	})
	if err != nil {
		t.Fatalf("PreviewGit() error = %v", err)
	}
	approved, err := fixture.app.ApproveGit(context.Background(), string(preview.Operation.ID()))
	if err != nil {
		t.Fatalf("ApproveGit() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	policyRepo.cancel = cancel
	policyRepo.cancelNext = true
	fixture.repo.rejectCanceledGitPersistence = true

	// When
	finished, err := fixture.app.ExecuteGit(ctx, string(approved.ID()))

	// Then
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ExecuteGit() error = %v, want context canceled", err)
	}
	if fixture.git.executions != 0 {
		t.Fatalf("Git Execute() calls = %d, want 0", fixture.git.executions)
	}
	operations, listErr := fixture.app.ListGitOperations(context.Background(), 10)
	if listErr != nil {
		t.Fatalf("ListGitOperations() error = %v", listErr)
	}
	if len(operations) != 1 || operations[0].Status() != domain.GitFailed {
		t.Fatalf("stored operations = %#v, want one failed operation", operations)
	}
	if finished.Status() != domain.GitFailed {
		t.Fatalf("ExecuteGit() status = %q, want failed", finished.Status())
	}
}

func Test_App_ExecuteGit_persists_terminal_state_when_output_event_fails(t *testing.T) {
	// Given
	fixture := newFixture(t)
	preview, err := fixture.app.PreviewGit(context.Background(), service.CreateGitOperationInput{
		ProjectID: string(fixture.project.ID()), Kind: string(domain.GitPull), Remote: "origin", Branch: "main",
	})
	if err != nil {
		t.Fatalf("PreviewGit() error = %v", err)
	}
	approved, err := fixture.app.ApproveGit(context.Background(), string(preview.Operation.ID()))
	if err != nil {
		t.Fatalf("ApproveGit() error = %v", err)
	}
	fixture.repo.nextGitEventErr = errors.New("event store unavailable")

	// When
	finished, err := fixture.app.ExecuteGit(context.Background(), string(approved.ID()))

	// Then
	if err == nil {
		t.Fatal("ExecuteGit() error = nil, want event failure")
	}
	if finished.Status() != domain.GitSucceeded {
		t.Fatalf("ExecuteGit() status = %q, want succeeded", finished.Status())
	}
	operations, listErr := fixture.app.ListGitOperations(context.Background(), 10)
	if listErr != nil {
		t.Fatalf("ListGitOperations() error = %v", listErr)
	}
	if len(operations) != 1 || operations[0].Status() != domain.GitSucceeded {
		t.Fatalf("stored operations = %#v, want one succeeded operation", operations)
	}
}

func Test_App_ExecuteGit_persists_terminal_state_when_request_canceled_after_mutation(t *testing.T) {
	// Given
	fixture := newFixture(t)
	preview, err := fixture.app.PreviewGit(context.Background(), service.CreateGitOperationInput{
		ProjectID: string(fixture.project.ID()), Kind: string(domain.GitPull), Remote: "origin", Branch: "main",
	})
	if err != nil {
		t.Fatalf("PreviewGit() error = %v", err)
	}
	approved, err := fixture.app.ApproveGit(context.Background(), string(preview.Operation.ID()))
	if err != nil {
		t.Fatalf("ApproveGit() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	fixture.git.cancel = cancel
	fixture.repo.rejectCanceledGitPersistence = true

	// When
	finished, err := fixture.app.ExecuteGit(ctx, string(approved.ID()))

	// Then
	if err != nil {
		t.Fatalf("ExecuteGit() error = %v", err)
	}
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("execution context error = %v, want canceled", ctx.Err())
	}
	if finished.Status() != domain.GitSucceeded {
		t.Fatalf("ExecuteGit() status = %q, want succeeded", finished.Status())
	}
	operations, listErr := fixture.app.ListGitOperations(context.Background(), 10)
	if listErr != nil {
		t.Fatalf("ListGitOperations() error = %v", listErr)
	}
	if len(operations) != 1 || operations[0].Status() != domain.GitSucceeded {
		t.Fatalf("stored operations = %#v, want one succeeded operation", operations)
	}
}

func Test_App_StartPlan_restores_approved_state_when_start_event_fails(t *testing.T) {
	// Given
	fixture := newFixture(t)
	plan := fixture.preview(t)
	approved, err := fixture.app.ApprovePlan(context.Background(), string(plan.ID()))
	if err != nil {
		t.Fatalf("ApprovePlan() error = %v", err)
	}
	fixture.repo.appendEventKind = domain.EventPlanStarted
	fixture.repo.appendEventErr = errors.New("event store unavailable")

	// When
	_, err = fixture.app.StartPlan(context.Background(), string(approved.ID()))

	// Then
	if err == nil {
		t.Fatal("StartPlan() error = nil, want event failure")
	}
	stored, getErr := fixture.app.GetPlan(context.Background(), string(approved.ID()))
	if getErr != nil {
		t.Fatalf("GetPlan() error = %v", getErr)
	}
	if stored.Status() != domain.PlanApproved {
		t.Fatalf("stored status = %q, want approved", stored.Status())
	}
	if len(fixture.executor.steps) != 0 {
		t.Fatalf("executed steps = %v, want none", fixture.executor.steps)
	}
}

func (f fixture) updatePolicy(t *testing.T, allowGitPush, allowSwarmDeploy bool) {
	t.Helper()
	_, err := f.app.UpdatePolicy(context.Background(), service.UpdatePolicyInput{
		AllowGitPush: allowGitPush, AllowSwarmDeploy: allowSwarmDeploy,
		AgentTools: []string{string(domain.AgentOpenCode)}, MaxConcurrentAgents: 1, AuditRetentionDays: 30,
	})
	if err != nil {
		t.Fatalf("UpdatePolicy() error = %v", err)
	}
}

type cancelingPolicyRepository struct {
	service.Repository
	cancel     context.CancelFunc
	cancelNext bool
}

func (r *cancelingPolicyRepository) GetPolicy(ctx context.Context) (domain.Policy, error) {
	if !r.cancelNext {
		return r.Repository.GetPolicy(ctx)
	}
	r.cancelNext = false
	r.cancel()
	return domain.Policy{}, ctx.Err()
}
