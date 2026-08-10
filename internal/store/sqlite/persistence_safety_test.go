package sqlite_test

import (
	"coffe/internal/domain"
	coffesqlite "coffe/internal/store/sqlite"
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func Test_Store_ListRunningPlans_returns_old_running_plan_with_101_newer_rows(t *testing.T) {
	// Given
	ctx := context.Background()
	store, err := coffesqlite.Open(ctx, filepath.Join(t.TempDir(), "coffe.db"))
	requireNoError(t, "Open", err)
	t.Cleanup(func() { requireNoError(t, "Close", store.Close()) })
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.UTC)
	project := newProject(t, now)
	target := newTarget(t, project.ID(), now)
	requireNoError(t, "CreateProject", store.CreateProject(ctx, project))
	requireNoError(t, "CreateTarget", store.CreateTarget(ctx, target))
	running := sqlitePlanWithStatus(t, sqlitePlanState{id: "running-plan", status: domain.PlanRunning, createdAt: now.Add(-102 * time.Hour), transitionAt: now})
	secondRunning := sqlitePlanWithStatus(t, sqlitePlanState{id: "second-running-plan", status: domain.PlanRunning, createdAt: now.Add(-103 * time.Hour), transitionAt: now})
	requireNoError(t, "SavePlan(running)", store.SavePlan(ctx, running))
	requireNoError(t, "SavePlan(second running)", store.SavePlan(ctx, secondRunning))
	for index := range 101 {
		plan := sqlitePlanWithStatus(t, sqlitePlanState{id: domain.PlanID(fmt.Sprintf("finished-plan-%03d", index)), status: domain.PlanSucceeded, createdAt: now.Add(time.Duration(index) * time.Hour), transitionAt: now})
		requireNoError(t, "SavePlan(finished)", store.SavePlan(ctx, plan))
	}

	// When
	plans, err := store.ListRunningPlans(ctx)

	// Then
	requireNoError(t, "ListRunningPlans", err)
	if len(plans) != 2 || plans[0].ID() != running.ID() || plans[1].ID() != secondRunning.ID() {
		t.Fatalf("ListRunningPlans() = %v, want %s and %s", plans, running.ID(), secondRunning.ID())
	}
}

func Test_Store_ListRunningAgentSessions_returns_old_running_agent_with_101_newer_rows(t *testing.T) {
	// Given
	ctx := context.Background()
	store, err := coffesqlite.Open(ctx, filepath.Join(t.TempDir(), "coffe.db"))
	requireNoError(t, "Open", err)
	t.Cleanup(func() { requireNoError(t, "Close", store.Close()) })
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.UTC)
	project := newProject(t, now)
	requireNoError(t, "CreateProject", store.CreateProject(ctx, project))
	running := sqliteAgentWithStatus(t, sqliteAgentState{id: "running-agent", status: domain.AgentRunning, createdAt: now.Add(-102 * time.Hour), transitionAt: now})
	secondRunning := sqliteAgentWithStatus(t, sqliteAgentState{id: "second-running-agent", status: domain.AgentRunning, createdAt: now.Add(-103 * time.Hour), transitionAt: now})
	requireNoError(t, "SaveAgentSession(running)", store.SaveAgentSession(ctx, running))
	requireNoError(t, "SaveAgentSession(second running)", store.SaveAgentSession(ctx, secondRunning))
	for index := range 101 {
		session := sqliteAgentWithStatus(t, sqliteAgentState{id: domain.AgentSessionID(fmt.Sprintf("finished-agent-%03d", index)), status: domain.AgentSucceeded, createdAt: now.Add(time.Duration(index) * time.Hour), transitionAt: now})
		requireNoError(t, "SaveAgentSession(finished)", store.SaveAgentSession(ctx, session))
	}

	// When
	sessions, err := store.ListRunningAgentSessions(ctx)

	// Then
	requireNoError(t, "ListRunningAgentSessions", err)
	if len(sessions) != 2 || sessions[0].ID() != running.ID() || sessions[1].ID() != secondRunning.ID() {
		t.Fatalf("ListRunningAgentSessions() = %v, want %s and %s", sessions, running.ID(), secondRunning.ID())
	}
}

type sqlitePlanState struct {
	id           domain.PlanID
	status       domain.PlanStatus
	createdAt    time.Time
	transitionAt time.Time
}

func sqlitePlanWithStatus(t *testing.T, state sqlitePlanState) domain.ActionPlan {
	t.Helper()
	plan, err := domain.RestoreActionPlan(domain.PlanValues{
		ID: state.id, ProjectID: "project-1", TargetID: "target-1", Kind: domain.PlanComposeDeploy,
		Status: state.status, Steps: []domain.StepKind{domain.StepInspectRepository},
		CreatedAt: state.createdAt, ApprovedAt: state.transitionAt,
	})
	requireNoError(t, "RestoreActionPlan", err)
	return plan
}

type sqliteAgentState struct {
	id           domain.AgentSessionID
	status       domain.AgentStatus
	createdAt    time.Time
	transitionAt time.Time
}

func sqliteAgentWithStatus(t *testing.T, state sqliteAgentState) domain.AgentSession {
	t.Helper()
	values := domain.AgentSessionValues{
		ID: state.id, ProjectID: "project-1", Tool: domain.AgentOpenCode,
		Status: state.status, CreatedAt: state.createdAt, StartedAt: state.transitionAt,
	}
	if state.status != domain.AgentRunning {
		values.FinishedAt = state.transitionAt
	}
	session, err := domain.NewAgentSession(values)
	requireNoError(t, "NewAgentSession", err)
	return session
}
