package worker_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"coffe/internal/worker"
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func Test_Client_Ping_authenticates_TLS_worker(t *testing.T) {
	// Given
	token := "0123456789abcdefghijklmn"
	handler, err := worker.NewServer(worker.ServerConfig{Token: token, Version: "v1.0.0"})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	registered := registeredWorker(t, server.URL)
	client := worker.NewClient(worker.ClientConfig{HTTPClient: server.Client()})

	// When
	heartbeat, err := client.Ping(context.Background(), registered, token)
	// Then
	if err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if heartbeat.Version != "v1.0.0" {
		t.Fatalf("heartbeat version = %q", heartbeat.Version)
	}
}

func Test_Client_executes_allowlisted_revision_task(t *testing.T) {
	// Given
	token := "0123456789abcdefghijklmn"
	handler, err := worker.NewServer(worker.ServerConfig{Token: token, Version: "v1.0.0"})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	execution := workerExecution(t)
	client := worker.NewClient(worker.ClientConfig{HTTPClient: server.Client()})

	// When
	result, err := client.Execute(context.Background(), registeredWorker(t, server.URL), token, execution)

	// Then
	if runtime.GOOS == "windows" {
		if err == nil {
			t.Fatal("Execute() error = nil on unsupported Windows worker")
		}
		return
	}
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Revision.IsZero() {
		t.Fatal("Execute() revision is empty")
	}
}

func workerExecution(t *testing.T) service.ExecutionStep {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init", "-b", "main")
	runGit(t, root, "config", "user.email", "coffe@example.invalid")
	runGit(t, root, "config", "user.name", "Coffe Worker Test")
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("worker\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	runGit(t, root, "add", "app.txt")
	runGit(t, root, "commit", "-m", "initial")
	target := taskTarget(t, root)
	composeFile, err := domain.NewComposeFile("compose.yml")
	if err != nil {
		t.Fatalf("NewComposeFile() error = %v", err)
	}
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	plan, err := domain.RestoreActionPlan(domain.PlanValues{
		ID: "plan-1", ProjectID: "project-1", TargetID: target.ID(), Kind: domain.PlanComposeDeploy,
		Status: domain.PlanRunning, Steps: []domain.StepKind{domain.StepCaptureRevision},
		CreatedAt: now, ApprovedAt: now,
	})
	if err != nil {
		t.Fatalf("RestoreActionPlan() error = %v", err)
	}
	return service.ExecutionStep{
		Plan: plan, Target: target, Workflow: domain.NewWorkflow(composeFile, false),
		Step: domain.StepCaptureRevision,
	}
}

func taskTarget(t *testing.T, root string) domain.Target {
	t.Helper()
	name, nameErr := domain.NewName("worker-target")
	host, hostErr := domain.NewHost("localhost")
	port, portErr := domain.NewPort(22)
	user, userErr := domain.NewSSHUser("worker")
	remote := filepath.ToSlash(root)
	if runtime.GOOS == "windows" {
		remote = "/tmp/not-used"
	}
	remotePath, pathErr := domain.NewRemotePath(remote)
	branch, branchErr := domain.NewBranch("main")
	if err := errors.Join(nameErr, hostErr, portErr, userErr, pathErr, branchErr); err != nil {
		t.Fatalf("target values error = %v", err)
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

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-C", root}, args...)
	if output, err := exec.Command("git", commandArgs...).CombinedOutput(); err != nil { //nolint:gosec // test invokes fixed git with test-owned arguments.
		t.Fatalf("git %v error = %v: %s", args, err, output)
	}
}

func registeredWorker(t *testing.T, baseURL string) domain.Worker {
	t.Helper()
	name, err := domain.NewName("test-worker")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}
	registered, err := domain.NewWorker(domain.WorkerValues{
		ID: "worker-1", TargetID: "target-1", Name: name,
		BaseURL: baseURL, SecretName: "worker-token", Enabled: true,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewWorker() error = %v", err)
	}
	return registered
}
