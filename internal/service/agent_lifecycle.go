package service

import (
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"
)

func (a *App) finishAgent(ctx context.Context, session domain.AgentSession, cause error) error {
	storageCtx := context.WithoutCancel(ctx)
	var finished domain.AgentSession
	var err error
	if errors.Is(cause, context.Canceled) {
		finished, err = session.Stop(a.clock.Now())
	} else {
		finished, err = session.Finish(cause == nil, a.clock.Now())
	}
	if err != nil {
		return err
	}
	if err := a.repository.SaveAgentSession(storageCtx, finished); err != nil {
		return fmt.Errorf("save finished agent: %w", err)
	}
	message := "agent succeeded"
	if finished.Status() == domain.AgentFailed {
		message = "agent failed"
	} else if finished.Status() == domain.AgentStopped {
		message = "agent stopped"
	}
	return a.appendAgentEvent(storageCtx, session.ID(), message)
}

func (a *App) StopAgent(ctx context.Context, rawID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("stop agent: %w", err)
	}
	id := domain.AgentSessionID(strings.TrimSpace(rawID))
	if id == "" || string(id) != rawID {
		return fmt.Errorf("agent session id: %w", ErrInvalidInput)
	}
	a.runMu.Lock()
	cancel, ok := a.agentRuns[id]
	a.runMu.Unlock()
	if !ok {
		return fmt.Errorf("agent session %s: %w", id, ErrNotRunning)
	}
	cancel()
	return nil
}

func (a *App) ListAgentSessions(ctx context.Context, limit int) ([]domain.AgentSession, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("agent session limit: %w", ErrInvalidInput)
	}
	sessions, err := a.repository.ListAgentSessions(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list agent sessions: %w", err)
	}
	return sessions, nil
}

func (a *App) ListAgentEvents(ctx context.Context, rawID string) ([]domain.AgentEvent, error) {
	id := domain.AgentSessionID(strings.TrimSpace(rawID))
	if id == "" || string(id) != rawID {
		return nil, fmt.Errorf("agent session id: %w", ErrInvalidInput)
	}
	events, err := a.repository.ListAgentEvents(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list agent events: %w", err)
	}
	return events, nil
}

func (a *App) appendAgentEvent(
	ctx context.Context,
	sessionID domain.AgentSessionID,
	message string,
) error {
	event, err := domain.NewAgentEvent(sessionID, message, a.clock.Now())
	if err != nil {
		return err
	}
	if err := a.repository.AppendAgentEvent(ctx, event); err != nil {
		return fmt.Errorf("append agent event: %w", err)
	}
	return nil
}

func (a *App) RecoverAgentSessions(ctx context.Context) error {
	sessions, err := a.repository.ListAgentSessions(ctx, 100)
	if err != nil {
		return fmt.Errorf("list interrupted agents: %w", err)
	}
	for _, session := range sessions {
		if session.Status() != domain.AgentRunning {
			continue
		}
		failed, err := session.Finish(false, a.clock.Now())
		if err != nil {
			return err
		}
		if err := a.repository.SaveAgentSession(ctx, failed); err != nil {
			return fmt.Errorf("save recovered agent: %w", err)
		}
		if err := a.appendAgentEvent(ctx, session.ID(), "interrupted agent recovered"); err != nil {
			return err
		}
	}
	return nil
}
