package service

import (
	"coffe/internal/domain"
	"coffe/internal/redact"
	"context"
	"fmt"
)

func (a *App) ExecutePlan(ctx context.Context, rawID string) (domain.ActionPlan, error) {
	id, err := parsePlanID(rawID)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	plan, target, workflow, err := a.startExecution(ctx, id)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	return a.runStartedPlan(ctx, plan, target, workflow)
}

func (a *App) runStartedPlan(
	ctx context.Context,
	plan domain.ActionPlan,
	target domain.Target,
	workflow domain.Workflow,
) (domain.ActionPlan, error) {
	var revisionBefore, revisionAfter domain.GitRevision
	for _, step := range plan.Steps() {
		if err := a.appendEvent(ctx, plan.ID(), domain.EventStepStarted, string(step)+" started"); err != nil {
			return a.failPlan(ctx, plan, err)
		}
		if step == domain.StepSwarmDeploy {
			if err := a.policyAllowsSwarm(ctx); err != nil {
				return a.failPlan(ctx, plan, err)
			}
		}
		result, err := a.executor.Execute(ctx, ExecutionStep{
			Plan: plan, Target: target, Workflow: workflow, Step: step,
			Output: func(line string) error {
				clean := redact.Line(line)
				if clean == "" {
					return nil
				}
				return a.appendEvent(ctx, plan.ID(), domain.EventStepOutput, clean)
			},
		})
		if err != nil {
			return a.failPlan(ctx, plan, fmt.Errorf("execute %s: %w", step, err))
		}
		switch step {
		case domain.StepInspectRepository:
			revisionBefore = result.Revision
		case domain.StepCaptureRevision:
			revisionAfter = result.Revision
		case domain.StepGitPull, domain.StepComposeConfig, domain.StepComposeBuild,
			domain.StepComposeUp, domain.StepComposeHealth, domain.StepGitReset,
			domain.StepSwarmConfig, domain.StepSwarmDeploy, domain.StepSwarmHealth:
		}
		if err := a.appendEvent(ctx, plan.ID(), domain.EventStepSucceeded, string(step)+" succeeded"); err != nil {
			return a.failPlan(ctx, plan, err)
		}
	}
	return a.completePlan(ctx, plan, target, revisionBefore, revisionAfter)
}

func (a *App) completePlan(
	ctx context.Context,
	plan domain.ActionPlan,
	target domain.Target,
	revisionBefore domain.GitRevision,
	revisionAfter domain.GitRevision,
) (domain.ActionPlan, error) {
	record, err := domain.NewDeploymentRecord(domain.DeploymentValues{
		PlanID: plan.ID(), TargetID: target.ID(), Kind: plan.Kind(),
		RevisionBefore: revisionBefore, RevisionAfter: revisionAfter,
		Health: domain.HealthHealthy, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return a.failPlan(ctx, plan, err)
	}
	if err := a.repository.SaveDeployment(ctx, record); err != nil {
		return a.failPlan(ctx, plan, fmt.Errorf("save deployment record: %w", err))
	}
	finished, err := plan.Finish(true)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	if err := a.repository.SavePlan(ctx, finished); err != nil {
		return domain.ActionPlan{}, fmt.Errorf("save succeeded plan: %w", err)
	}
	if err := a.appendEvent(ctx, plan.ID(), domain.EventPlanSucceeded, "deployment succeeded"); err != nil {
		return domain.ActionPlan{}, err
	}
	return finished, nil
}
