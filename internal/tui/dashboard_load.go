package tui

import (
	"coffe/internal/domain"

	tea "charm.land/bubbletea/v2"
)

type dashboardOverview struct {
	agents        []domain.AgentSession
	databases     []domain.DatabaseConnection
	gitOperations []domain.GitOperation
	providers     []domain.AIProvider
	aiUsage       []domain.AIUsage
	policy        domain.Policy
	audit         []domain.AuditEntry
	workers       []domain.Worker
	plugins       []domain.Plugin
}

func (m Model) fetchDashboard() tea.Msg {
	projects, targets, err := m.fetchRegistry()
	if err != nil {
		return errorMsg{err: err}
	}
	plans, events, err := m.fetchPlans()
	if err != nil {
		return errorMsg{err: err}
	}
	overview, err := m.fetchOverview()
	if err != nil {
		return errorMsg{err: err}
	}
	return dashboardMsg{
		projects: projects, targets: targets, plans: plans, events: events,
		agents: overview.agents, databases: overview.databases,
		gitOperations: overview.gitOperations, providers: overview.providers, aiUsage: overview.aiUsage,
		policy: overview.policy, audit: overview.audit, workers: overview.workers, plugins: overview.plugins,
	}
}

func (m Model) fetchRegistry() ([]domain.Project, []domain.Target, error) {
	projects, err := m.app.ListProjects(m.ctx)
	if err != nil || len(projects) == 0 {
		return projects, nil, err
	}
	projectID := projects[0].ID()
	if selected, ok := m.selectedProject(); ok {
		projectID = selected.ID()
	}
	targets, err := m.app.ListTargets(m.ctx, string(projectID))
	return projects, targets, err
}

func (m Model) fetchPlans() ([]domain.ActionPlan, []domain.Event, error) {
	plans, err := m.app.ListPlans(m.ctx, 25)
	if err != nil || len(plans) == 0 {
		return plans, nil, err
	}
	planID := plans[0].ID()
	if selected, ok := m.selectedPlan(); ok {
		planID = selected.ID()
	}
	events, err := m.app.ListEvents(m.ctx, string(planID))
	return plans, events, err
}

func (m Model) fetchOverview() (dashboardOverview, error) {
	agents, err := m.app.ListAgentSessions(m.ctx, 25)
	if err != nil {
		return dashboardOverview{}, err
	}
	databases, err := m.app.ListDatabases(m.ctx)
	if err != nil {
		return dashboardOverview{}, err
	}
	gitOperations, err := m.app.ListGitOperations(m.ctx, 25)
	if err != nil {
		return dashboardOverview{}, err
	}
	providers, err := m.app.ListAIProviders(m.ctx)
	if err != nil {
		return dashboardOverview{}, err
	}
	aiUsage, err := m.app.ListAIUsage(m.ctx, 25)
	if err != nil {
		return dashboardOverview{}, err
	}
	policy, err := m.app.GetPolicy(m.ctx)
	if err != nil {
		return dashboardOverview{}, err
	}
	audit, err := m.app.ListAudit(m.ctx, 25)
	if err != nil {
		return dashboardOverview{}, err
	}
	workers, err := m.app.ListWorkers(m.ctx)
	if err != nil {
		return dashboardOverview{}, err
	}
	plugins, err := m.app.ListPlugins(m.ctx)
	if err != nil {
		return dashboardOverview{}, err
	}
	return dashboardOverview{
		agents: agents, databases: databases, gitOperations: gitOperations,
		providers: providers, aiUsage: aiUsage, policy: policy, audit: audit,
		workers: workers, plugins: plugins,
	}, nil
}
