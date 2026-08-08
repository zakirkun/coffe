package httpapi_test

import (
	"coffe/internal/config"
	"coffe/internal/dbtool"
	"coffe/internal/deploy"
	"coffe/internal/domain"
	"coffe/internal/gittool"
	"coffe/internal/pluginexec"
	"coffe/internal/service"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	coffesqlite "coffe/internal/store/sqlite"
)

func newApplication(t *testing.T) (*service.App, func(), string) {
	t.Helper()
	root := t.TempDir()
	workflow := []byte("version: 1\ndeploy:\n  compose_file: docker-compose.yml\n  build: true\n")
	if err := os.WriteFile(filepath.Join(root, "coffe.yaml"), workflow, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	store, err := coffesqlite.Open(context.Background(), filepath.Join(t.TempDir(), "coffe.db"))
	if err != nil {
		t.Fatalf("sqlite.Open() error = %v", err)
	}
	app := service.New(service.Dependencies{
		Logger:     slog.New(slog.DiscardHandler),
		Repository: store,
		Workflows:  config.WorkflowLoader{},
		Planner:    deployPlanner{},
		Executor:   successfulExecutor{},
		Renderer:   deploy.CommandRenderer{},
		Swarm:      deploy.NewSwarmInspector(deploy.BinaryRunner{}),
		Agents:     successfulAgentRunner{},
		Databases:  dbtool.NewReader(time.Second, 100),
		Secrets:    fixedSecretReader{},
		Git:        gittool.Tool{},
		AIModels:   fixedAIModelCatalog{},
		Workers:    fixedWorkerProbe{},
		Plugins:    pluginexec.NewRunner(time.Second),
		Clock:      testClock{},
		IDs:        &testIDs{},
	})
	if err := app.EnsureDefaultPolicy(context.Background()); err != nil {
		t.Fatalf("EnsureDefaultPolicy() error = %v", err)
	}
	closeApp := func() {
		if err := store.Close(); err != nil {
			t.Errorf("Store.Close() error = %v", err)
		}
	}
	return app, closeApp, root
}

type deployPlanner struct{}

func (deployPlanner) Classify(context.Context, string) (domain.Intent, error) {
	return domain.IntentDeploy, nil
}

type successfulExecutor struct{}

func (successfulExecutor) Execute(
	_ context.Context,
	execution service.ExecutionStep,
) (service.StepResult, error) {
	var rawRevision string
	switch execution.Step {
	case domain.StepInspectRepository:
		rawRevision = "0123456789abcdef0123456789abcdef01234567"
	case domain.StepCaptureRevision:
		rawRevision = "89abcdef0123456789abcdef0123456789abcdef"
	case domain.StepGitPull, domain.StepComposeConfig, domain.StepComposeBuild,
		domain.StepComposeUp, domain.StepComposeHealth, domain.StepGitReset,
		domain.StepSwarmConfig, domain.StepSwarmDeploy, domain.StepSwarmHealth:
		return service.StepResult{}, nil
	}
	revision, err := domain.NewGitRevision(rawRevision)
	if err != nil {
		return service.StepResult{}, err
	}
	return service.StepResult{Revision: revision}, nil
}

type successfulAgentRunner struct{}

func (successfulAgentRunner) Run(_ context.Context, run service.AgentRun) error {
	if run.Output != nil {
		return run.Output("agent completed")
	}
	return nil
}

type fixedSecretReader struct{}

func (fixedSecretReader) Get(context.Context, string) (string, error) { return "password", nil }

type fixedAIModelCatalog struct{}

func (fixedAIModelCatalog) Models(_ context.Context, provider domain.AIProvider) ([]string, error) {
	return []string{provider.Model()}, nil
}

type fixedWorkerProbe struct{}

func (fixedWorkerProbe) Ping(context.Context, domain.Worker, string) (service.WorkerHeartbeat, error) {
	return service.WorkerHeartbeat{Version: "v1.0.0"}, nil
}

type testClock struct{}

func (testClock) Now() time.Time {
	return time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
}

type testIDs struct{ next int }

func (i *testIDs) New() (string, error) {
	i.next++
	return fmt.Sprintf("id-%d", i.next), nil
}
