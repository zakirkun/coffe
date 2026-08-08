package mcpserver

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type gitProjectInput struct {
	ProjectID string `json:"project_id" jsonschema:"registered project identifier"`
}

type (
	gitStatusOutput struct {
		Status domain.GitStatus `json:"status"`
	}
	gitDiffOutput struct {
		Diff string `json:"diff"`
	}
)

type gitPreviewInput struct {
	ProjectID string `json:"project_id" jsonschema:"registered project identifier"`
	Kind      string `json:"kind" jsonschema:"pull or push"`
	Remote    string `json:"remote" jsonschema:"configured Git remote name"`
	Branch    string `json:"branch" jsonschema:"branch to pull or push"`
}

type gitOperationOutput struct {
	ID        string                    `json:"id"`
	ProjectID string                    `json:"project_id"`
	Kind      domain.GitOperationKind   `json:"kind"`
	Status    domain.GitOperationStatus `json:"status"`
	Command   string                    `json:"command,omitempty"`
}

type (
	gitPreviewOutput struct {
		Operation gitOperationOutput `json:"operation"`
	}
	listGitOperationsInput  struct{}
	listGitOperationsOutput struct {
		Operations []gitOperationOutput `json:"operations"`
	}
)

func registerGitTools(server *mcp.Server, app Application) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "git_status", Description: "Read branch and working-tree status for one project",
	}, gitStatusHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "git_diff", Description: "Read a capped unstaged diff for one project",
	}, gitDiffHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "preview_git_operation", Description: "Create a pull or push preview requiring human approval",
	}, previewGitHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_git_operations", Description: "List recent Git operation approvals and outcomes",
	}, listGitOperationsHandler(app))
}

func gitStatusHandler(app Application) mcp.ToolHandlerFor[gitProjectInput, gitStatusOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input gitProjectInput,
	) (*mcp.CallToolResult, gitStatusOutput, error) {
		status, err := app.GitStatus(ctx, input.ProjectID)
		if err != nil {
			return nil, gitStatusOutput{}, fmt.Errorf("git status: %w", err)
		}
		return nil, gitStatusOutput{Status: status}, nil
	}
}

func gitDiffHandler(app Application) mcp.ToolHandlerFor[gitProjectInput, gitDiffOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input gitProjectInput,
	) (*mcp.CallToolResult, gitDiffOutput, error) {
		diff, err := app.GitDiff(ctx, input.ProjectID)
		if err != nil {
			return nil, gitDiffOutput{}, fmt.Errorf("git diff: %w", err)
		}
		return nil, gitDiffOutput{Diff: diff}, nil
	}
}

func previewGitHandler(app Application) mcp.ToolHandlerFor[gitPreviewInput, gitPreviewOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input gitPreviewInput,
	) (*mcp.CallToolResult, gitPreviewOutput, error) {
		preview, err := app.PreviewGit(ctx, service.CreateGitOperationInput(input))
		if err != nil {
			return nil, gitPreviewOutput{}, fmt.Errorf("preview git: %w", err)
		}
		return nil, gitPreviewOutput{Operation: gitOperationOutput{
			ID: string(preview.Operation.ID()), ProjectID: string(preview.Operation.ProjectID()),
			Kind: preview.Operation.Kind(), Status: preview.Operation.Status(), Command: preview.Command,
		}}, nil
	}
}

func listGitOperationsHandler(app Application) mcp.ToolHandlerFor[listGitOperationsInput, listGitOperationsOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ listGitOperationsInput,
	) (*mcp.CallToolResult, listGitOperationsOutput, error) {
		operations, err := app.ListGitOperations(ctx, 50)
		if err != nil {
			return nil, listGitOperationsOutput{}, fmt.Errorf("list git operations: %w", err)
		}
		output := make([]gitOperationOutput, 0, len(operations))
		for _, operation := range operations {
			output = append(output, gitOperationOutput{
				ID: string(operation.ID()), ProjectID: string(operation.ProjectID()),
				Kind: operation.Kind(), Status: operation.Status(),
			})
		}
		return nil, listGitOperationsOutput{Operations: output}, nil
	}
}
