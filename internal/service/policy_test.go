package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"errors"
	"testing"
)

func Test_App_policy_blocks_Git_push_preview(t *testing.T) {
	// Given
	fixture := newFixture(t)
	_, err := fixture.app.UpdatePolicy(context.Background(), service.UpdatePolicyInput{
		AllowGitPush: false, AllowSwarmDeploy: true,
		AgentTools:          []string{string(domain.AgentOpenCode)},
		MaxConcurrentAgents: 1, AuditRetentionDays: 30,
	})
	if err != nil {
		t.Fatalf("UpdatePolicy() error = %v", err)
	}

	// When
	_, err = fixture.app.PreviewGit(context.Background(), service.CreateGitOperationInput{
		ProjectID: string(fixture.project.ID()), Kind: string(domain.GitPush),
		Remote: "origin", Branch: "main",
	})

	// Then
	if !errors.Is(err, service.ErrPolicyDenied) {
		t.Fatalf("PreviewGit() error = %v, want ErrPolicyDenied", err)
	}
	if len(fixture.repo.audit) != 2 {
		t.Fatalf("audit entries = %d, want policy update and denial", len(fixture.repo.audit))
	}
}
