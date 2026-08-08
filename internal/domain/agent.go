package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidAgentSession = errors.New("domain: invalid agent session")
	ErrAgentTransition     = errors.New("domain: invalid agent transition")
)

type (
	AgentSessionID string
	AgentTool      string
	AgentStatus    string
)

const (
	AgentOpenCode AgentTool = "opencode"
	AgentClaude   AgentTool = "claude"
	AgentCodex    AgentTool = "codex"
	AgentGemini   AgentTool = "gemini"
	AgentAider    AgentTool = "aider"
)

const (
	AgentPending   AgentStatus = "pending"
	AgentRunning   AgentStatus = "running"
	AgentSucceeded AgentStatus = "succeeded"
	AgentFailed    AgentStatus = "failed"
	AgentStopped   AgentStatus = "stopped"
)

func (t AgentTool) Valid() bool {
	switch t {
	case AgentOpenCode, AgentClaude, AgentCodex, AgentGemini, AgentAider:
		return true
	default:
		return false
	}
}

func (s AgentStatus) Valid() bool {
	switch s {
	case AgentPending, AgentRunning, AgentSucceeded, AgentFailed, AgentStopped:
		return true
	default:
		return false
	}
}

type AgentSession struct {
	createdAt  time.Time
	startedAt  time.Time
	finishedAt time.Time
	id         AgentSessionID
	projectID  ProjectID
	tool       AgentTool
	status     AgentStatus
}

type AgentSessionValues struct {
	ID         AgentSessionID
	ProjectID  ProjectID
	Tool       AgentTool
	Status     AgentStatus
	CreatedAt  time.Time
	StartedAt  time.Time
	FinishedAt time.Time
}

func NewAgentSession(values AgentSessionValues) (AgentSession, error) {
	if values.ID == "" || values.ProjectID == "" || !values.Tool.Valid() ||
		!values.Status.Valid() || values.CreatedAt.IsZero() {
		return AgentSession{}, fmt.Errorf("agent session: %w", ErrInvalidAgentSession)
	}
	if values.Status == AgentRunning && values.StartedAt.IsZero() {
		return AgentSession{}, fmt.Errorf("agent start time: %w", ErrInvalidAgentSession)
	}
	if (values.Status == AgentSucceeded || values.Status == AgentFailed || values.Status == AgentStopped) &&
		values.FinishedAt.IsZero() {
		return AgentSession{}, fmt.Errorf("agent finish time: %w", ErrInvalidAgentSession)
	}
	return AgentSession{
		id: values.ID, projectID: values.ProjectID, tool: values.Tool, status: values.Status,
		createdAt: values.CreatedAt.UTC(), startedAt: values.StartedAt.UTC(), finishedAt: values.FinishedAt.UTC(),
	}, nil
}

func (s AgentSession) Start(at time.Time) (AgentSession, error) {
	if s.status != AgentPending || at.IsZero() {
		return AgentSession{}, fmt.Errorf("start agent %s: %w", s.status, ErrAgentTransition)
	}
	s.status = AgentRunning
	s.startedAt = at.UTC()
	return s, nil
}

func (s AgentSession) Finish(succeeded bool, at time.Time) (AgentSession, error) {
	if s.status != AgentRunning || at.IsZero() {
		return AgentSession{}, fmt.Errorf("finish agent %s: %w", s.status, ErrAgentTransition)
	}
	if succeeded {
		s.status = AgentSucceeded
	} else {
		s.status = AgentFailed
	}
	s.finishedAt = at.UTC()
	return s, nil
}

func (s AgentSession) Stop(at time.Time) (AgentSession, error) {
	if s.status != AgentRunning || at.IsZero() {
		return AgentSession{}, fmt.Errorf("stop agent %s: %w", s.status, ErrAgentTransition)
	}
	s.status = AgentStopped
	s.finishedAt = at.UTC()
	return s, nil
}

func (s AgentSession) ID() AgentSessionID    { return s.id }
func (s AgentSession) ProjectID() ProjectID  { return s.projectID }
func (s AgentSession) Tool() AgentTool       { return s.tool }
func (s AgentSession) Status() AgentStatus   { return s.status }
func (s AgentSession) CreatedAt() time.Time  { return s.createdAt }
func (s AgentSession) StartedAt() time.Time  { return s.startedAt }
func (s AgentSession) FinishedAt() time.Time { return s.finishedAt }

type AgentEvent struct {
	occurredAt time.Time
	sessionID  AgentSessionID
	message    string
}

func NewAgentEvent(sessionID AgentSessionID, message string, occurredAt time.Time) (AgentEvent, error) {
	if sessionID == "" || message == "" || occurredAt.IsZero() || hasControl(message) {
		return AgentEvent{}, fmt.Errorf("agent event: %w", ErrInvalidAgentSession)
	}
	return AgentEvent{sessionID: sessionID, message: message, occurredAt: occurredAt.UTC()}, nil
}

func (e AgentEvent) SessionID() AgentSessionID { return e.sessionID }
func (e AgentEvent) Message() string           { return e.message }
func (e AgentEvent) OccurredAt() time.Time     { return e.occurredAt }
