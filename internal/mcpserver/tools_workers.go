package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	listWorkersInput struct{}
	workerOutput     struct {
		ID       string    `json:"id"`
		TargetID string    `json:"target_id"`
		Name     string    `json:"name"`
		BaseURL  string    `json:"base_url"`
		Version  string    `json:"version,omitempty"`
		Enabled  bool      `json:"enabled"`
		LastSeen time.Time `json:"last_seen,omitempty"`
	}
)

type workersOutput struct {
	Workers []workerOutput `json:"workers"`
}

func registerWorkerTools(server *mcp.Server, app Application) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_workers", Description: "List registered remote workers and heartbeat state",
	}, listWorkersHandler(app))
}

func listWorkersHandler(app Application) mcp.ToolHandlerFor[listWorkersInput, workersOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ listWorkersInput,
	) (*mcp.CallToolResult, workersOutput, error) {
		workers, err := app.ListWorkers(ctx)
		if err != nil {
			return nil, workersOutput{}, fmt.Errorf("list workers: %w", err)
		}
		output := make([]workerOutput, 0, len(workers))
		for _, worker := range workers {
			output = append(output, workerOutput{
				ID: string(worker.ID()), TargetID: string(worker.TargetID()), Name: worker.Name().String(),
				BaseURL: worker.BaseURL(), Version: worker.Version(),
				Enabled: worker.Enabled(), LastSeen: worker.LastSeen(),
			})
		}
		return nil, workersOutput{Workers: output}, nil
	}
}
