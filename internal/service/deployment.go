package service

import (
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"
)

type PreviewInput struct {
	ProjectID string
	TargetID  string
	Prompt    string
}

type Preview struct {
	Plan     domain.ActionPlan
	Commands []string
}

func (a *App) PreviewDeploy(ctx context.Context, input PreviewInput) (domain.ActionPlan, error) {
	projectID, targetID, err := parsePreviewInput(input)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	intent, err := a.planner.Classify(ctx, input.Prompt)
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("classify prompt: %w", err)
	}
	if intent != domain.IntentDeploy {
		return domain.ActionPlan{}, fmt.Errorf("intent %q: %w", intent, ErrUnsupportedIntent)
	}
	project, target, workflow, err := a.loadDeployment(ctx, projectID, targetID)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	steps := []domain.StepKind{
		domain.StepInspectRepository, domain.StepGitPull, domain.StepCaptureRevision,
	}
	kind := domain.PlanComposeDeploy
	if workflow.Mode() == domain.DeploymentSwarm {
		if err := a.policyAllowsSwarm(ctx); err != nil {
			auditErr := a.appendAudit(ctx, auditInput{
				category: domain.AuditDeploy, action: "swarm_preview", resourceID: string(target.ID()),
				outcome: domain.AuditDenied, message: "Swarm deployment denied by policy",
			})
			return domain.ActionPlan{}, errors.Join(err, auditErr)
		}
		kind = domain.PlanSwarmDeploy
		steps = append(steps, domain.StepSwarmConfig, domain.StepSwarmDeploy, domain.StepSwarmHealth)
	} else {
		steps = append(steps, domain.StepComposeConfig)
		if workflow.Build() {
			steps = append(steps, domain.StepComposeBuild)
		}
		steps = append(steps, domain.StepComposeUp, domain.StepComposeHealth)
	}
	rawID, err := a.ids.New()
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("new plan id: %w", err)
	}
	plan, err := domain.NewActionPlan(domain.PlanDraftValues{
		ID: domain.PlanID(rawID), ProjectID: project.ID(), TargetID: target.ID(), Kind: kind,
		Steps: steps, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("new plan: %w", err)
	}
	if err := a.repository.SavePlan(ctx, plan); err != nil {
		return domain.ActionPlan{}, fmt.Errorf("save plan: %w", err)
	}
	if err := a.appendEvent(ctx, plan.ID(), domain.EventPlanCreated, "deployment preview created"); err != nil {
		return domain.ActionPlan{}, err
	}
	return plan, nil
}

func (a *App) ApprovePlan(ctx context.Context, rawID string) (domain.ActionPlan, error) {
	id, err := parsePlanID(rawID)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	plan, err := a.repository.GetPlan(ctx, id)
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("get plan for approval: %w", err)
	}
	approved, err := plan.Approve(a.clock.Now())
	if err != nil {
		return domain.ActionPlan{}, err
	}
	if err := a.repository.SavePlan(ctx, approved); err != nil {
		return domain.ActionPlan{}, fmt.Errorf("save approved plan: %w", err)
	}
	if err := a.appendEvent(ctx, id, domain.EventPlanApproved, "explicit approval recorded"); err != nil {
		return domain.ActionPlan{}, err
	}
	if err := a.appendAudit(ctx, auditInput{
		category: domain.AuditDeploy, action: "approve", resourceID: string(id),
		outcome: domain.AuditSuccess, message: "deployment plan approved",
	}); err != nil {
		return domain.ActionPlan{}, err
	}
	return approved, nil
}

func (a *App) PlanPreview(ctx context.Context, rawID string) (Preview, error) {
	id, err := parsePlanID(rawID)
	if err != nil {
		return Preview{}, err
	}
	plan, err := a.repository.GetPlan(ctx, id)
	if err != nil {
		return Preview{}, fmt.Errorf("get plan preview: %w", err)
	}
	project, target, workflow, err := a.loadDeployment(ctx, plan.ProjectID(), plan.TargetID())
	if err != nil {
		return Preview{}, err
	}
	if project.ID() != plan.ProjectID() {
		return Preview{}, fmt.Errorf("preview project: %w", ErrInvalidInput)
	}
	commands := make([]string, 0, len(plan.Steps()))
	for _, step := range plan.Steps() {
		command, err := a.renderer.Render(ExecutionStep{
			Plan: plan, Target: target, Workflow: workflow, Step: step,
		})
		if err != nil {
			return Preview{}, fmt.Errorf("render %s: %w", step, err)
		}
		commands = append(commands, command)
	}
	return Preview{Plan: plan, Commands: commands}, nil
}

func parsePreviewInput(input PreviewInput) (domain.ProjectID, domain.TargetID, error) {
	projectID, err := parseProjectID(input.ProjectID)
	if err != nil {
		return "", "", err
	}
	targetID, err := parseTargetID(input.TargetID)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(input.Prompt) == "" {
		return "", "", fmt.Errorf("prompt: %w", ErrInvalidInput)
	}
	return projectID, targetID, nil
}
