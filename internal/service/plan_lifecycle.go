package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
)

func (a *App) RetryPlan(ctx context.Context, rawID string) (domain.ActionPlan, error) {
	id, err := parsePlanID(rawID)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	plan, err := a.repository.GetPlan(ctx, id)
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("get plan for retry: %w", err)
	}
	rawRetryID, err := a.ids.New()
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("new retry plan id: %w", err)
	}
	retry, err := plan.Retry(domain.PlanID(rawRetryID), a.clock.Now())
	if err != nil {
		return domain.ActionPlan{}, err
	}
	if err := a.repository.SavePlan(ctx, retry); err != nil {
		return domain.ActionPlan{}, fmt.Errorf("save retry plan: %w", err)
	}
	if err := a.appendEvent(ctx, retry.ID(), domain.EventPlanRetried, "retry preview created"); err != nil {
		return domain.ActionPlan{}, err
	}
	return retry, nil
}

func (a *App) PreviewRollback(ctx context.Context, rawTargetID string) (domain.ActionPlan, error) {
	targetID, err := parseTargetID(rawTargetID)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	target, err := a.repository.GetTarget(ctx, targetID)
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("get rollback target: %w", err)
	}
	project, err := a.repository.GetProject(ctx, target.ProjectID())
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("get rollback project: %w", err)
	}
	workflow, err := a.workflows.Load(ctx, project)
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("load rollback workflow: %w", err)
	}
	if workflow.Mode() != domain.DeploymentCompose {
		return domain.ActionPlan{}, fmt.Errorf("swarm rollback: %w", ErrInvalidInput)
	}
	latest, err := a.repository.LatestDeployment(ctx, targetID)
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("latest deployment: %w", err)
	}
	rawID, err := a.ids.New()
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("new rollback plan id: %w", err)
	}
	steps := []domain.StepKind{
		domain.StepInspectRepository, domain.StepGitReset, domain.StepCaptureRevision,
		domain.StepComposeConfig,
	}
	if workflow.Build() {
		steps = append(steps, domain.StepComposeBuild)
	}
	steps = append(steps, domain.StepComposeUp, domain.StepComposeHealth)
	plan, err := domain.NewActionPlan(domain.PlanDraftValues{
		ID: domain.PlanID(rawID), ProjectID: project.ID(), TargetID: target.ID(),
		SourceID: latest.PlanID(), Kind: domain.PlanComposeRollback,
		Revision: latest.RevisionBefore(), Steps: steps, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("new rollback plan: %w", err)
	}
	if err := a.repository.SavePlan(ctx, plan); err != nil {
		return domain.ActionPlan{}, fmt.Errorf("save rollback plan: %w", err)
	}
	if err := a.appendEvent(ctx, plan.ID(), domain.EventPlanCreated, "rollback preview created"); err != nil {
		return domain.ActionPlan{}, err
	}
	return plan, nil
}

func (a *App) RecoverInterruptedPlans(ctx context.Context) error {
	plans, err := a.repository.ListPlans(ctx, 100)
	if err != nil {
		return fmt.Errorf("list interrupted plans: %w", err)
	}
	for _, plan := range plans {
		if plan.Status() != domain.PlanRunning {
			continue
		}
		failed, err := plan.Finish(false)
		if err != nil {
			return err
		}
		if err := a.repository.SavePlan(ctx, failed); err != nil {
			return fmt.Errorf("save recovered plan: %w", err)
		}
		if err := a.appendEvent(ctx, plan.ID(), domain.EventPlanRecovered, "interrupted plan recovered"); err != nil {
			return err
		}
	}
	return nil
}
