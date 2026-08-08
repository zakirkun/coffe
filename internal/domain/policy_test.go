package domain_test

import (
	"coffe/internal/domain"
	"testing"
	"time"
)

func Test_DefaultPolicy_keeps_approval_controls_and_agent_limit(t *testing.T) {
	// Given
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)

	// When
	policy := domain.DefaultPolicy(now)

	// Then
	if !policy.AllowGitPush() || !policy.AllowSwarmDeploy() || policy.MaxConcurrentAgents() != 4 {
		t.Fatalf("default policy = %#v", policy)
	}
	if !policy.AllowsAgent(domain.AgentOpenCode) {
		t.Fatal("default policy must allow OpenCode")
	}
}
