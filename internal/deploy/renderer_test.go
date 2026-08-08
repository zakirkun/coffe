package deploy_test

import (
	"coffe/internal/deploy"
	"coffe/internal/domain"
	"coffe/internal/service"
	"errors"
	"testing"
	"time"
)

func Test_Render_returns_fixed_command_for_each_step(t *testing.T) {
	// Given
	target := newTarget(t, "/srv/payment's api")
	workflow := domain.NewWorkflow(newComposeFile(t, "deploy/compose.yml"), true)
	plan := newPlan(t, domain.PlanComposeDeploy, domain.GitRevision{})
	tests := []struct {
		step domain.StepKind
		want string
	}{
		{
			step: domain.StepInspectRepository,
			want: "cd -- '/srv/payment'\"'\"'s api' && test \"$(git branch --show-current)\" = 'main' && test -z \"$(git status --porcelain)\" && git rev-parse HEAD",
		},
		{
			step: domain.StepGitPull,
			want: "cd -- '/srv/payment'\"'\"'s api' && git pull --ff-only origin 'main'",
		},
		{
			step: domain.StepCaptureRevision,
			want: "cd -- '/srv/payment'\"'\"'s api' && git rev-parse HEAD",
		},
		{
			step: domain.StepComposeConfig,
			want: "cd -- '/srv/payment'\"'\"'s api' && docker compose -f 'deploy/compose.yml' config --quiet",
		},
		{
			step: domain.StepComposeBuild,
			want: "cd -- '/srv/payment'\"'\"'s api' && docker compose -f 'deploy/compose.yml' build",
		},
		{
			step: domain.StepComposeUp,
			want: "cd -- '/srv/payment'\"'\"'s api' && docker compose -f 'deploy/compose.yml' up -d",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.step), func(t *testing.T) {
			// When
			got, err := deploy.Render(service.ExecutionStep{
				Plan: plan, Target: target, Workflow: workflow, Step: tt.step,
			})
			// Then
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_Render_rejects_unknown_step(t *testing.T) {
	// Given
	target := newTarget(t, "/srv/payments")
	workflow := domain.NewWorkflow(newComposeFile(t, "compose.yml"), true)
	plan := newPlan(t, domain.PlanComposeDeploy, domain.GitRevision{})

	// When
	_, err := deploy.Render(service.ExecutionStep{
		Plan: plan, Target: target, Workflow: workflow, Step: domain.StepKind("shell"),
	})

	// Then
	if !errors.Is(err, deploy.ErrUnsupportedStep) {
		t.Fatalf("Render() error = %v, want ErrUnsupportedStep", err)
	}
}

func Test_SSHArguments_enforces_noninteractive_host_verification(t *testing.T) {
	// Given
	target := newTarget(t, "/srv/payments")

	// When
	got := deploy.SSHArguments(target, "true")

	// Then
	want := []string{
		"-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=yes", "-o", "ConnectTimeout=10",
		"-p", "22", "deploy@server.example.com", "true",
	}
	if len(got) != len(want) {
		t.Fatalf("SSHArguments() len = %d, want %d: %v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("SSHArguments()[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func newTarget(t *testing.T, remote string) domain.Target {
	t.Helper()
	name, nameErr := domain.NewName("office")
	host, hostErr := domain.NewHost("server.example.com")
	port, portErr := domain.NewPort(22)
	user, userErr := domain.NewSSHUser("deploy")
	remotePath, err := domain.NewRemotePath(remote)
	branch, branchErr := domain.NewBranch("main")
	if err := errors.Join(nameErr, hostErr, portErr, userErr, err, branchErr); err != nil {
		t.Fatalf("target value error = %v", err)
	}
	target, err := domain.NewTarget(domain.TargetValues{
		ID: "target-1", ProjectID: "project-1", Name: name, Host: host, Port: port,
		User: user, RemotePath: remotePath, Branch: branch,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewTarget() error = %v", err)
	}
	return target
}

func newComposeFile(t *testing.T, raw string) domain.ComposeFile {
	t.Helper()
	file, err := domain.NewComposeFile(raw)
	if err != nil {
		t.Fatalf("NewComposeFile() error = %v", err)
	}
	return file
}

func newPlan(t *testing.T, kind domain.PlanKind, revision domain.GitRevision) domain.ActionPlan {
	t.Helper()
	step := domain.StepInspectRepository
	if kind == domain.PlanComposeRollback {
		step = domain.StepGitReset
	}
	plan, err := domain.NewActionPlan(domain.PlanDraftValues{
		ID: "plan-1", ProjectID: "project-1", TargetID: "target-1",
		Kind: kind, Revision: revision, Steps: []domain.StepKind{step},
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewActionPlan() error = %v", err)
	}
	return plan
}
