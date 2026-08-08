package service

import (
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"
)

type UpdatePolicyInput struct {
	AllowGitPush        bool
	AllowSwarmDeploy    bool
	AgentTools          []string
	MaxConcurrentAgents int
	AuditRetentionDays  int
}

func (a *App) EnsureDefaultPolicy(ctx context.Context) error {
	if _, err := a.repository.GetPolicy(ctx); err == nil {
		return nil
	} else if !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("get policy: %w", err)
	}
	if err := a.repository.SavePolicy(ctx, domain.DefaultPolicy(a.clock.Now())); err != nil {
		return fmt.Errorf("save default policy: %w", err)
	}
	return nil
}

func (a *App) GetPolicy(ctx context.Context) (domain.Policy, error) {
	policy, err := a.repository.GetPolicy(ctx)
	if err != nil {
		return domain.Policy{}, fmt.Errorf("get policy: %w", err)
	}
	return policy, nil
}

func (a *App) UpdatePolicy(ctx context.Context, input UpdatePolicyInput) (domain.Policy, error) {
	tools := make([]domain.AgentTool, 0, len(input.AgentTools))
	for _, raw := range input.AgentTools {
		tool := domain.AgentTool(strings.TrimSpace(raw))
		if !tool.Valid() {
			return domain.Policy{}, fmt.Errorf("agent tool %q: %w", raw, ErrInvalidInput)
		}
		tools = append(tools, tool)
	}
	policy, err := domain.NewPolicy(domain.PolicyValues{
		AllowGitPush: input.AllowGitPush, AllowSwarmDeploy: input.AllowSwarmDeploy,
		AgentTools: tools, MaxConcurrentAgents: input.MaxConcurrentAgents,
		AuditRetentionDays: input.AuditRetentionDays, UpdatedAt: a.clock.Now(),
	})
	if err != nil {
		return domain.Policy{}, err
	}
	if err := a.repository.SavePolicy(ctx, policy); err != nil {
		return domain.Policy{}, fmt.Errorf("save policy: %w", err)
	}
	if err := a.appendAudit(ctx, auditInput{
		category: domain.AuditPolicy, action: "update", resourceID: "local-policy",
		outcome: domain.AuditSuccess, message: "local policy updated",
	}); err != nil {
		return domain.Policy{}, err
	}
	return policy, nil
}

func (a *App) policyAllowsGit(ctx context.Context, kind domain.GitOperationKind) error {
	if kind != domain.GitPush {
		return nil
	}
	policy, err := a.repository.GetPolicy(ctx)
	if err != nil {
		return fmt.Errorf("get Git policy: %w", err)
	}
	if !policy.AllowGitPush() {
		return fmt.Errorf("git push: %w", ErrPolicyDenied)
	}
	return nil
}

func (a *App) policyAllowsSwarm(ctx context.Context) error {
	policy, err := a.repository.GetPolicy(ctx)
	if err != nil {
		return fmt.Errorf("get Swarm policy: %w", err)
	}
	if !policy.AllowSwarmDeploy() {
		return fmt.Errorf("swarm deploy: %w", ErrPolicyDenied)
	}
	return nil
}

func (a *App) policyAllowsAgent(ctx context.Context, tool domain.AgentTool) error {
	policy, err := a.repository.GetPolicy(ctx)
	if err != nil {
		return fmt.Errorf("get agent policy: %w", err)
	}
	if !policy.AllowsAgent(tool) {
		return fmt.Errorf("agent tool %s: %w", tool, ErrPolicyDenied)
	}
	sessions, err := a.repository.ListAgentSessions(ctx, 100)
	if err != nil {
		return fmt.Errorf("list running agents: %w", err)
	}
	running := 0
	for _, session := range sessions {
		if session.Status() == domain.AgentRunning {
			running++
		}
	}
	if running >= policy.MaxConcurrentAgents() {
		return fmt.Errorf("agent concurrency limit: %w", ErrPolicyDenied)
	}
	return nil
}
