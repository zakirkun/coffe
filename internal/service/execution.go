package service

import (
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
)

func (a *App) loadDeployment(
	ctx context.Context,
	projectID domain.ProjectID,
	targetID domain.TargetID,
) (domain.Project, domain.Target, domain.Workflow, error) {
	project, err := a.repository.GetProject(ctx, projectID)
	if err != nil {
		return domain.Project{}, domain.Target{}, domain.Workflow{}, fmt.Errorf("get project: %w", err)
	}
	target, err := a.repository.GetTarget(ctx, targetID)
	if err != nil {
		return domain.Project{}, domain.Target{}, domain.Workflow{}, fmt.Errorf("get target: %w", err)
	}
	if target.ProjectID() != project.ID() {
		return domain.Project{}, domain.Target{}, domain.Workflow{}, fmt.Errorf("target project: %w", ErrInvalidInput)
	}
	workflow, err := a.workflows.Load(ctx, project)
	if err != nil {
		return domain.Project{}, domain.Target{}, domain.Workflow{}, fmt.Errorf("load workflow: %w", err)
	}
	return project, target, workflow, nil
}

func (a *App) startExecution(
	ctx context.Context,
	id domain.PlanID,
) (domain.ActionPlan, domain.Target, domain.Workflow, error) {
	a.startMu.Lock()
	defer a.startMu.Unlock()
	plan, err := a.repository.GetPlan(ctx, id)
	if err != nil {
		return domain.ActionPlan{}, domain.Target{}, domain.Workflow{}, fmt.Errorf("get plan for execution: %w", err)
	}
	project, target, workflow, err := a.loadDeployment(ctx, plan.ProjectID(), plan.TargetID())
	if err != nil {
		return domain.ActionPlan{}, domain.Target{}, domain.Workflow{}, err
	}
	if project.ID() != plan.ProjectID() {
		return domain.ActionPlan{}, domain.Target{}, domain.Workflow{}, fmt.Errorf("plan project: %w", ErrInvalidInput)
	}
	running, err := plan.Start()
	if err != nil {
		return domain.ActionPlan{}, domain.Target{}, domain.Workflow{}, err
	}
	if err := a.repository.SavePlan(ctx, running); err != nil {
		return domain.ActionPlan{}, domain.Target{}, domain.Workflow{}, fmt.Errorf("save running plan: %w", err)
	}
	if err := a.appendEvent(ctx, id, domain.EventPlanStarted, "deployment started"); err != nil {
		return domain.ActionPlan{}, domain.Target{}, domain.Workflow{}, err
	}
	return running, target, workflow, nil
}

func (a *App) failPlan(
	ctx context.Context,
	plan domain.ActionPlan,
	cause error,
) (domain.ActionPlan, error) {
	if errors.Is(cause, context.Canceled) {
		canceled, err := plan.Cancel()
		if err != nil {
			return domain.ActionPlan{}, errors.Join(cause, err)
		}
		if err := a.repository.SavePlan(context.WithoutCancel(ctx), canceled); err != nil {
			return domain.ActionPlan{}, errors.Join(cause, fmt.Errorf("save canceled plan: %w", err))
		}
		if err := a.appendEvent(
			context.WithoutCancel(ctx), plan.ID(), domain.EventPlanCanceled, "deployment canceled",
		); err != nil {
			return canceled, errors.Join(cause, err)
		}
		return canceled, cause
	}
	failed, err := plan.Finish(false)
	if err != nil {
		return domain.ActionPlan{}, errors.Join(cause, err)
	}
	if err := a.repository.SavePlan(ctx, failed); err != nil {
		return domain.ActionPlan{}, errors.Join(cause, fmt.Errorf("save failed plan: %w", err))
	}
	if err := a.appendEvent(ctx, plan.ID(), domain.EventPlanFailed, "deployment failed"); err != nil {
		return failed, errors.Join(cause, err)
	}
	return failed, cause
}

func (a *App) appendEvent(ctx context.Context, planID domain.PlanID, kind domain.EventKind, message string) error {
	rawID, err := a.ids.New()
	if err != nil {
		return fmt.Errorf("new event id: %w", err)
	}
	event, err := domain.NewEvent(domain.EventValues{
		ID: domain.EventID(rawID), PlanID: planID, Kind: kind, Message: message, OccurredAt: a.clock.Now(),
	})
	if err != nil {
		return fmt.Errorf("new event: %w", err)
	}
	if err := a.repository.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	return nil
}
