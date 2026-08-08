package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

type App struct {
	logger     *slog.Logger
	repository Repository
	workflows  WorkflowLoader
	planner    IntentPlanner
	executor   Executor
	renderer   CommandRenderer
	swarm      SwarmReader
	agents     AgentRunner
	databases  DatabaseReader
	secrets    SecretReader
	git        GitTool
	aiModels   AIModelCatalog
	workers    WorkerProbe
	plugins    PluginRunner
	clock      Clock
	ids        IDSource
	startMu    sync.Mutex
	runMu      sync.Mutex
	runs       map[domain.PlanID]context.CancelFunc
	agentRuns  map[domain.AgentSessionID]context.CancelFunc
	runWG      sync.WaitGroup
	policyMu   sync.Mutex
}

func New(dependencies Dependencies) *App {
	return &App{
		logger:     dependencies.Logger,
		repository: dependencies.Repository,
		workflows:  dependencies.Workflows,
		planner:    dependencies.Planner,
		executor:   dependencies.Executor,
		renderer:   dependencies.Renderer,
		swarm:      dependencies.Swarm,
		agents:     dependencies.Agents,
		databases:  dependencies.Databases,
		secrets:    dependencies.Secrets,
		git:        dependencies.Git,
		aiModels:   dependencies.AIModels,
		workers:    dependencies.Workers,
		plugins:    dependencies.Plugins,
		clock:      dependencies.Clock,
		ids:        dependencies.IDs,
		runs:       make(map[domain.PlanID]context.CancelFunc),
		agentRuns:  make(map[domain.AgentSessionID]context.CancelFunc),
	}
}

func (a *App) ListProjects(ctx context.Context) ([]domain.Project, error) {
	projects, err := a.repository.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projects, nil
}

func (a *App) ListTargets(ctx context.Context, projectID string) ([]domain.Target, error) {
	id, err := parseProjectID(projectID)
	if err != nil {
		return nil, err
	}
	targets, err := a.repository.ListTargets(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list targets: %w", err)
	}
	return targets, nil
}

func (a *App) GetPlan(ctx context.Context, planID string) (domain.ActionPlan, error) {
	id, err := parsePlanID(planID)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	plan, err := a.repository.GetPlan(ctx, id)
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("get plan: %w", err)
	}
	return plan, nil
}

func (a *App) ListPlans(ctx context.Context, limit int) ([]domain.ActionPlan, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("plan limit: %w", ErrInvalidInput)
	}
	plans, err := a.repository.ListPlans(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	return plans, nil
}

func (a *App) ListEvents(ctx context.Context, planID string) ([]domain.Event, error) {
	id, err := parsePlanID(planID)
	if err != nil {
		return nil, err
	}
	events, err := a.repository.ListEvents(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return events, nil
}

func parseProjectID(raw string) (domain.ProjectID, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" || clean != raw {
		return "", fmt.Errorf("project id: %w", ErrInvalidInput)
	}
	return domain.ProjectID(clean), nil
}

func parseTargetID(raw string) (domain.TargetID, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" || clean != raw {
		return "", fmt.Errorf("target id: %w", ErrInvalidInput)
	}
	return domain.TargetID(clean), nil
}

func parsePlanID(raw string) (domain.PlanID, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" || clean != raw {
		return "", fmt.Errorf("plan id: %w", ErrInvalidInput)
	}
	return domain.PlanID(clean), nil
}
