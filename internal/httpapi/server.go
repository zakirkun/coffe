package httpapi

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type Application interface {
	CreateProject(context.Context, service.CreateProjectInput) (domain.Project, error)
	ListProjects(context.Context) ([]domain.Project, error)
	CreateTarget(context.Context, service.CreateTargetInput) (domain.Target, error)
	ListTargets(context.Context, string) ([]domain.Target, error)
	PreviewDeploy(context.Context, service.PreviewInput) (domain.ActionPlan, error)
	PlanPreview(context.Context, string) (service.Preview, error)
	ApprovePlan(context.Context, string) (domain.ActionPlan, error)
	ExecutePlan(context.Context, string) (domain.ActionPlan, error)
	StartPlan(context.Context, string) (domain.ActionPlan, error)
	CancelPlan(context.Context, string) error
	RetryPlan(context.Context, string) (domain.ActionPlan, error)
	PreviewRollback(context.Context, string) (domain.ActionPlan, error)
	ListSwarmNodes(context.Context, string) ([]domain.SwarmNode, error)
	ListSwarmServices(context.Context, string) ([]domain.SwarmService, error)
	StartAgent(context.Context, service.StartAgentInput) (domain.AgentSession, error)
	StopAgent(context.Context, string) error
	ListAgentSessions(context.Context, int) ([]domain.AgentSession, error)
	ListAgentEvents(context.Context, string) ([]domain.AgentEvent, error)
	CreateDatabase(context.Context, service.CreateDatabaseInput) (domain.DatabaseConnection, error)
	ListDatabases(context.Context) ([]domain.DatabaseConnection, error)
	QueryDatabase(context.Context, string, string) (domain.QueryResult, error)
	ListDatabaseTables(context.Context, string) ([]domain.DatabaseTable, error)
	ListDatabaseColumns(context.Context, string, string) ([]domain.DatabaseColumn, error)
	QueryDatabasePage(context.Context, string, service.QueryPageInput) (domain.QueryResult, error)
	ExplainDatabase(context.Context, string, string) (domain.QueryResult, error)
	ExportDatabaseCSV(context.Context, string, string) ([]byte, error)
	ListQueryHistory(context.Context, string, int) ([]domain.QueryHistory, error)
	GitStatus(context.Context, string) (domain.GitStatus, error)
	GitBranches(context.Context, string) ([]string, error)
	GitDiff(context.Context, string) (string, error)
	AllGitStatuses(context.Context) ([]service.ProjectGitStatus, error)
	PreviewGit(context.Context, service.CreateGitOperationInput) (service.GitPreview, error)
	ApproveGit(context.Context, string) (domain.GitOperation, error)
	ExecuteGit(context.Context, string) (domain.GitOperation, error)
	ListGitOperations(context.Context, int) ([]domain.GitOperation, error)
	ListGitEvents(context.Context, string) ([]domain.GitEvent, error)
	CreateAIProvider(context.Context, service.CreateAIProviderInput) (domain.AIProvider, error)
	ListAIProviders(context.Context) ([]domain.AIProvider, error)
	ListAIModels(context.Context, string) ([]string, error)
	ListAIUsage(context.Context, int) ([]domain.AIUsage, error)
	GetPolicy(context.Context) (domain.Policy, error)
	UpdatePolicy(context.Context, service.UpdatePolicyInput) (domain.Policy, error)
	ListAudit(context.Context, int) ([]domain.AuditEntry, error)
	ExportAuditCSV(context.Context, int) ([]byte, error)
	RegisterWorker(context.Context, service.RegisterWorkerInput) (domain.Worker, error)
	HeartbeatWorker(context.Context, string) (domain.Worker, error)
	ListWorkers(context.Context) ([]domain.Worker, error)
	RegisterPlugin(context.Context, service.RegisterPluginInput) (domain.Plugin, error)
	ListPlugins(context.Context) ([]domain.Plugin, error)
	InvokePlugin(context.Context, service.PluginInvocationInput) (json.RawMessage, error)
	GetPlan(context.Context, string) (domain.ActionPlan, error)
	ListPlans(context.Context, int) ([]domain.ActionPlan, error)
	ListEvents(context.Context, string) ([]domain.Event, error)
}

type Server struct {
	app    Application
	logger *slog.Logger
	mux    *http.ServeMux
}

func New(app Application, logger *slog.Logger) *Server {
	server := &Server{app: app, logger: logger, mux: http.NewServeMux()}
	server.routes()
	return server
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.HandleFunc("GET /api/v1/projects", s.listProjects)
	s.mux.HandleFunc("POST /api/v1/projects", s.createProject)
	s.mux.HandleFunc("POST /api/v1/targets", s.createTarget)
	s.mux.HandleFunc("GET /api/v1/projects/{projectID}/targets", s.listTargets)
	s.mux.HandleFunc("POST /api/v1/plans/preview", s.previewPlan)
	s.mux.HandleFunc("GET /api/v1/plans", s.listPlans)
	s.mux.HandleFunc("GET /api/v1/plans/{planID}", s.getPlan)
	s.mux.HandleFunc("POST /api/v1/plans/{planID}/approve", s.approvePlan)
	s.mux.HandleFunc("POST /api/v1/plans/{planID}/execute", s.executePlan)
	s.mux.HandleFunc("POST /api/v1/plans/{planID}/cancel", s.cancelPlan)
	s.mux.HandleFunc("POST /api/v1/plans/{planID}/retry", s.retryPlan)
	s.mux.HandleFunc("POST /api/v1/targets/{targetID}/rollback-preview", s.previewRollback)
	s.mux.HandleFunc("GET /api/v1/targets/{targetID}/swarm/nodes", s.listSwarmNodes)
	s.mux.HandleFunc("GET /api/v1/targets/{targetID}/swarm/services", s.listSwarmServices)
	s.mux.HandleFunc("POST /api/v1/agents", s.startAgent)
	s.mux.HandleFunc("GET /api/v1/agents", s.listAgents)
	s.mux.HandleFunc("POST /api/v1/agents/{sessionID}/stop", s.stopAgent)
	s.mux.HandleFunc("GET /api/v1/agents/{sessionID}/events", s.listAgentEvents)
	s.mux.HandleFunc("POST /api/v1/databases", s.createDatabase)
	s.mux.HandleFunc("GET /api/v1/databases", s.listDatabases)
	s.mux.HandleFunc("POST /api/v1/databases/{databaseID}/query", s.queryDatabase)
	s.mux.HandleFunc("GET /api/v1/databases/{databaseID}/tables", s.listDatabaseTables)
	s.mux.HandleFunc("GET /api/v1/databases/{databaseID}/tables/{tableName}/columns", s.listDatabaseColumns)
	s.mux.HandleFunc("POST /api/v1/databases/{databaseID}/explain", s.explainDatabase)
	s.mux.HandleFunc("POST /api/v1/databases/{databaseID}/export.csv", s.exportDatabaseCSV)
	s.mux.HandleFunc("GET /api/v1/databases/{databaseID}/history", s.listQueryHistory)
	s.mux.HandleFunc("GET /api/v1/projects/{projectID}/git/status", s.gitStatus)
	s.mux.HandleFunc("GET /api/v1/projects/{projectID}/git/branches", s.gitBranches)
	s.mux.HandleFunc("GET /api/v1/projects/{projectID}/git/diff", s.gitDiff)
	s.mux.HandleFunc("GET /api/v1/git/statuses", s.allGitStatuses)
	s.mux.HandleFunc("POST /api/v1/git/operations/preview", s.previewGit)
	s.mux.HandleFunc("GET /api/v1/git/operations", s.listGitOperations)
	s.mux.HandleFunc("POST /api/v1/git/operations/{operationID}/approve", s.approveGit)
	s.mux.HandleFunc("POST /api/v1/git/operations/{operationID}/execute", s.executeGit)
	s.mux.HandleFunc("GET /api/v1/git/operations/{operationID}/events", s.listGitEvents)
	s.mux.HandleFunc("POST /api/v1/ai/providers", s.createAIProvider)
	s.mux.HandleFunc("GET /api/v1/ai/providers", s.listAIProviders)
	s.mux.HandleFunc("GET /api/v1/ai/providers/{providerID}/models", s.listAIModels)
	s.mux.HandleFunc("GET /api/v1/ai/usage", s.listAIUsage)
	s.mux.HandleFunc("GET /api/v1/policy", s.getPolicy)
	s.mux.HandleFunc("PUT /api/v1/policy", s.updatePolicy)
	s.mux.HandleFunc("GET /api/v1/audit", s.listAudit)
	s.mux.HandleFunc("GET /api/v1/audit/export.csv", s.exportAuditCSV)
	s.mux.HandleFunc("POST /api/v1/workers", s.registerWorker)
	s.mux.HandleFunc("GET /api/v1/workers", s.listWorkers)
	s.mux.HandleFunc("POST /api/v1/workers/{workerID}/heartbeat", s.heartbeatWorker)
	s.mux.HandleFunc("POST /api/v1/plugins", s.registerPlugin)
	s.mux.HandleFunc("GET /api/v1/plugins", s.listPlugins)
	s.mux.HandleFunc("POST /api/v1/plugins/{pluginID}/invoke", s.invokePlugin)
	s.mux.HandleFunc("GET /api/v1/plans/{planID}/events", s.listEvents)
}

func (s *Server) health(writer http.ResponseWriter, request *http.Request) {
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: map[string]string{"status": "ready"}})
}
