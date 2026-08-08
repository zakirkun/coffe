package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
)

func (a *App) ListSwarmNodes(ctx context.Context, rawTargetID string) ([]domain.SwarmNode, error) {
	targetID, err := parseTargetID(rawTargetID)
	if err != nil {
		return nil, err
	}
	target, err := a.repository.GetTarget(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("get swarm target: %w", err)
	}
	nodes, err := a.swarm.ListNodes(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("list swarm nodes: %w", err)
	}
	return nodes, nil
}

func (a *App) ListSwarmServices(ctx context.Context, rawTargetID string) ([]domain.SwarmService, error) {
	targetID, err := parseTargetID(rawTargetID)
	if err != nil {
		return nil, err
	}
	target, err := a.repository.GetTarget(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("get swarm target: %w", err)
	}
	project, err := a.repository.GetProject(ctx, target.ProjectID())
	if err != nil {
		return nil, fmt.Errorf("get swarm project: %w", err)
	}
	workflow, err := a.workflows.Load(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("load swarm workflow: %w", err)
	}
	if workflow.Mode() != domain.DeploymentSwarm {
		return nil, fmt.Errorf("target is not configured for swarm: %w", ErrInvalidInput)
	}
	services, err := a.swarm.ListServices(ctx, target, workflow.StackName())
	if err != nil {
		return nil, fmt.Errorf("list swarm services: %w", err)
	}
	return services, nil
}
