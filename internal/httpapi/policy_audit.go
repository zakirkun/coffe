package httpapi

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"log/slog"
	"net/http"
	"time"
)

type updatePolicyRequest struct {
	AllowGitPush        bool     `json:"allow_git_push"`
	AllowSwarmDeploy    bool     `json:"allow_swarm_deploy"`
	AgentTools          []string `json:"agent_tools"`
	MaxConcurrentAgents int      `json:"max_concurrent_agents"`
	AuditRetentionDays  int      `json:"audit_retention_days"`
}

type policyDTO struct {
	AllowGitPush        bool               `json:"allow_git_push"`
	AllowSwarmDeploy    bool               `json:"allow_swarm_deploy"`
	AgentTools          []domain.AgentTool `json:"agent_tools"`
	MaxConcurrentAgents int                `json:"max_concurrent_agents"`
	AuditRetentionDays  int                `json:"audit_retention_days"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

type auditDTO struct {
	ID         string               `json:"id"`
	Category   domain.AuditCategory `json:"category"`
	Action     string               `json:"action"`
	ResourceID string               `json:"resource_id"`
	Outcome    domain.AuditOutcome  `json:"outcome"`
	Message    string               `json:"message"`
	OccurredAt time.Time            `json:"occurred_at"`
}

func (s *Server) getPolicy(writer http.ResponseWriter, request *http.Request) {
	policy, err := s.app.GetPolicy(request.Context())
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: toPolicyDTO(policy)})
}

func (s *Server) updatePolicy(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[updatePolicyRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	policy, err := s.app.UpdatePolicy(request.Context(), service.UpdatePolicyInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: toPolicyDTO(policy)})
}

func (s *Server) listAudit(writer http.ResponseWriter, request *http.Request) {
	entries, err := s.app.ListAudit(request.Context(), 200)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]auditDTO, 0, len(entries))
	for _, entry := range entries {
		result = append(result, auditDTO{
			ID: string(entry.ID()), Category: entry.Category(), Action: entry.Action(),
			ResourceID: entry.ResourceID(), Outcome: entry.Outcome(),
			Message: entry.Message(), OccurredAt: entry.OccurredAt(),
		})
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) exportAuditCSV(writer http.ResponseWriter, request *http.Request) {
	content, err := s.app.ExportAuditCSV(request.Context(), 1000)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
	writer.Header().Set("Content-Disposition", `attachment; filename="audit.csv"`)
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(content); err != nil { //nolint:gosec // response is text/csv and formula-leading fields are escaped.
		s.logger.ErrorContext(request.Context(), "audit CSV response write failed", slog.Any("err", err))
	}
}

func toPolicyDTO(policy domain.Policy) policyDTO {
	return policyDTO{
		AllowGitPush: policy.AllowGitPush(), AllowSwarmDeploy: policy.AllowSwarmDeploy(),
		AgentTools: policy.AgentTools(), MaxConcurrentAgents: policy.MaxConcurrentAgents(),
		AuditRetentionDays: policy.AuditRetentionDays(), UpdatedAt: policy.UpdatedAt(),
	}
}
