package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"strings"
	"testing"
	"time"
)

func Test_App_StartAgent_persists_lifecycle_without_prompt(t *testing.T) {
	// Given
	fixture := newFixture(t)
	prompt := "secret prompt payload"

	// When
	session, err := fixture.app.StartAgent(context.Background(), service.StartAgentInput{
		ProjectID: string(fixture.project.ID()), Tool: string(domain.AgentOpenCode), Prompt: prompt,
	})
	if err != nil {
		t.Fatalf("StartAgent() error = %v", err)
	}
	waitCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := fixture.app.Wait(waitCtx); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	stored, err := fixture.repo.GetAgentSession(context.Background(), session.ID())
	if err != nil {
		t.Fatalf("GetAgentSession() error = %v", err)
	}
	events, err := fixture.repo.ListAgentEvents(context.Background(), session.ID())
	// Then
	if err != nil {
		t.Fatalf("ListAgentEvents() error = %v", err)
	}
	if stored.Status() != domain.AgentSucceeded {
		t.Fatalf("agent status = %q, want succeeded", stored.Status())
	}
	for _, event := range events {
		if strings.Contains(event.Message(), prompt) {
			t.Fatalf("agent event stored prompt: %q", event.Message())
		}
	}
}
