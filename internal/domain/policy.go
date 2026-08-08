package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

var ErrInvalidPolicy = errors.New("domain: invalid policy")

type Policy struct {
	updatedAt           time.Time
	agentTools          []AgentTool
	maxConcurrentAgents int
	auditRetentionDays  int
	allowGitPush        bool
	allowSwarmDeploy    bool
}

type PolicyValues struct {
	AllowGitPush        bool
	AllowSwarmDeploy    bool
	AgentTools          []AgentTool
	MaxConcurrentAgents int
	AuditRetentionDays  int
	UpdatedAt           time.Time
}

func NewPolicy(values PolicyValues) (Policy, error) {
	if values.MaxConcurrentAgents < 1 || values.MaxConcurrentAgents > 32 ||
		values.AuditRetentionDays < 1 || values.AuditRetentionDays > 3650 || values.UpdatedAt.IsZero() {
		return Policy{}, fmt.Errorf("policy limits: %w", ErrInvalidPolicy)
	}
	tools := slices.Clone(values.AgentTools)
	slices.Sort(tools)
	tools = slices.Compact(tools)
	for _, tool := range tools {
		if !tool.Valid() {
			return Policy{}, fmt.Errorf("policy agent tool %q: %w", tool, ErrInvalidPolicy)
		}
	}
	return Policy{
		allowGitPush: values.AllowGitPush, allowSwarmDeploy: values.AllowSwarmDeploy,
		agentTools: tools, maxConcurrentAgents: values.MaxConcurrentAgents,
		auditRetentionDays: values.AuditRetentionDays, updatedAt: values.UpdatedAt.UTC(),
	}, nil
}

func DefaultPolicy(now time.Time) Policy {
	return Policy{
		allowGitPush: true, allowSwarmDeploy: true,
		agentTools:          []AgentTool{AgentAider, AgentClaude, AgentCodex, AgentGemini, AgentOpenCode},
		maxConcurrentAgents: 4, auditRetentionDays: 365, updatedAt: now.UTC(),
	}
}

func (p Policy) AllowGitPush() bool              { return p.allowGitPush }
func (p Policy) AllowSwarmDeploy() bool          { return p.allowSwarmDeploy }
func (p Policy) AgentTools() []AgentTool         { return slices.Clone(p.agentTools) }
func (p Policy) MaxConcurrentAgents() int        { return p.maxConcurrentAgents }
func (p Policy) AuditRetentionDays() int         { return p.auditRetentionDays }
func (p Policy) UpdatedAt() time.Time            { return p.updatedAt }
func (p Policy) AllowsAgent(tool AgentTool) bool { return slices.Contains(p.agentTools, tool) }
