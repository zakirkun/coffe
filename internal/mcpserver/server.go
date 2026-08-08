package mcpserver

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Application interface {
	ListProjects(context.Context) ([]domain.Project, error)
	ListTargets(context.Context, string) ([]domain.Target, error)
	PreviewDeploy(context.Context, service.PreviewInput) (domain.ActionPlan, error)
	PlanPreview(context.Context, string) (service.Preview, error)
	GetPlan(context.Context, string) (domain.ActionPlan, error)
	ListSwarmNodes(context.Context, string) ([]domain.SwarmNode, error)
	ListSwarmServices(context.Context, string) ([]domain.SwarmService, error)
	ListAgentSessions(context.Context, int) ([]domain.AgentSession, error)
	ListAgentEvents(context.Context, string) ([]domain.AgentEvent, error)
	ListDatabases(context.Context) ([]domain.DatabaseConnection, error)
	QueryDatabase(context.Context, string, string) (domain.QueryResult, error)
	ListDatabaseTables(context.Context, string) ([]domain.DatabaseTable, error)
	ListDatabaseColumns(context.Context, string, string) ([]domain.DatabaseColumn, error)
	QueryDatabasePage(context.Context, string, service.QueryPageInput) (domain.QueryResult, error)
	ExplainDatabase(context.Context, string, string) (domain.QueryResult, error)
	ListQueryHistory(context.Context, string, int) ([]domain.QueryHistory, error)
	GitStatus(context.Context, string) (domain.GitStatus, error)
	GitDiff(context.Context, string) (string, error)
	PreviewGit(context.Context, service.CreateGitOperationInput) (service.GitPreview, error)
	ListGitOperations(context.Context, int) ([]domain.GitOperation, error)
	ListAIProviders(context.Context) ([]domain.AIProvider, error)
	ListAIModels(context.Context, string) ([]string, error)
	ListAIUsage(context.Context, int) ([]domain.AIUsage, error)
	GetPolicy(context.Context) (domain.Policy, error)
	ListAudit(context.Context, int) ([]domain.AuditEntry, error)
	ListWorkers(context.Context) ([]domain.Worker, error)
	ListPlugins(context.Context) ([]domain.Plugin, error)
}

func New(app Application) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "coffe", Version: "0.1.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_projects", Description: "List projects registered on this Coffe workstation",
	}, listProjectsHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_targets", Description: "List deployment targets for one registered project",
	}, listTargetsHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "preview_deploy", Description: "Create a deployment preview that still requires human approval",
	}, previewDeployHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_plan", Description: "Read one deployment plan and its current status",
	}, getPlanHandler(app))
	registerSwarmTools(server, app)
	registerAgentTools(server, app)
	registerDatabaseTools(server, app)
	registerGitTools(server, app)
	registerAIProviderTools(server, app)
	registerPolicyAuditTools(server, app)
	registerWorkerTools(server, app)
	registerPluginTools(server, app)
	return server
}

func RunStdio(ctx context.Context, app Application) error {
	return New(app).Run(ctx, &mcp.StdioTransport{})
}

func HTTPHandler(app Application) http.Handler {
	server := New(app)
	return mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{JSONResponse: true},
	)
}
