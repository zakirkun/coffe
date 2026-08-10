package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

type fixture struct {
	app      *service.App
	repo     *memoryRepository
	executor *recordingExecutor
	git      *fixedGitTool
	project  domain.Project
	target   domain.Target
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	return newFixtureWithWorkflow(t, domain.NewWorkflow(mustComposeFile(t), true))
}

func newFixtureWithWorkflow(t *testing.T, workflow domain.Workflow) fixture {
	t.Helper()
	return newFixtureWithRepository(t, workflow, func(repo *memoryRepository) service.Repository { return repo })
}

func newFixtureWithRepository(
	t *testing.T,
	workflow domain.Workflow,
	wrap func(*memoryRepository) service.Repository,
) fixture {
	t.Helper()
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	project := mustProject(t, now)
	target := mustTarget(t, project.ID(), now)
	repo := newMemoryRepository()
	if err := repo.CreateProject(context.Background(), project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if err := repo.CreateTarget(context.Background(), target); err != nil {
		t.Fatalf("CreateTarget() error = %v", err)
	}
	executor := &recordingExecutor{started: make(chan struct{}), finished: make(chan struct{})}
	git := &fixedGitTool{}
	app := service.New(service.Dependencies{
		Logger: slog.New(slog.DiscardHandler), Repository: wrap(repo),
		Workflows: fixedWorkflowLoader{workflow: workflow},
		Planner:   fixedPlanner{intent: domain.IntentDeploy}, Executor: executor,
		Renderer: fixedRenderer{}, Swarm: fixedSwarmReader{}, Agents: successfulAgentRunner{},
		Databases: fixedDatabaseReader{}, Secrets: fixedSecretReader{},
		Git:      git,
		AIModels: fixedAIModelCatalog{},
		Workers:  fixedWorkerProbe{},
		Plugins:  fixedPluginRunner{},
		Clock:    fixedClock{now: now}, IDs: &sequenceIDs{},
	})
	if err := app.EnsureDefaultPolicy(context.Background()); err != nil {
		t.Fatalf("EnsureDefaultPolicy() error = %v", err)
	}
	return fixture{app: app, repo: repo, executor: executor, git: git, project: project, target: target}
}

func (f fixture) preview(t *testing.T) domain.ActionPlan {
	t.Helper()
	plan, err := f.app.PreviewDeploy(context.Background(), service.PreviewInput{
		ProjectID: string(f.project.ID()), TargetID: string(f.target.ID()), Prompt: "deploy",
	})
	if err != nil {
		t.Fatalf("PreviewDeploy() error = %v", err)
	}
	return plan
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type sequenceIDs struct{ next int }

func (s *sequenceIDs) New() (string, error) {
	s.next++
	return fmt.Sprintf("id-%d", s.next), nil
}

func mustProject(t *testing.T, now time.Time) domain.Project {
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

func mustTarget(t *testing.T, projectID domain.ProjectID, now time.Time) domain.Target {
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

func mustComposeFile(t *testing.T) domain.ComposeFile {
	t.Helper()
	file, err := domain.NewComposeFile("docker-compose.yml")
	if err != nil {
		t.Fatalf("NewComposeFile() error = %v", err)
	}
	return file
}
