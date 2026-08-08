package httpapi

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"net/http"
	"time"
)

type previewGitRequest struct {
	ProjectID string `json:"project_id"`
	Kind      string `json:"kind"`
	Remote    string `json:"remote"`
	Branch    string `json:"branch"`
}

type gitOperationDTO struct {
	ID         string                    `json:"id"`
	ProjectID  string                    `json:"project_id"`
	Kind       domain.GitOperationKind   `json:"kind"`
	Remote     string                    `json:"remote"`
	Branch     string                    `json:"branch"`
	Status     domain.GitOperationStatus `json:"status"`
	Command    string                    `json:"command,omitempty"`
	CreatedAt  time.Time                 `json:"created_at"`
	ApprovedAt *time.Time                `json:"approved_at,omitempty"`
}

type gitEventDTO struct {
	OperationID string    `json:"operation_id"`
	Message     string    `json:"message"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type projectGitStatusDTO struct {
	Project projectDTO       `json:"project"`
	Status  domain.GitStatus `json:"status"`
	Error   string           `json:"error,omitempty"`
}

func (s *Server) gitStatus(writer http.ResponseWriter, request *http.Request) {
	status, err := s.app.GitStatus(request.Context(), request.PathValue("projectID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: status})
}

func (s *Server) gitBranches(writer http.ResponseWriter, request *http.Request) {
	branches, err := s.app.GitBranches(request.Context(), request.PathValue("projectID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: branches})
}

func (s *Server) gitDiff(writer http.ResponseWriter, request *http.Request) {
	diff, err := s.app.GitDiff(request.Context(), request.PathValue("projectID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: map[string]string{"diff": diff}})
}

func (s *Server) allGitStatuses(writer http.ResponseWriter, request *http.Request) {
	statuses, err := s.app.AllGitStatuses(request.Context())
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]projectGitStatusDTO, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, projectGitStatusDTO{
			Project: toProjectDTO(status.Project), Status: status.Status, Error: status.Error,
		})
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) previewGit(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[previewGitRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	preview, err := s.app.PreviewGit(request.Context(), service.CreateGitOperationInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{
		status: http.StatusCreated, body: toGitOperationDTO(preview.Operation, preview.Command),
	})
}

func (s *Server) approveGit(writer http.ResponseWriter, request *http.Request) {
	operation, err := s.app.ApproveGit(request.Context(), request.PathValue("operationID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: toGitOperationDTO(operation, "")})
}

func (s *Server) executeGit(writer http.ResponseWriter, request *http.Request) {
	operation, err := s.app.ExecuteGit(request.Context(), request.PathValue("operationID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: toGitOperationDTO(operation, "")})
}

func (s *Server) listGitOperations(writer http.ResponseWriter, request *http.Request) {
	operations, err := s.app.ListGitOperations(request.Context(), 50)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]gitOperationDTO, 0, len(operations))
	for _, operation := range operations {
		result = append(result, toGitOperationDTO(operation, ""))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) listGitEvents(writer http.ResponseWriter, request *http.Request) {
	events, err := s.app.ListGitEvents(request.Context(), request.PathValue("operationID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]gitEventDTO, 0, len(events))
	for _, event := range events {
		result = append(result, gitEventDTO{
			OperationID: string(event.OperationID()), Message: event.Message(), OccurredAt: event.OccurredAt(),
		})
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func toGitOperationDTO(operation domain.GitOperation, command string) gitOperationDTO {
	result := gitOperationDTO{
		ID: string(operation.ID()), ProjectID: string(operation.ProjectID()), Kind: operation.Kind(),
		Remote: operation.Remote().String(), Branch: operation.Branch().String(),
		Status: operation.Status(), Command: command, CreatedAt: operation.CreatedAt(),
	}
	if approvedAt, ok := operation.ApprovedAt(); ok {
		result.ApprovedAt = &approvedAt
	}
	return result
}
