package service

import (
	"coffe/internal/domain"
	"coffe/internal/redact"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type StartAgentInput struct {
	ProjectID string
	Tool      string
	Prompt    string
}

func (a *App) StartAgent(ctx context.Context, input StartAgentInput) (domain.AgentSession, error) {
	projectID, err := parseProjectID(input.ProjectID)
	if err != nil {
		return domain.AgentSession{}, err
	}
	project, err := a.repository.GetProject(ctx, projectID)
	if err != nil {
		return domain.AgentSession{}, fmt.Errorf("get agent project: %w", err)
	}
	tool := domain.AgentTool(strings.TrimSpace(input.Tool))
	if !tool.Valid() || strings.TrimSpace(input.Prompt) == "" {
		return domain.AgentSession{}, fmt.Errorf("agent tool or prompt: %w", ErrInvalidInput)
	}
	a.policyMu.Lock()
	defer a.policyMu.Unlock()
	if err := a.policyAllowsAgent(ctx, tool); err != nil {
		auditErr := a.appendAudit(ctx, auditInput{
			category: domain.AuditAgent, action: "start", resourceID: string(project.ID()),
			outcome: domain.AuditDenied, message: "coding agent denied by policy",
		})
		return domain.AgentSession{}, errors.Join(err, auditErr)
	}
	rawID, err := a.ids.New()
	if err != nil {
		return domain.AgentSession{}, fmt.Errorf("new agent session id: %w", err)
	}
	session, err := domain.NewAgentSession(domain.AgentSessionValues{
		ID: domain.AgentSessionID(rawID), ProjectID: project.ID(), Tool: tool,
		Status: domain.AgentPending, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return domain.AgentSession{}, err
	}
	running, err := session.Start(a.clock.Now())
	if err != nil {
		return domain.AgentSession{}, err
	}
	if err := a.repository.SaveAgentSession(ctx, running); err != nil {
		return domain.AgentSession{}, fmt.Errorf("save running agent: %w", err)
	}
	if err := a.appendAgentEvent(ctx, running.ID(), "agent started"); err != nil {
		return domain.AgentSession{}, err
	}
	if err := a.appendAudit(ctx, auditInput{
		category: domain.AuditAgent, action: "start", resourceID: string(running.ID()),
		outcome: domain.AuditSuccess, message: "coding agent started",
	}); err != nil {
		a.logger.ErrorContext(ctx, "agent audit failed", slog.Any("err", err))
	}
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	a.runMu.Lock()
	a.agentRuns[running.ID()] = cancel
	a.runWG.Add(1)
	a.runMu.Unlock()
	go func() {
		defer cancel()
		a.runAgent(runCtx, running, project.RootPath(), input.Prompt)
	}()
	return running, nil
}

func (a *App) runAgent(
	ctx context.Context,
	session domain.AgentSession,
	workingDir string,
	prompt string,
) {
	defer a.runWG.Done()
	defer func() {
		a.runMu.Lock()
		delete(a.agentRuns, session.ID())
		a.runMu.Unlock()
	}()
	err := a.agents.Run(ctx, AgentRun{
		Tool: session.Tool(), WorkingDir: workingDir, Prompt: prompt,
		Output: func(line string) error {
			clean := redact.Line(line)
			if clean == "" {
				return nil
			}
			return a.appendAgentEvent(ctx, session.ID(), clean)
		},
	})
	if finishErr := a.finishAgent(ctx, session, err); finishErr != nil {
		a.logger.ErrorContext(
			context.WithoutCancel(ctx), "agent finalization failed",
			slog.String("session_id", string(session.ID())), slog.Any("err", finishErr),
		)
	}
}
