package domain_test

import (
	"coffe/internal/domain"
	"testing"
	"time"
)

func Test_AgentSession_lifecycle_starts_and_stops(t *testing.T) {
	// Given
	createdAt := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	session, err := domain.NewAgentSession(domain.AgentSessionValues{
		ID: "agent-1", ProjectID: "project-1", Tool: domain.AgentOpenCode,
		Status: domain.AgentPending, CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("NewAgentSession() error = %v", err)
	}

	// When
	running, err := session.Start(createdAt.Add(time.Second))
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	stopped, err := running.Stop(createdAt.Add(2 * time.Second))
	// Then
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if stopped.Status() != domain.AgentStopped {
		t.Fatalf("Stop() status = %q, want stopped", stopped.Status())
	}
}
