package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
)

type ProjectGitStatus struct {
	Project domain.Project   `json:"project"`
	Status  domain.GitStatus `json:"status"`
	Error   string           `json:"error,omitempty"`
}

func (a *App) GitStatus(ctx context.Context, rawProjectID string) (domain.GitStatus, error) {
	project, err := a.gitProject(ctx, rawProjectID)
	if err != nil {
		return domain.GitStatus{}, err
	}
	status, err := a.git.Status(ctx, project)
	if err != nil {
		return domain.GitStatus{}, fmt.Errorf("read git status: %w", err)
	}
	return status, nil
}

func (a *App) GitBranches(ctx context.Context, rawProjectID string) ([]string, error) {
	project, err := a.gitProject(ctx, rawProjectID)
	if err != nil {
		return nil, err
	}
	branches, err := a.git.Branches(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("read git branches: %w", err)
	}
	return branches, nil
}

func (a *App) GitDiff(ctx context.Context, rawProjectID string) (string, error) {
	project, err := a.gitProject(ctx, rawProjectID)
	if err != nil {
		return "", err
	}
	diff, err := a.git.Diff(ctx, project)
	if err != nil {
		return "", fmt.Errorf("read git diff: %w", err)
	}
	return diff, nil
}

func (a *App) AllGitStatuses(ctx context.Context) ([]ProjectGitStatus, error) {
	projects, err := a.repository.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list git projects: %w", err)
	}
	statuses := make([]ProjectGitStatus, 0, len(projects))
	for _, project := range projects {
		status, err := a.git.Status(ctx, project)
		if err != nil {
			statuses = append(statuses, ProjectGitStatus{Project: project, Error: err.Error()})
			continue
		}
		statuses = append(statuses, ProjectGitStatus{Project: project, Status: status})
	}
	return statuses, nil
}

func (a *App) gitProject(ctx context.Context, rawID string) (domain.Project, error) {
	id, err := parseProjectID(rawID)
	if err != nil {
		return domain.Project{}, err
	}
	project, err := a.repository.GetProject(ctx, id)
	if err != nil {
		return domain.Project{}, fmt.Errorf("get git project: %w", err)
	}
	return project, nil
}
