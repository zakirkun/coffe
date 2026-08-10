package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"encoding/json"
	"errors"
	"sync"
)

type fixedWorkflowLoader struct{ workflow domain.Workflow }

func (l fixedWorkflowLoader) Load(context.Context, domain.Project) (domain.Workflow, error) {
	return l.workflow, nil
}

type fixedPlanner struct{ intent domain.Intent }

func (p fixedPlanner) Classify(context.Context, string) (domain.Intent, error) { return p.intent, nil }

type fixedRenderer struct{}

func (fixedRenderer) Render(execution service.ExecutionStep) (string, error) {
	return "rendered:" + string(execution.Step), nil
}

type fixedSwarmReader struct{}

func (fixedSwarmReader) ListNodes(context.Context, domain.Target) ([]domain.SwarmNode, error) {
	return nil, nil
}

func (fixedSwarmReader) ListServices(
	context.Context,
	domain.Target,
	domain.StackName,
) ([]domain.SwarmService, error) {
	return nil, nil
}

type successfulAgentRunner struct{}

func (successfulAgentRunner) Run(_ context.Context, run service.AgentRun) error {
	if run.Output != nil {
		return run.Output("agent completed")
	}
	return nil
}

type fixedDatabaseReader struct{}

func (fixedDatabaseReader) Query(context.Context, service.DatabaseQuery) (domain.QueryResult, error) {
	return domain.QueryResult{}, nil
}

func (fixedDatabaseReader) Tables(context.Context, service.DatabaseAccess) ([]domain.DatabaseTable, error) {
	return nil, nil
}

func (fixedDatabaseReader) Columns(
	context.Context,
	service.DatabaseAccess,
	string,
) ([]domain.DatabaseColumn, error) {
	return nil, nil
}

type fixedSecretReader struct{}

func (fixedSecretReader) Get(context.Context, string) (string, error) { return "password", nil }

type fixedGitTool struct {
	cancel     context.CancelFunc
	executions int
}

func (fixedGitTool) Status(context.Context, domain.Project) (domain.GitStatus, error) {
	return domain.GitStatus{Branch: "main"}, nil
}

func (fixedGitTool) Branches(context.Context, domain.Project) ([]string, error) {
	return []string{"main"}, nil
}

func (fixedGitTool) Diff(context.Context, domain.Project) (string, error) { return "", nil }

func (fixedGitTool) Render(operation domain.GitOperation) string {
	return "git " + string(operation.Kind())
}

func (g *fixedGitTool) Execute(context.Context, domain.Project, domain.GitOperation) (string, error) {
	g.executions++
	if g.cancel != nil {
		g.cancel()
	}
	return "git completed", nil
}

type fixedAIModelCatalog struct{}

func (fixedAIModelCatalog) Models(_ context.Context, provider domain.AIProvider) ([]string, error) {
	return []string{provider.Model()}, nil
}

type fixedWorkerProbe struct{}

func (fixedWorkerProbe) Ping(context.Context, domain.Worker, string) (service.WorkerHeartbeat, error) {
	return service.WorkerHeartbeat{Version: "v1.0.0"}, nil
}

type fixedPluginRunner struct{}

func (fixedPluginRunner) Verify(domain.Plugin) error { return nil }

func (fixedPluginRunner) Invoke(
	context.Context,
	domain.Plugin,
	string,
	json.RawMessage,
) (json.RawMessage, error) {
	return json.RawMessage(`{"ok":true}`), nil
}

type recordingExecutor struct {
	steps      []domain.StepKind
	failAt     domain.StepKind
	blockAt    domain.StepKind
	started    chan struct{}
	finished   chan struct{}
	startOnce  sync.Once
	finishOnce sync.Once
}

func (e *recordingExecutor) Execute(
	ctx context.Context,
	execution service.ExecutionStep,
) (service.StepResult, error) {
	e.steps = append(e.steps, execution.Step)
	if execution.Step == e.blockAt {
		e.startOnce.Do(func() { close(e.started) })
		<-ctx.Done()
		e.finishOnce.Do(func() { close(e.finished) })
		return service.StepResult{}, ctx.Err()
	}
	if execution.Step == e.failAt {
		return service.StepResult{}, errors.New("executor failed")
	}
	if execution.Output != nil {
		if err := execution.Output("token=secret build output"); err != nil {
			return service.StepResult{}, err
		}
	}
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
