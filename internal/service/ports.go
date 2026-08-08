package service

import (
	"coffe/internal/domain"
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

type Repository interface {
	CreateProject(context.Context, domain.Project) error
	ListProjects(context.Context) ([]domain.Project, error)
	GetProject(context.Context, domain.ProjectID) (domain.Project, error)
	CreateTarget(context.Context, domain.Target) error
	ListTargets(context.Context, domain.ProjectID) ([]domain.Target, error)
	GetTarget(context.Context, domain.TargetID) (domain.Target, error)
	SavePlan(context.Context, domain.ActionPlan) error
	GetPlan(context.Context, domain.PlanID) (domain.ActionPlan, error)
	ListPlans(context.Context, int) ([]domain.ActionPlan, error)
	SaveDeployment(context.Context, domain.DeploymentRecord) error
	LatestDeployment(context.Context, domain.TargetID) (domain.DeploymentRecord, error)
	SaveAgentSession(context.Context, domain.AgentSession) error
	GetAgentSession(context.Context, domain.AgentSessionID) (domain.AgentSession, error)
	ListAgentSessions(context.Context, int) ([]domain.AgentSession, error)
	AppendAgentEvent(context.Context, domain.AgentEvent) error
	ListAgentEvents(context.Context, domain.AgentSessionID) ([]domain.AgentEvent, error)
	CreateDatabaseConnection(context.Context, domain.DatabaseConnection) error
	GetDatabaseConnection(context.Context, domain.DatabaseID) (domain.DatabaseConnection, error)
	ListDatabaseConnections(context.Context) ([]domain.DatabaseConnection, error)
	SaveGitOperation(context.Context, domain.GitOperation) error
	GetGitOperation(context.Context, domain.GitOperationID) (domain.GitOperation, error)
	ListGitOperations(context.Context, int) ([]domain.GitOperation, error)
	AppendGitEvent(context.Context, domain.GitOperationID, string, time.Time) error
	ListGitEvents(context.Context, domain.GitOperationID) ([]domain.GitEvent, error)
	SaveAIProvider(context.Context, domain.AIProvider) error
	GetAIProvider(context.Context, domain.AIProviderID) (domain.AIProvider, error)
	ListAIProviders(context.Context) ([]domain.AIProvider, error)
	SaveAIUsage(context.Context, domain.AIUsage) error
	ListAIUsage(context.Context, int) ([]domain.AIUsage, error)
	SaveQueryHistory(context.Context, domain.QueryHistory) error
	ListQueryHistory(context.Context, domain.DatabaseID, int) ([]domain.QueryHistory, error)
	SavePolicy(context.Context, domain.Policy) error
	GetPolicy(context.Context) (domain.Policy, error)
	AppendAudit(context.Context, domain.AuditEntry) error
	ListAudit(context.Context, int) ([]domain.AuditEntry, error)
	DeleteAuditBefore(context.Context, time.Time) error
	SaveWorker(context.Context, domain.Worker) error
	GetWorker(context.Context, domain.WorkerID) (domain.Worker, error)
	GetWorkerByTarget(context.Context, domain.TargetID) (domain.Worker, error)
	ListWorkers(context.Context) ([]domain.Worker, error)
	SavePlugin(context.Context, domain.Plugin) error
	GetPlugin(context.Context, domain.PluginID) (domain.Plugin, error)
	ListPlugins(context.Context) ([]domain.Plugin, error)
	AppendEvent(context.Context, domain.Event) error
	ListEvents(context.Context, domain.PlanID) ([]domain.Event, error)
}

type WorkflowLoader interface {
	Load(context.Context, domain.Project) (domain.Workflow, error)
}

type IntentPlanner interface {
	Classify(context.Context, string) (domain.Intent, error)
}

type Executor interface {
	Execute(context.Context, ExecutionStep) (StepResult, error)
}

type CommandRenderer interface {
	Render(ExecutionStep) (string, error)
}

type SwarmReader interface {
	ListNodes(context.Context, domain.Target) ([]domain.SwarmNode, error)
	ListServices(context.Context, domain.Target, domain.StackName) ([]domain.SwarmService, error)
}

type AgentRunner interface {
	Run(context.Context, AgentRun) error
}

type DatabaseReader interface {
	Query(context.Context, DatabaseQuery) (domain.QueryResult, error)
	Tables(context.Context, DatabaseAccess) ([]domain.DatabaseTable, error)
	Columns(context.Context, DatabaseAccess, string) ([]domain.DatabaseColumn, error)
}

type SecretReader interface {
	Get(context.Context, string) (string, error)
}

type GitTool interface {
	Status(context.Context, domain.Project) (domain.GitStatus, error)
	Branches(context.Context, domain.Project) ([]string, error)
	Diff(context.Context, domain.Project) (string, error)
	Render(domain.GitOperation) string
	Execute(context.Context, domain.Project, domain.GitOperation) (string, error)
}

type AIModelCatalog interface {
	Models(context.Context, domain.AIProvider) ([]string, error)
}

type WorkerProbe interface {
	Ping(context.Context, domain.Worker, string) (WorkerHeartbeat, error)
}

type WorkerHeartbeat struct{ Version string }

type PluginRunner interface {
	Verify(domain.Plugin) error
	Invoke(context.Context, domain.Plugin, string, json.RawMessage) (json.RawMessage, error)
}

type DatabaseAccess struct {
	Connection domain.DatabaseConnection
	Password   string
}

type DatabaseQuery struct {
	Connection domain.DatabaseConnection
	Password   string
	SQL        string
	Offset     int
	Limit      int
}

type AgentRun struct {
	Tool       domain.AgentTool
	WorkingDir string
	Prompt     string
	Output     func(string) error
}

type ExecutionStep struct {
	Plan     domain.ActionPlan
	Target   domain.Target
	Workflow domain.Workflow
	Step     domain.StepKind
	Output   func(string) error
}

type StepResult struct{ Revision domain.GitRevision }

type Clock interface {
	Now() time.Time
}

type IDSource interface {
	New() (string, error)
}

type Dependencies struct {
	Logger     *slog.Logger
	Repository Repository
	Workflows  WorkflowLoader
	Planner    IntentPlanner
	Executor   Executor
	Renderer   CommandRenderer
	Swarm      SwarmReader
	Agents     AgentRunner
	Databases  DatabaseReader
	Secrets    SecretReader
	Git        GitTool
	AIModels   AIModelCatalog
	Workers    WorkerProbe
	Plugins    PluginRunner
	Clock      Clock
	IDs        IDSource
}
