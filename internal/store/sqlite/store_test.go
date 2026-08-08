package sqlite_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	coffesqlite "coffe/internal/store/sqlite"
)

func Test_Store_roundtrips_project_target_plan_and_event(t *testing.T) {
	// Given
	ctx := context.Background()
	store, err := coffesqlite.Open(ctx, filepath.Join(t.TempDir(), "coffe.db"))
	requireNoError(t, "Open", err)
	t.Cleanup(func() {
		if closeErr := store.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	project := newProject(t, now)
	target := newTarget(t, project.ID(), now)
	plan, err := domain.NewActionPlan(domain.PlanDraftValues{
		ID: "plan-1", ProjectID: project.ID(), TargetID: target.ID(), Kind: domain.PlanComposeDeploy,
		Steps: []domain.StepKind{domain.StepInspectRepository}, CreatedAt: now,
	})
	requireNoError(t, "NewActionPlan", err)
	event, err := domain.NewEvent(domain.EventValues{
		ID: "event-1", PlanID: plan.ID(), Kind: domain.EventPlanCreated,
		Message: "preview created", OccurredAt: now,
	})
	requireNoError(t, "NewEvent", err)
	before, err := domain.NewGitRevision("0123456789abcdef0123456789abcdef01234567")
	requireNoError(t, "NewGitRevision(before)", err)
	after, err := domain.NewGitRevision("89abcdef0123456789abcdef0123456789abcdef")
	requireNoError(t, "NewGitRevision(after)", err)
	deployment, err := domain.NewDeploymentRecord(domain.DeploymentValues{
		PlanID: plan.ID(), TargetID: target.ID(), Kind: domain.PlanComposeDeploy,
		RevisionBefore: before, RevisionAfter: after, Health: domain.HealthHealthy, CreatedAt: now,
	})
	requireNoError(t, "NewDeploymentRecord", err)

	// When
	requireNoError(t, "CreateProject", store.CreateProject(ctx, project))
	requireNoError(t, "CreateTarget", store.CreateTarget(ctx, target))
	requireNoError(t, "SavePlan", store.SavePlan(ctx, plan))
	requireNoError(t, "AppendEvent", store.AppendEvent(ctx, event))
	requireNoError(t, "SaveDeployment", store.SaveDeployment(ctx, deployment))
	gotPlan, err := store.GetPlan(ctx, plan.ID())
	requireNoError(t, "GetPlan", err)
	events, err := store.ListEvents(ctx, plan.ID())
	requireNoError(t, "ListEvents", err)
	latest, err := store.LatestDeployment(ctx, target.ID())
	requireNoError(t, "LatestDeployment", err)

	// Then
	if gotPlan.Status() != domain.PlanDraft {
		t.Fatalf("GetPlan() status = %q, want draft", gotPlan.Status())
	}
	if len(events) != 1 || events[0].Message() != "preview created" {
		t.Fatalf("ListEvents() = %v, want one preview event", events)
	}
	if latest.RevisionBefore() != before || latest.RevisionAfter() != after {
		t.Fatalf("LatestDeployment() revisions = %q/%q", latest.RevisionBefore().String(), latest.RevisionAfter().String())
	}
}

func requireNoError(t *testing.T, operation string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s() error = %v", operation, err)
	}
}

func Test_Store_GetProject_returns_not_found_for_missing_id(t *testing.T) {
	// Given
	ctx := context.Background()
	store, err := coffesqlite.Open(ctx, filepath.Join(t.TempDir(), "coffe.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := store.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})

	// When
	_, err = store.GetProject(ctx, "missing")

	// Then
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("GetProject() error = %v, want ErrNotFound", err)
	}
}

func newProject(t *testing.T, now time.Time) domain.Project {
	t.Helper()
	name, err := domain.NewName("payments")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}
	project, err := domain.NewProject(domain.ProjectValues{
		ID: "project-1", Name: name, RootPath: t.TempDir(), CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("NewProject() error = %v", err)
	}
	return project
}

func newTarget(t *testing.T, projectID domain.ProjectID, now time.Time) domain.Target {
	t.Helper()
	name, nameErr := domain.NewName("office")
	host, hostErr := domain.NewHost("deploy.example.com")
	port, portErr := domain.NewPort(22)
	user, userErr := domain.NewSSHUser("deploy")
	remotePath, pathErr := domain.NewRemotePath("/srv/payments")
	branch, branchErr := domain.NewBranch("main")
	if err := errors.Join(nameErr, hostErr, portErr, userErr, pathErr, branchErr); err != nil {
		t.Fatalf("target value error = %v", err)
	}
	target, err := domain.NewTarget(domain.TargetValues{
		ID: "target-1", ProjectID: projectID, Name: name, Host: host, Port: port,
		User: user, RemotePath: remotePath, Branch: branch, CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("NewTarget() error = %v", err)
	}
	return target
}
