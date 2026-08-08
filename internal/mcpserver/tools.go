package mcpserver

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type projectView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RootPath string `json:"root_path"`
}

type targetView struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	RemotePath string `json:"remote_path"`
	Branch     string `json:"branch"`
}

type planView struct {
	ID        string            `json:"id"`
	ProjectID string            `json:"project_id"`
	TargetID  string            `json:"target_id"`
	SourceID  string            `json:"source_plan_id,omitempty"`
	Kind      domain.PlanKind   `json:"kind"`
	Status    domain.PlanStatus `json:"status"`
	Revision  string            `json:"revision,omitempty"`
	Steps     []domain.StepKind `json:"steps"`
	Commands  []string          `json:"commands,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type (
	listProjectsInput  struct{}
	listProjectsOutput struct {
		Projects []projectView `json:"projects"`
	}
)

func listProjectsHandler(app Application) mcp.ToolHandlerFor[listProjectsInput, listProjectsOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ listProjectsInput,
	) (*mcp.CallToolResult, listProjectsOutput, error) {
		projects, err := app.ListProjects(ctx)
		if err != nil {
			return nil, listProjectsOutput{}, fmt.Errorf("list projects: %w", err)
		}
		views := make([]projectView, 0, len(projects))
		for _, project := range projects {
			views = append(views, projectView{
				ID: string(project.ID()), Name: project.Name().String(), RootPath: project.RootPath(),
			})
		}
		return nil, listProjectsOutput{Projects: views}, nil
	}
}

type listTargetsInput struct {
	ProjectID string `json:"project_id" jsonschema:"registered project identifier"`
}

type listTargetsOutput struct {
	Targets []targetView `json:"targets"`
}

func listTargetsHandler(app Application) mcp.ToolHandlerFor[listTargetsInput, listTargetsOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input listTargetsInput,
	) (*mcp.CallToolResult, listTargetsOutput, error) {
		targets, err := app.ListTargets(ctx, input.ProjectID)
		if err != nil {
			return nil, listTargetsOutput{}, fmt.Errorf("list targets: %w", err)
		}
		views := make([]targetView, 0, len(targets))
		for _, target := range targets {
			views = append(views, targetView{
				ID: string(target.ID()), Name: target.Name().String(), Host: target.Host().String(),
				RemotePath: target.RemotePath().String(), Branch: target.Branch().String(),
			})
		}
		return nil, listTargetsOutput{Targets: views}, nil
	}
}

type previewDeployInput struct {
	ProjectID string `json:"project_id" jsonschema:"registered project identifier"`
	TargetID  string `json:"target_id" jsonschema:"registered deployment target identifier"`
	Prompt    string `json:"prompt" jsonschema:"deployment request to classify"`
}

type previewDeployOutput struct {
	Plan planView `json:"plan"`
}

func previewDeployHandler(app Application) mcp.ToolHandlerFor[previewDeployInput, previewDeployOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input previewDeployInput,
	) (*mcp.CallToolResult, previewDeployOutput, error) {
		plan, err := app.PreviewDeploy(ctx, service.PreviewInput{
			ProjectID: input.ProjectID, TargetID: input.TargetID, Prompt: input.Prompt,
		})
		if err != nil {
			return nil, previewDeployOutput{}, fmt.Errorf("preview deploy: %w", err)
		}
		preview, err := app.PlanPreview(ctx, string(plan.ID()))
		if err != nil {
			return nil, previewDeployOutput{}, fmt.Errorf("render preview: %w", err)
		}
		return nil, previewDeployOutput{Plan: toPlanView(preview.Plan, preview.Commands)}, nil
	}
}

type getPlanInput struct {
	PlanID string `json:"plan_id" jsonschema:"deployment plan identifier"`
}

type getPlanOutput struct {
	Plan planView `json:"plan"`
}

func getPlanHandler(app Application) mcp.ToolHandlerFor[getPlanInput, getPlanOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input getPlanInput,
	) (*mcp.CallToolResult, getPlanOutput, error) {
		plan, err := app.GetPlan(ctx, input.PlanID)
		if err != nil {
			return nil, getPlanOutput{}, fmt.Errorf("get plan: %w", err)
		}
		return nil, getPlanOutput{Plan: toPlanView(plan, nil)}, nil
	}
}

func toPlanView(plan domain.ActionPlan, commands []string) planView {
	result := planView{
		ID: string(plan.ID()), ProjectID: string(plan.ProjectID()), TargetID: string(plan.TargetID()),
		SourceID: string(plan.SourcePlanID()), Kind: plan.Kind(), Status: plan.Status(),
		Steps: plan.Steps(), Commands: commands, CreatedAt: plan.CreatedAt(),
	}
	if revision, ok := plan.Revision(); ok {
		result.Revision = revision.String()
	}
	return result
}
