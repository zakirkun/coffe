package mcpserver

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listAgentSessionsInput struct{}

type agentSessionOutput struct {
	ID        string             `json:"id"`
	ProjectID string             `json:"project_id"`
	Tool      domain.AgentTool   `json:"tool"`
	Status    domain.AgentStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}

type agentSessionsOutput struct {
	Sessions []agentSessionOutput `json:"sessions"`
}

type agentEventsInput struct {
	SessionID string `json:"session_id" jsonschema:"coding agent session identifier"`
}

type agentEventOutput struct {
	Message    string    `json:"message"`
	OccurredAt time.Time `json:"occurred_at"`
}

type agentEventsOutput struct {
	Events []agentEventOutput `json:"events"`
}

func registerAgentTools(server *mcp.Server, app Application) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_agent_sessions", Description: "List local coding-agent sessions without prompt contents",
	}, listAgentSessionsHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_agent_events", Description: "Read redacted output events for one coding-agent session",
	}, getAgentEventsHandler(app))
}

func listAgentSessionsHandler(app Application) mcp.ToolHandlerFor[listAgentSessionsInput, agentSessionsOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ listAgentSessionsInput,
	) (*mcp.CallToolResult, agentSessionsOutput, error) {
		sessions, err := app.ListAgentSessions(ctx, 50)
		if err != nil {
			return nil, agentSessionsOutput{}, fmt.Errorf("list agent sessions: %w", err)
		}
		output := make([]agentSessionOutput, 0, len(sessions))
		for _, session := range sessions {
			output = append(output, agentSessionOutput{
				ID: string(session.ID()), ProjectID: string(session.ProjectID()),
				Tool: session.Tool(), Status: session.Status(), CreatedAt: session.CreatedAt(),
			})
		}
		return nil, agentSessionsOutput{Sessions: output}, nil
	}
}

func getAgentEventsHandler(app Application) mcp.ToolHandlerFor[agentEventsInput, agentEventsOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input agentEventsInput,
	) (*mcp.CallToolResult, agentEventsOutput, error) {
		events, err := app.ListAgentEvents(ctx, input.SessionID)
		if err != nil {
			return nil, agentEventsOutput{}, fmt.Errorf("list agent events: %w", err)
		}
		output := make([]agentEventOutput, 0, len(events))
		for _, event := range events {
			output = append(output, agentEventOutput{
				Message: event.Message(), OccurredAt: event.OccurredAt(),
			})
		}
		return nil, agentEventsOutput{Events: output}, nil
	}
}
