package tui_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"coffe/internal/tui"
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func Test_Model_View_renders_dashboard_regions(t *testing.T) {
	// Given
	model := tui.NewModel(context.Background(), emptyApplication{})

	// When
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
	view := updated.(tui.Model).View().Content

	// Then
	for _, label := range []string{"Projects / Targets", "Prompt", "Jobs", "Live Logs"} {
		if !strings.Contains(view, label) {
			t.Fatalf("View() missing %q:\n%s", label, view)
		}
	}
}

func Test_Model_View_requests_resize_below_minimum(t *testing.T) {
	// Given
	model := tui.NewModel(context.Background(), emptyApplication{})

	// When
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 70, Height: 20})
	view := updated.(tui.Model).View().Content

	// Then
	if !strings.Contains(view, "80x24") {
		t.Fatalf("View() = %q, want resize guidance", view)
	}
}

type emptyApplication struct{}

func (emptyApplication) ListProjects(context.Context) ([]domain.Project, error) { return nil, nil }

func (emptyApplication) ListTargets(context.Context, string) ([]domain.Target, error) {
	return nil, nil
}

func (emptyApplication) ListPlans(context.Context, int) ([]domain.ActionPlan, error) { return nil, nil }

func (emptyApplication) ListEvents(context.Context, string) ([]domain.Event, error) { return nil, nil }

func (emptyApplication) PreviewDeploy(context.Context, service.PreviewInput) (domain.ActionPlan, error) {
	return domain.ActionPlan{}, nil
}

func (emptyApplication) PlanPreview(context.Context, string) (service.Preview, error) {
	return service.Preview{}, nil
}

func (emptyApplication) ApprovePlan(context.Context, string) (domain.ActionPlan, error) {
	return domain.ActionPlan{}, nil
}

func (emptyApplication) ExecutePlan(context.Context, string) (domain.ActionPlan, error) {
	return domain.ActionPlan{}, nil
}

func (emptyApplication) StartPlan(context.Context, string) (domain.ActionPlan, error) {
	return domain.ActionPlan{}, nil
}

func (emptyApplication) CancelPlan(context.Context, string) error { return nil }

func (emptyApplication) RetryPlan(context.Context, string) (domain.ActionPlan, error) {
	return domain.ActionPlan{}, nil
}

func (emptyApplication) PreviewRollback(context.Context, string) (domain.ActionPlan, error) {
	return domain.ActionPlan{}, nil
}

func (emptyApplication) StartAgent(context.Context, service.StartAgentInput) (domain.AgentSession, error) {
	return domain.AgentSession{}, nil
}

func (emptyApplication) ListAgentSessions(context.Context, int) ([]domain.AgentSession, error) {
	return nil, nil
}

func (emptyApplication) ListDatabases(context.Context) ([]domain.DatabaseConnection, error) {
	return nil, nil
}

func (emptyApplication) QueryDatabase(context.Context, string, string) (domain.QueryResult, error) {
	return domain.QueryResult{}, nil
}

func (emptyApplication) ListSwarmNodes(context.Context, string) ([]domain.SwarmNode, error) {
	return nil, nil
}

func (emptyApplication) ListSwarmServices(context.Context, string) ([]domain.SwarmService, error) {
	return nil, nil
}

func (emptyApplication) GitStatus(context.Context, string) (domain.GitStatus, error) {
	return domain.GitStatus{}, nil
}

func (emptyApplication) PreviewGit(
	context.Context,
	service.CreateGitOperationInput,
) (service.GitPreview, error) {
	return service.GitPreview{}, nil
}

func (emptyApplication) ApproveGit(context.Context, string) (domain.GitOperation, error) {
	return domain.GitOperation{}, nil
}

func (emptyApplication) ExecuteGit(context.Context, string) (domain.GitOperation, error) {
	return domain.GitOperation{}, nil
}

func (emptyApplication) ListGitOperations(context.Context, int) ([]domain.GitOperation, error) {
	return nil, nil
}

func (emptyApplication) ListAIProviders(context.Context) ([]domain.AIProvider, error) {
	return nil, nil
}

func (emptyApplication) ListAIUsage(context.Context, int) ([]domain.AIUsage, error) {
	return nil, nil
}

func (emptyApplication) GetPolicy(context.Context) (domain.Policy, error) {
	return domain.DefaultPolicy(time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)), nil
}

func (emptyApplication) ListAudit(context.Context, int) ([]domain.AuditEntry, error) {
	return nil, nil
}

func (emptyApplication) ListWorkers(context.Context) ([]domain.Worker, error) { return nil, nil }

func (emptyApplication) ListPlugins(context.Context) ([]domain.Plugin, error) { return nil, nil }
