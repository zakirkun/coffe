package mcpserver_test

import (
	"coffe/internal/domain"
	"coffe/internal/mcpserver"
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Test_Server_registers_git_branches_and_list_plan_events(t *testing.T) {
	// Given
	session := connectMCPClient(t, emptyApplication{})

	// When
	result, err := session.ListTools(context.Background(), nil)

	// Then
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	names := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	if !names["git_branches"] || !names["list_plan_events"] {
		t.Fatalf("registered tools = %v", names)
	}
	for name := range names {
		if strings.Contains(name, "approve") || strings.Contains(name, "execute") {
			t.Fatalf("mutation tool %q must not be registered", name)
		}
	}
}

func Test_Read_tools_call_application_with_resource_identifiers(t *testing.T) {
	// Given
	app := &recordingApplication{}
	session := connectMCPClient(t, app)
	ctx := context.Background()

	// When
	branches, branchErr := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "git_branches", Arguments: map[string]any{"project_id": "project-1"},
	})
	events, eventErr := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_plan_events", Arguments: map[string]any{"plan_id": "plan-1"},
	})

	// Then
	if branchErr != nil || branches.IsError {
		t.Fatalf("git_branches error = %v, result = %#v", branchErr, branches)
	}
	if eventErr != nil || events.IsError {
		t.Fatalf("list_plan_events error = %v, result = %#v", eventErr, events)
	}
	if app.branchProjectID != "project-1" || app.eventPlanID != "plan-1" {
		t.Fatalf("application identifiers = %q/%q", app.branchProjectID, app.eventPlanID)
	}
}

type recordingApplication struct {
	emptyApplication
	branchProjectID string
	eventPlanID     string
}

func (a *recordingApplication) GitBranches(_ context.Context, projectID string) ([]string, error) {
	a.branchProjectID = projectID
	return []string{"main"}, nil
}

func (a *recordingApplication) ListEvents(_ context.Context, planID string) ([]domain.Event, error) {
	a.eventPlanID = planID
	return nil, nil
}

func connectMCPClient(t *testing.T, app mcpserver.Application) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := mcpserver.New(app).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server Connect() error = %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := clientSession.Close(); closeErr != nil {
			t.Errorf("client Close() error = %v", closeErr)
		}
		if closeErr := serverSession.Close(); closeErr != nil {
			t.Errorf("server Close() error = %v", closeErr)
		}
	})
	return clientSession
}
