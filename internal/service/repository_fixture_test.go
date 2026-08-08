package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"slices"
)

type memoryRepository struct {
	projects      map[domain.ProjectID]domain.Project
	targets       map[domain.TargetID]domain.Target
	plans         map[domain.PlanID]domain.ActionPlan
	deployments   map[domain.TargetID]domain.DeploymentRecord
	agentSessions map[domain.AgentSessionID]domain.AgentSession
	agentEvents   map[domain.AgentSessionID][]domain.AgentEvent
	databases     map[domain.DatabaseID]domain.DatabaseConnection
	events        map[domain.PlanID][]domain.Event
	gitOperations map[domain.GitOperationID]domain.GitOperation
	gitEvents     map[domain.GitOperationID][]domain.GitEvent
	aiProviders   map[domain.AIProviderID]domain.AIProvider
	aiUsage       []domain.AIUsage
	queryHistory  map[domain.DatabaseID][]domain.QueryHistory
	policy        domain.Policy
	hasPolicy     bool
	audit         []domain.AuditEntry
	workers       map[domain.WorkerID]domain.Worker
	plugins       map[domain.PluginID]domain.Plugin
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		projects: make(map[domain.ProjectID]domain.Project), targets: make(map[domain.TargetID]domain.Target),
		plans: make(map[domain.PlanID]domain.ActionPlan), deployments: make(map[domain.TargetID]domain.DeploymentRecord),
		agentSessions: make(map[domain.AgentSessionID]domain.AgentSession),
		agentEvents:   make(map[domain.AgentSessionID][]domain.AgentEvent),
		databases:     make(map[domain.DatabaseID]domain.DatabaseConnection), events: make(map[domain.PlanID][]domain.Event),
		gitOperations: make(map[domain.GitOperationID]domain.GitOperation),
		gitEvents:     make(map[domain.GitOperationID][]domain.GitEvent),
		aiProviders:   make(map[domain.AIProviderID]domain.AIProvider),
		queryHistory:  make(map[domain.DatabaseID][]domain.QueryHistory),
		workers:       make(map[domain.WorkerID]domain.Worker),
		plugins:       make(map[domain.PluginID]domain.Plugin),
	}
}

func (r *memoryRepository) CreateProject(_ context.Context, project domain.Project) error {
	r.projects[project.ID()] = project
	return nil
}

func (r *memoryRepository) ListProjects(_ context.Context) ([]domain.Project, error) {
	projects := make([]domain.Project, 0, len(r.projects))
	for _, project := range r.projects {
		projects = append(projects, project)
	}
	return projects, nil
}

func (r *memoryRepository) GetProject(_ context.Context, id domain.ProjectID) (domain.Project, error) {
	project, ok := r.projects[id]
	if !ok {
		return domain.Project{}, service.ErrNotFound
	}
	return project, nil
}

func (r *memoryRepository) CreateTarget(_ context.Context, target domain.Target) error {
	r.targets[target.ID()] = target
	return nil
}

func (r *memoryRepository) ListTargets(_ context.Context, projectID domain.ProjectID) ([]domain.Target, error) {
	var targets []domain.Target
	for _, target := range r.targets {
		if target.ProjectID() == projectID {
			targets = append(targets, target)
		}
	}
	return targets, nil
}

func (r *memoryRepository) GetTarget(_ context.Context, id domain.TargetID) (domain.Target, error) {
	target, ok := r.targets[id]
	if !ok {
		return domain.Target{}, service.ErrNotFound
	}
	return target, nil
}

func (r *memoryRepository) SavePlan(_ context.Context, plan domain.ActionPlan) error {
	r.plans[plan.ID()] = plan
	return nil
}

func (r *memoryRepository) GetPlan(_ context.Context, id domain.PlanID) (domain.ActionPlan, error) {
	plan, ok := r.plans[id]
	if !ok {
		return domain.ActionPlan{}, service.ErrNotFound
	}
	return plan, nil
}

func (r *memoryRepository) ListPlans(_ context.Context, limit int) ([]domain.ActionPlan, error) {
	plans := make([]domain.ActionPlan, 0, min(limit, len(r.plans)))
	for _, plan := range r.plans {
		if len(plans) == limit {
			break
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

func (r *memoryRepository) SaveDeployment(_ context.Context, record domain.DeploymentRecord) error {
	r.deployments[record.TargetID()] = record
	return nil
}

func (r *memoryRepository) LatestDeployment(
	_ context.Context,
	targetID domain.TargetID,
) (domain.DeploymentRecord, error) {
	record, ok := r.deployments[targetID]
	if !ok {
		return domain.DeploymentRecord{}, service.ErrNotFound
	}
	return record, nil
}

func (r *memoryRepository) SaveAgentSession(_ context.Context, session domain.AgentSession) error {
	r.agentSessions[session.ID()] = session
	return nil
}

func (r *memoryRepository) GetAgentSession(
	_ context.Context,
	id domain.AgentSessionID,
) (domain.AgentSession, error) {
	session, ok := r.agentSessions[id]
	if !ok {
		return domain.AgentSession{}, service.ErrNotFound
	}
	return session, nil
}

func (r *memoryRepository) ListAgentSessions(_ context.Context, limit int) ([]domain.AgentSession, error) {
	sessions := make([]domain.AgentSession, 0, min(limit, len(r.agentSessions)))
	for _, session := range r.agentSessions {
		if len(sessions) == limit {
			break
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (r *memoryRepository) AppendAgentEvent(_ context.Context, event domain.AgentEvent) error {
	r.agentEvents[event.SessionID()] = append(r.agentEvents[event.SessionID()], event)
	return nil
}

func (r *memoryRepository) ListAgentEvents(
	_ context.Context,
	id domain.AgentSessionID,
) ([]domain.AgentEvent, error) {
	return slices.Clone(r.agentEvents[id]), nil
}

func (r *memoryRepository) CreateDatabaseConnection(
	_ context.Context,
	connection domain.DatabaseConnection,
) error {
	r.databases[connection.ID()] = connection
	return nil
}

func (r *memoryRepository) GetDatabaseConnection(
	_ context.Context,
	id domain.DatabaseID,
) (domain.DatabaseConnection, error) {
	connection, ok := r.databases[id]
	if !ok {
		return domain.DatabaseConnection{}, service.ErrNotFound
	}
	return connection, nil
}

func (r *memoryRepository) ListDatabaseConnections(_ context.Context) ([]domain.DatabaseConnection, error) {
	connections := make([]domain.DatabaseConnection, 0, len(r.databases))
	for _, connection := range r.databases {
		connections = append(connections, connection)
	}
	return connections, nil
}

func (r *memoryRepository) AppendEvent(_ context.Context, event domain.Event) error {
	r.events[event.PlanID()] = append(r.events[event.PlanID()], event)
	return nil
}

func (r *memoryRepository) ListEvents(_ context.Context, planID domain.PlanID) ([]domain.Event, error) {
	return slices.Clone(r.events[planID]), nil
}
