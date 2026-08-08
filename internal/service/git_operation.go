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
	output, runErr := a.git.Execute(ctx, project, running)
	if clean := redact.Line(output); clean != "" {
		if err := a.appendGitEvent(ctx, running.ID(), clean); err != nil {
			return domain.GitOperation{}, err
		}
	}
	finished, err := running.Finish(runErr == nil)
	if err != nil {
		return domain.GitOperation{}, err
	}
	if err := a.repository.SaveGitOperation(ctx, finished); err != nil {
		return domain.GitOperation{}, fmt.Errorf("save finished git operation: %w", err)
	}
	if runErr != nil {
		auditErr := a.appendAudit(ctx, auditInput{
			category: domain.AuditGit, action: "execute", resourceID: string(finished.ID()),
			outcome: domain.AuditFailure, message: "Git operation failed",
		})
		if err := a.appendGitEvent(ctx, finished.ID(), "git operation failed"); err != nil {
			return finished, errors.Join(runErr, err, auditErr)
		}
		return finished, errors.Join(runErr, auditErr)
	}
	if err := a.appendGitEvent(ctx, finished.ID(), "git operation succeeded"); err != nil {
		return domain.GitOperation{}, err
	}
	if err := a.appendAudit(ctx, auditInput{
		category: domain.AuditGit, action: "execute", resourceID: string(finished.ID()),
		outcome: domain.AuditSuccess, message: "Git operation succeeded",
	}); err != nil {
		return domain.GitOperation{}, err
	}
	return finished, nil
}
