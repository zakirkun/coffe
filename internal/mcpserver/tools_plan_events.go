package mcpserver

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listPlanEventsInput struct {
	PlanID string `json:"plan_id" jsonschema:"deployment plan identifier"`
}

type planEventView struct {
	ID         string           `json:"id"`
	PlanID     string           `json:"plan_id"`
	Kind       domain.EventKind `json:"kind"`
	Message    string           `json:"message"`
	OccurredAt time.Time        `json:"occurred_at"`
}

type listPlanEventsOutput struct {
	Events []planEventView `json:"events"`
}

func listPlanEventsHandler(app Application) mcp.ToolHandlerFor[listPlanEventsInput, listPlanEventsOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input listPlanEventsInput,
	) (*mcp.CallToolResult, listPlanEventsOutput, error) {
		events, err := app.ListEvents(ctx, input.PlanID)
		if err != nil {
			return nil, listPlanEventsOutput{}, fmt.Errorf("list plan events: %w", err)
		}
		views := make([]planEventView, 0, len(events))
		for _, event := range events {
			views = append(views, planEventView{
				ID: string(event.ID()), PlanID: string(event.PlanID()), Kind: event.Kind(),
				Message: event.Message(), OccurredAt: event.OccurredAt(),
			})
		}
		return nil, listPlanEventsOutput{Events: views}, nil
	}
}
