package sqlite

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (s *Store) SavePolicy(ctx context.Context, policy domain.Policy) error {
	tools, err := json.Marshal(policy.AgentTools())
	if err != nil {
		return fmt.Errorf("encode policy tools: %w", err)
	}
	const query = `INSERT INTO policy (
		id, allow_git_push, allow_swarm_deploy, agent_tools_json,
		max_concurrent_agents, audit_retention_days, updated_at
	) VALUES (1, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		allow_git_push = excluded.allow_git_push,
		allow_swarm_deploy = excluded.allow_swarm_deploy,
		agent_tools_json = excluded.agent_tools_json,
		max_concurrent_agents = excluded.max_concurrent_agents,
		audit_retention_days = excluded.audit_retention_days,
		updated_at = excluded.updated_at`
	if _, err := s.db.ExecContext(ctx, query,
		policy.AllowGitPush(), policy.AllowSwarmDeploy(), string(tools),
		policy.MaxConcurrentAgents(), policy.AuditRetentionDays(), formatTime(policy.UpdatedAt()),
	); err != nil {
		return fmt.Errorf("save policy: %w", err)
	}
	return nil
}

func (s *Store) GetPolicy(ctx context.Context) (domain.Policy, error) {
	const query = `SELECT allow_git_push, allow_swarm_deploy, agent_tools_json,
		max_concurrent_agents, audit_retention_days, updated_at FROM policy WHERE id = 1`
	var rawTools, rawUpdatedAt string
	var allowGitPush, allowSwarmDeploy bool
	var maxConcurrentAgents, auditRetentionDays int
	if err := s.db.QueryRowContext(ctx, query).Scan(
		&allowGitPush, &allowSwarmDeploy, &rawTools,
		&maxConcurrentAgents, &auditRetentionDays, &rawUpdatedAt,
	); errors.Is(err, sql.ErrNoRows) {
		return domain.Policy{}, service.ErrNotFound
	} else if err != nil {
		return domain.Policy{}, fmt.Errorf("scan policy: %w", err)
	}
	var tools []domain.AgentTool
	if err := json.Unmarshal([]byte(rawTools), &tools); err != nil {
		return domain.Policy{}, fmt.Errorf("decode policy tools: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, rawUpdatedAt)
	if err != nil {
		return domain.Policy{}, fmt.Errorf("parse policy updated_at: %w", err)
	}
	policy, err := domain.NewPolicy(domain.PolicyValues{
		AllowGitPush: allowGitPush, AllowSwarmDeploy: allowSwarmDeploy,
		AgentTools: tools, MaxConcurrentAgents: maxConcurrentAgents,
		AuditRetentionDays: auditRetentionDays, UpdatedAt: updatedAt,
	})
	if err != nil {
		return domain.Policy{}, fmt.Errorf("restore policy: %w", err)
	}
	return policy, nil
}
