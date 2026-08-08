package httpapi

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"net/http"
	"time"
)

type startAgentRequest struct {
	ProjectID string `json:"project_id"`
	Tool      string `json:"tool"`
	Prompt    string `json:"prompt"`
}

type agentSessionDTO struct {
	ID         string             `json:"id"`
	ProjectID  string             `json:"project_id"`
	Tool       domain.AgentTool   `json:"tool"`
	Status     domain.AgentStatus `json:"status"`
	CreatedAt  time.Time          `json:"created_at"`
	StartedAt  *time.Time         `json:"started_at,omitempty"`
	FinishedAt *time.Time         `json:"finished_at,omitempty"`
}

type agentEventDTO struct {
	Message    string    `json:"message"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (s *Server) startAgent(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[startAgentRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	session, err := s.app.StartAgent(request.Context(), service.StartAgentInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusAccepted, body: toAgentSessionDTO(session)})
}

func (s *Server) stopAgent(writer http.ResponseWriter, request *http.Request) {
	if err := s.app.StopAgent(request.Context(), request.PathValue("sessionID")); err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{
		status: http.StatusAccepted, body: map[string]string{"status": "stop_requested"},
	})
}

func (s *Server) listAgents(writer http.ResponseWriter, request *http.Request) {
	sessions, err := s.app.ListAgentSessions(request.Context(), 50)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]agentSessionDTO, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, toAgentSessionDTO(session))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) listAgentEvents(writer http.ResponseWriter, request *http.Request) {
	events, err := s.app.ListAgentEvents(request.Context(), request.PathValue("sessionID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]agentEventDTO, 0, len(events))
	for _, event := range events {
		result = append(result, agentEventDTO{Message: event.Message(), OccurredAt: event.OccurredAt()})
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func toAgentSessionDTO(session domain.AgentSession) agentSessionDTO {
	result := agentSessionDTO{
		ID: string(session.ID()), ProjectID: string(session.ProjectID()),
		Tool: session.Tool(), Status: session.Status(), CreatedAt: session.CreatedAt(),
	}
	if startedAt := session.StartedAt(); !startedAt.IsZero() {
		result.StartedAt = &startedAt
	}
	if finishedAt := session.FinishedAt(); !finishedAt.IsZero() {
		result.FinishedAt = &finishedAt
	}
	return result
}
