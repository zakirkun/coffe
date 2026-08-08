package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
)

type CreateProjectInput struct {
	Name     string
	RootPath string
}

func (a *App) CreateProject(ctx context.Context, input CreateProjectInput) (domain.Project, error) {
	name, err := domain.NewName(input.Name)
	if err != nil {
		return domain.Project{}, fmt.Errorf("project name: %w", err)
	}
	rawID, err := a.ids.New()
	if err != nil {
		return domain.Project{}, fmt.Errorf("new project id: %w", err)
	}
	project, err := domain.NewProject(domain.ProjectValues{
		ID: domain.ProjectID(rawID), Name: name, RootPath: input.RootPath, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return domain.Project{}, fmt.Errorf("new project: %w", err)
	}
	if err := a.repository.CreateProject(ctx, project); err != nil {
		return domain.Project{}, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

type CreateTargetInput struct {
	ProjectID  string
	Name       string
	Host       string
	Port       int
	User       string
	RemotePath string
	Branch     string
}

func (a *App) CreateTarget(ctx context.Context, input CreateTargetInput) (domain.Target, error) {
	projectID, err := parseProjectID(input.ProjectID)
	if err != nil {
		return domain.Target{}, err
	}
	if _, err := a.repository.GetProject(ctx, projectID); err != nil {
		return domain.Target{}, fmt.Errorf("get target project: %w", err)
	}
	rawID, err := a.ids.New()
	if err != nil {
		return domain.Target{}, fmt.Errorf("new target id: %w", err)
	}
	name, err := domain.NewName(input.Name)
	if err != nil {
		return domain.Target{}, fmt.Errorf("target name: %w", err)
	}
	host, err := domain.NewHost(input.Host)
	if err != nil {
		return domain.Target{}, fmt.Errorf("target host: %w", err)
	}
	port, err := domain.NewPort(input.Port)
	if err != nil {
		return domain.Target{}, fmt.Errorf("target port: %w", err)
	}
	user, err := domain.NewSSHUser(input.User)
	if err != nil {
		return domain.Target{}, fmt.Errorf("target user: %w", err)
	}
	remotePath, err := domain.NewRemotePath(input.RemotePath)
	if err != nil {
		return domain.Target{}, fmt.Errorf("target remote path: %w", err)
	}
	branch, err := domain.NewBranch(input.Branch)
	if err != nil {
		return domain.Target{}, fmt.Errorf("target branch: %w", err)
	}
	target, err := domain.NewTarget(domain.TargetValues{
		ID: domain.TargetID(rawID), ProjectID: projectID, Name: name, Host: host, Port: port,
		User: user, RemotePath: remotePath, Branch: branch, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return domain.Target{}, fmt.Errorf("new target: %w", err)
	}
	if err := a.repository.CreateTarget(ctx, target); err != nil {
		return domain.Target{}, fmt.Errorf("create target: %w", err)
	}
	return target, nil
}
