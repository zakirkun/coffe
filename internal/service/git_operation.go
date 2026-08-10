package service

import (
	"coffe/internal/domain"
	"coffe/internal/redact"
	"context"
	"errors"
	"fmt"
	"strings"
)

type CreateGitOperationInput struct {
	ProjectID string
	Kind      string
	Remote    string
	Branch    string
}

type GitPreview struct {
	Operation domain.GitOperation
	Command   string
}

func (a *App) PreviewGit(
	ctx context.Context,
	input CreateGitOperationInput,
) (GitPreview, error) {
	project, err := a.gitProject(ctx, input.ProjectID)
	if err != nil {
		return GitPreview{}, err
	}
	kind := domain.GitOperationKind(strings.TrimSpace(input.Kind))
	if err := a.policyAllowsGit(ctx, kind); err != nil {
		auditErr := a.appendAudit(ctx, auditInput{
			category: domain.AuditGit, action: "preview", resourceID: string(project.ID()),
			outcome: domain.AuditDenied, message: "Git operation denied by policy",
		})
		return GitPreview{}, errors.Join(err, auditErr)
	}
	remote, err := domain.NewGitRemote(input.Remote)
	if err != nil {
		return GitPreview{}, err
	}
	branch, err := domain.NewBranch(input.Branch)
	if err != nil {
		return GitPreview{}, err
	}
	rawID, err := a.ids.New()
	if err != nil {
		return GitPreview{}, fmt.Errorf("new git operation id: %w", err)
	}
	operation, err := domain.NewGitOperation(domain.GitOperationValues{
		ID: domain.GitOperationID(rawID), ProjectID: project.ID(), Kind: kind,
		Remote: remote, Branch: branch, Status: domain.GitDraft, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return GitPreview{}, err
	}
	if err := a.repository.SaveGitOperation(ctx, operation); err != nil {
		return GitPreview{}, fmt.Errorf("save git preview: %w", err)
	}
	if err := a.appendGitEvent(ctx, operation.ID(), "git preview created"); err != nil {
		return GitPreview{}, err
	}
	if err := a.appendAudit(ctx, auditInput{
		category: domain.AuditGit, action: "preview", resourceID: string(operation.ID()),
		outcome: domain.AuditSuccess, message: "Git operation preview created",
	}); err != nil {
		return GitPreview{}, err
	}
	return GitPreview{Operation: operation, Command: a.git.Render(operation)}, nil
}

func (a *App) ApproveGit(ctx context.Context, rawID string) (domain.GitOperation, error) {
	operation, err := a.getGitOperation(ctx, rawID)
	if err != nil {
		return domain.GitOperation{}, err
	}
	approved, err := operation.Approve(a.clock.Now())
	if err != nil {
		return domain.GitOperation{}, err
	}
	if err := a.repository.SaveGitOperation(ctx, approved); err != nil {
		return domain.GitOperation{}, fmt.Errorf("save approved git operation: %w", err)
	}
	if err := a.appendGitEvent(ctx, approved.ID(), "git approval recorded"); err != nil {
		return domain.GitOperation{}, err
	}
	if err := a.appendAudit(ctx, auditInput{
		category: domain.AuditGit, action: "approve", resourceID: string(approved.ID()),
		outcome: domain.AuditSuccess, message: "Git operation approved",
	}); err != nil {
		return domain.GitOperation{}, err
	}
	return approved, nil
}

func (a *App) ExecuteGit(ctx context.Context, rawID string) (domain.GitOperation, error) {
	operation, err := a.getGitOperation(ctx, rawID)
	if err != nil {
		return domain.GitOperation{}, err
	}
	project, err := a.repository.GetProject(ctx, operation.ProjectID())
	if err != nil {
		return domain.GitOperation{}, fmt.Errorf("get git operation project: %w", err)
	}
	running, err := operation.Start()
	if err != nil {
		return domain.GitOperation{}, err
	}
	if err := a.repository.SaveGitOperation(ctx, running); err != nil {
		return domain.GitOperation{}, fmt.Errorf("save running git operation: %w", err)
	}
	persistenceCtx := context.WithoutCancel(ctx)
	var output string
	runErr := a.policyAllowsGit(ctx, running.Kind())
	if runErr == nil {
		output, runErr = a.git.Execute(ctx, project, running)
	}
	var outputEventErr error
	if clean := redact.Line(output); clean != "" {
		outputEventErr = a.appendGitEvent(persistenceCtx, running.ID(), clean)
	}
	finished, err := running.Finish(runErr == nil)
	if err != nil {
		return domain.GitOperation{}, errors.Join(outputEventErr, err)
	}
	if err := a.repository.SaveGitOperation(persistenceCtx, finished); err != nil {
		return domain.GitOperation{}, errors.Join(
			outputEventErr, fmt.Errorf("save finished git operation: %w", err),
		)
	}
	if runErr != nil {
		auditErr := a.appendAudit(persistenceCtx, auditInput{
			category: domain.AuditGit, action: "execute", resourceID: string(finished.ID()),
			outcome: domain.AuditFailure, message: "Git operation failed",
		})
		eventErr := a.appendGitEvent(persistenceCtx, finished.ID(), "git operation failed")
		return finished, errors.Join(runErr, outputEventErr, eventErr, auditErr)
	}
	eventErr := a.appendGitEvent(persistenceCtx, finished.ID(), "git operation succeeded")
	auditErr := a.appendAudit(persistenceCtx, auditInput{
		category: domain.AuditGit, action: "execute", resourceID: string(finished.ID()),
		outcome: domain.AuditSuccess, message: "Git operation succeeded",
	})
	return finished, errors.Join(outputEventErr, eventErr, auditErr)
}
