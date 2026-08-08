package mcpserver

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	getPolicyInput struct{}
	policyOutput   struct {
		AllowGitPush        bool               `json:"allow_git_push"`
		AllowSwarmDeploy    bool               `json:"allow_swarm_deploy"`
		AgentTools          []domain.AgentTool `json:"agent_tools"`
		MaxConcurrentAgents int                `json:"max_concurrent_agents"`
		AuditRetentionDays  int                `json:"audit_retention_days"`
	}
)

type (
	listAuditInput struct{}
	auditRecord    struct {
		Category   domain.AuditCategory `json:"category"`
		Action     string               `json:"action"`
		ResourceID string               `json:"resource_id"`
		Outcome    domain.AuditOutcome  `json:"outcome"`
		Message    string               `json:"message"`
		OccurredAt time.Time            `json:"occurred_at"`
	}
)

type auditOutput struct {
	Entries []auditRecord `json:"entries"`
}

func registerPolicyAuditTools(server *mcp.Server, app Application) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_policy", Description: "Read local Coffe safety policy",
	}, getPolicyHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_audit", Description: "Read redacted local audit records",
	}, listAuditHandler(app))
}

func getPolicyHandler(app Application) mcp.ToolHandlerFor[getPolicyInput, policyOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ getPolicyInput,
	) (*mcp.CallToolResult, policyOutput, error) {
		policy, err := app.GetPolicy(ctx)
		if err != nil {
			return nil, policyOutput{}, fmt.Errorf("get policy: %w", err)
		}
		return nil, policyOutput{
			AllowGitPush: policy.AllowGitPush(), AllowSwarmDeploy: policy.AllowSwarmDeploy(),
			AgentTools: policy.AgentTools(), MaxConcurrentAgents: policy.MaxConcurrentAgents(),
			AuditRetentionDays: policy.AuditRetentionDays(),
		}, nil
	}
}

func listAuditHandler(app Application) mcp.ToolHandlerFor[listAuditInput, auditOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ listAuditInput,
	) (*mcp.CallToolResult, auditOutput, error) {
		entries, err := app.ListAudit(ctx, 200)
		if err != nil {
			return nil, auditOutput{}, fmt.Errorf("list audit: %w", err)
		}
		output := make([]auditRecord, 0, len(entries))
		for _, entry := range entries {
			output = append(output, auditRecord{
				Category: entry.Category(), Action: entry.Action(), ResourceID: entry.ResourceID(),
				Outcome: entry.Outcome(), Message: entry.Message(), OccurredAt: entry.OccurredAt(),
			})
		}
		return nil, auditOutput{Entries: output}, nil
	}
}
