package tui

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

const (
	minWidth  = 80
	minHeight = 24
)

type Application interface {
	ListProjects(context.Context) ([]domain.Project, error)
	ListTargets(context.Context, string) ([]domain.Target, error)
	ListPlans(context.Context, int) ([]domain.ActionPlan, error)
	ListEvents(context.Context, string) ([]domain.Event, error)
	PreviewDeploy(context.Context, service.PreviewInput) (domain.ActionPlan, error)
	PlanPreview(context.Context, string) (service.Preview, error)
	ApprovePlan(context.Context, string) (domain.ActionPlan, error)
	ExecutePlan(context.Context, string) (domain.ActionPlan, error)
	StartPlan(context.Context, string) (domain.ActionPlan, error)
	CancelPlan(context.Context, string) error
	RetryPlan(context.Context, string) (domain.ActionPlan, error)
	PreviewRollback(context.Context, string) (domain.ActionPlan, error)
	StartAgent(context.Context, service.StartAgentInput) (domain.AgentSession, error)
	ListAgentSessions(context.Context, int) ([]domain.AgentSession, error)
	ListDatabases(context.Context) ([]domain.DatabaseConnection, error)
	QueryDatabase(context.Context, string, string) (domain.QueryResult, error)
	ListSwarmNodes(context.Context, string) ([]domain.SwarmNode, error)
	ListSwarmServices(context.Context, string) ([]domain.SwarmService, error)
	GitStatus(context.Context, string) (domain.GitStatus, error)
	PreviewGit(context.Context, service.CreateGitOperationInput) (service.GitPreview, error)
	ApproveGit(context.Context, string) (domain.GitOperation, error)
	ExecuteGit(context.Context, string) (domain.GitOperation, error)
	ListGitOperations(context.Context, int) ([]domain.GitOperation, error)
	ListAIProviders(context.Context) ([]domain.AIProvider, error)
	ListAIUsage(context.Context, int) ([]domain.AIUsage, error)
	GetPolicy(context.Context) (domain.Policy, error)
	ListAudit(context.Context, int) ([]domain.AuditEntry, error)
	ListWorkers(context.Context) ([]domain.Worker, error)
	ListPlugins(context.Context) ([]domain.Plugin, error)
}

type focusArea uint8

const (
	focusProjects focusArea = iota
	focusTargets
	focusPrompt
	focusJobs
)

type Model struct {
	ctx           context.Context
	app           Application
	styles        styles
	prompt        textarea.Model
	projects      []domain.Project
	targets       []domain.Target
	plans         []domain.ActionPlan
	events        []domain.Event
	agents        []domain.AgentSession
	databases     []domain.DatabaseConnection
	swarmNodes    []domain.SwarmNode
	swarmServices []domain.SwarmService
	queryResult   domain.QueryResult
	gitStatus     domain.GitStatus
	gitOperations []domain.GitOperation
	providers     []domain.AIProvider
	aiUsage       []domain.AIUsage
	policy        domain.Policy
	audit         []domain.AuditEntry
	workers       []domain.Worker
	plugins       []domain.Plugin
	commands      []string
	projectIndex  int
	targetIndex   int
	planIndex     int
	width         int
	height        int
	focus         focusArea
	status        string
	loading       bool
}

func NewModel(ctx context.Context, app Application) Model {
	prompt := textarea.New()
	prompt.Placeholder = "deploy, git, agent TOOL, or sql query"
	prompt.ShowLineNumbers = false
	prompt.CharLimit = 2_000
	prompt.SetVirtualCursor(false)
	prompt.SetHeight(3)
	prompt.SetWidth(50)
	prompt.SetStyles(promptStyles())
	return Model{
		ctx: ctx, app: app, styles: newStyles(), prompt: prompt,
		focus: focusProjects, status: "Ready",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadDashboard(), tick())
}

func (m Model) selectedProject() (domain.Project, bool) {
	if m.projectIndex < 0 || m.projectIndex >= len(m.projects) {
		return domain.Project{}, false
	}
	return m.projects[m.projectIndex], true
}

func (m Model) selectedTarget() (domain.Target, bool) {
	if m.targetIndex < 0 || m.targetIndex >= len(m.targets) {
		return domain.Target{}, false
	}
	return m.targets[m.targetIndex], true
}

func (m Model) selectedPlan() (domain.ActionPlan, bool) {
	if m.planIndex < 0 || m.planIndex >= len(m.plans) {
		return domain.ActionPlan{}, false
	}
	return m.plans[m.planIndex], true
}
