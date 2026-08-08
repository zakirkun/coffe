package mcpserver_test

import (
	"bytes"
	"coffe/internal/domain"
	"coffe/internal/mcpserver"
	"coffe/internal/service"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_HTTPHandler_initializes_coffe_server(t *testing.T) {
	// Given
	handler := mcpserver.HTTPHandler(emptyApplication{})
	body := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	request := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	recorder := httptest.NewRecorder()

	// When
	handler.ServeHTTP(recorder, request)

	// Then
	if recorder.Code != http.StatusOK {
		t.Fatalf("initialize status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Result struct {
			ServerInfo struct {
				Name string `json:"name"`
			} `json:"serverInfo"`
		} `json:"result"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if response.Result.ServerInfo.Name != "coffe" {
		t.Fatalf("server name = %q, want coffe", response.Result.ServerInfo.Name)
	}
}

type emptyApplication struct{}

func (emptyApplication) ListProjects(context.Context) ([]domain.Project, error) {
	return nil, nil
}

func (emptyApplication) ListTargets(context.Context, string) ([]domain.Target, error) {
	return nil, nil
}

func (emptyApplication) PreviewDeploy(context.Context, service.PreviewInput) (domain.ActionPlan, error) {
	return domain.ActionPlan{}, nil
}

func (emptyApplication) PlanPreview(context.Context, string) (service.Preview, error) {
	return service.Preview{}, nil
}

func (emptyApplication) GetPlan(context.Context, string) (domain.ActionPlan, error) {
	return domain.ActionPlan{}, nil
}

func (emptyApplication) ListSwarmNodes(context.Context, string) ([]domain.SwarmNode, error) {
	return nil, nil
}

func (emptyApplication) ListSwarmServices(context.Context, string) ([]domain.SwarmService, error) {
	return nil, nil
}

func (emptyApplication) ListAgentSessions(context.Context, int) ([]domain.AgentSession, error) {
	return nil, nil
}

func (emptyApplication) ListAgentEvents(context.Context, string) ([]domain.AgentEvent, error) {
	return nil, nil
}

func (emptyApplication) ListDatabases(context.Context) ([]domain.DatabaseConnection, error) {
	return nil, nil
}

func (emptyApplication) QueryDatabase(context.Context, string, string) (domain.QueryResult, error) {
	return domain.QueryResult{}, nil
}

func (emptyApplication) ListDatabaseTables(context.Context, string) ([]domain.DatabaseTable, error) {
	return nil, nil
}

func (emptyApplication) ListDatabaseColumns(
	context.Context,
	string,
	string,
) ([]domain.DatabaseColumn, error) {
	return nil, nil
}

func (emptyApplication) QueryDatabasePage(
	context.Context,
	string,
	service.QueryPageInput,
) (domain.QueryResult, error) {
	return domain.QueryResult{}, nil
}

func (emptyApplication) ExplainDatabase(context.Context, string, string) (domain.QueryResult, error) {
	return domain.QueryResult{}, nil
}

func (emptyApplication) ListQueryHistory(context.Context, string, int) ([]domain.QueryHistory, error) {
	return nil, nil
}

func (emptyApplication) GitStatus(context.Context, string) (domain.GitStatus, error) {
	return domain.GitStatus{}, nil
}

func (emptyApplication) GitDiff(context.Context, string) (string, error) { return "", nil }

func (emptyApplication) PreviewGit(
	context.Context,
	service.CreateGitOperationInput,
) (service.GitPreview, error) {
	return service.GitPreview{}, nil
}

func (emptyApplication) ListGitOperations(context.Context, int) ([]domain.GitOperation, error) {
	return nil, nil
}

func (emptyApplication) ListAIProviders(context.Context) ([]domain.AIProvider, error) {
	return nil, nil
}

func (emptyApplication) ListAIModels(context.Context, string) ([]string, error) { return nil, nil }

func (emptyApplication) ListAIUsage(context.Context, int) ([]domain.AIUsage, error) { return nil, nil }

func (emptyApplication) GetPolicy(context.Context) (domain.Policy, error) {
	return domain.DefaultPolicy(time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)), nil
}

func (emptyApplication) ListAudit(context.Context, int) ([]domain.AuditEntry, error) {
	return nil, nil
}

func (emptyApplication) ListWorkers(context.Context) ([]domain.Worker, error) { return nil, nil }

func (emptyApplication) ListPlugins(context.Context) ([]domain.Plugin, error) { return nil, nil }
