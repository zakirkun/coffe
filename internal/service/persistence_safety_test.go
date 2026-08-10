package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_App_RecoverInterruptedPlans_recovers_running_plan_older_than_100_newer_rows(t *testing.T) {
	// Given
	fixture := newFixture(t)
	now := fixture.repo.policy.UpdatedAt()
	running := mustPlanWithStatus(t, planState{id: "running-plan", status: domain.PlanRunning, createdAt: now.Add(-102 * time.Hour), transitionAt: now})
	secondRunning := mustPlanWithStatus(t, planState{id: "second-running-plan", status: domain.PlanRunning, createdAt: now.Add(-103 * time.Hour), transitionAt: now})
	fixture.repo.plans[running.ID()] = running
	fixture.repo.plans[secondRunning.ID()] = secondRunning
	for index := range 101 {
		plan := mustPlanWithStatus(t, planState{id: domain.PlanID(fmt.Sprintf("finished-plan-%03d", index)), status: domain.PlanSucceeded, createdAt: now.Add(time.Duration(index) * time.Hour), transitionAt: now})
		fixture.repo.plans[plan.ID()] = plan
	}

	// When
	err := fixture.app.RecoverInterruptedPlans(context.Background())

	// Then
	if err != nil {
		t.Fatalf("RecoverInterruptedPlans() error = %v", err)
	}
	stored, err := fixture.repo.GetPlan(context.Background(), running.ID())
	if err != nil {
		t.Fatalf("GetPlan() error = %v", err)
	}
	if stored.Status() != domain.PlanFailed {
		t.Fatalf("recovered plan status = %q, want failed", stored.Status())
	}
	secondStored, err := fixture.repo.GetPlan(context.Background(), secondRunning.ID())
	if err != nil {
		t.Fatalf("GetPlan(second) error = %v", err)
	}
	if secondStored.Status() != domain.PlanFailed {
		t.Fatalf("second recovered plan status = %q, want failed", secondStored.Status())
	}
}

func Test_App_RecoverAgentSessions_recovers_running_agent_older_than_100_newer_rows(t *testing.T) {
	// Given
	fixture := newFixture(t)
	now := fixture.repo.policy.UpdatedAt()
	running := mustAgentWithStatus(t, agentState{id: "running-agent", status: domain.AgentRunning, createdAt: now.Add(-102 * time.Hour), transitionAt: now})
	secondRunning := mustAgentWithStatus(t, agentState{id: "second-running-agent", status: domain.AgentRunning, createdAt: now.Add(-103 * time.Hour), transitionAt: now})
	fixture.repo.agentSessions[running.ID()] = running
	fixture.repo.agentSessions[secondRunning.ID()] = secondRunning
	for index := range 101 {
		session := mustAgentWithStatus(t, agentState{id: domain.AgentSessionID(fmt.Sprintf("finished-agent-%03d", index)), status: domain.AgentSucceeded, createdAt: now.Add(time.Duration(index) * time.Hour), transitionAt: now})
		fixture.repo.agentSessions[session.ID()] = session
	}

	// When
	err := fixture.app.RecoverAgentSessions(context.Background())

	// Then
	if err != nil {
		t.Fatalf("RecoverAgentSessions() error = %v", err)
	}
	stored, err := fixture.repo.GetAgentSession(context.Background(), running.ID())
	if err != nil {
		t.Fatalf("GetAgentSession() error = %v", err)
	}
	if stored.Status() != domain.AgentFailed {
		t.Fatalf("recovered agent status = %q, want failed", stored.Status())
	}
	secondStored, err := fixture.repo.GetAgentSession(context.Background(), secondRunning.ID())
	if err != nil {
		t.Fatalf("GetAgentSession(second) error = %v", err)
	}
	if secondStored.Status() != domain.AgentFailed {
		t.Fatalf("second recovered agent status = %q, want failed", secondStored.Status())
	}
}

func Test_App_RecordAudit_prunes_expired_entries_during_activity(t *testing.T) {
	// Given
	fixture := newFixture(t)
	now := fixture.repo.policy.UpdatedAt()
	expired, err := domain.NewAuditEntry(domain.AuditValues{
		ID: "expired", Category: domain.AuditDatabase, Action: "query", ResourceID: "db-1",
		Outcome: domain.AuditSuccess, Message: "old query", OccurredAt: now.Add(-366 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("NewAuditEntry() error = %v", err)
	}
	fixture.repo.audit = append(fixture.repo.audit, expired)

	// When
	err = fixture.app.RecordAudit(context.Background(), service.AuditRecordInput{
		Category: domain.AuditDatabase, Action: "query", ResourceID: "db-1",
		Outcome: domain.AuditSuccess, Message: "new query",
	})

	// Then
	if err != nil {
		t.Fatalf("RecordAudit() error = %v", err)
	}
	if len(fixture.repo.audit) != 1 || fixture.repo.audit[0].ID() == expired.ID() {
		t.Fatalf("audit entries = %v, want only current activity", fixture.repo.audit)
	}
	wantCutoff := now.Add(-365 * 24 * time.Hour)
	if !fixture.repo.auditDeleteAt.Equal(wantCutoff) {
		t.Fatalf("audit cutoff = %s, want %s", fixture.repo.auditDeleteAt, wantCutoff)
	}
}

type planState struct {
	id           domain.PlanID
	status       domain.PlanStatus
	createdAt    time.Time
	transitionAt time.Time
}

func mustPlanWithStatus(t *testing.T, state planState) domain.ActionPlan {
	t.Helper()
	plan, err := domain.RestoreActionPlan(domain.PlanValues{
		ID: state.id, ProjectID: "project-1", TargetID: "target-1", Kind: domain.PlanComposeDeploy,
		Status: state.status, Steps: []domain.StepKind{domain.StepInspectRepository},
		CreatedAt: state.createdAt, ApprovedAt: state.transitionAt,
	})
	if err != nil {
		t.Fatalf("RestoreActionPlan() error = %v", err)
	}
	return plan
}

type agentState struct {
	id           domain.AgentSessionID
	status       domain.AgentStatus
	createdAt    time.Time
	transitionAt time.Time
}

func mustAgentWithStatus(t *testing.T, state agentState) domain.AgentSession {
	t.Helper()
	values := domain.AgentSessionValues{
		ID: state.id, ProjectID: "project-1", Tool: domain.AgentOpenCode,
		Status: state.status, CreatedAt: state.createdAt, StartedAt: state.transitionAt,
	}
	if state.status != domain.AgentRunning {
		values.FinishedAt = state.transitionAt
	}
	session, err := domain.NewAgentSession(values)
	if err != nil {
		t.Fatalf("NewAgentSession() error = %v", err)
	}
	return session
}
