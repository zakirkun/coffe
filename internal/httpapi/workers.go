package httpapi

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"net/http"
	"time"
)

type registerWorkerRequest struct {
	TargetID   string `json:"target_id"`
	Name       string `json:"name"`
	BaseURL    string `json:"base_url"`
	SecretName string `json:"secret_name"`
	Enabled    bool   `json:"enabled"`
}

type workerDTO struct {
	ID        string    `json:"id"`
	TargetID  string    `json:"target_id"`
	Name      string    `json:"name"`
	BaseURL   string    `json:"base_url"`
	Version   string    `json:"version,omitempty"`
	Enabled   bool      `json:"enabled"`
	LastSeen  time.Time `json:"last_seen,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Server) registerWorker(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[registerWorkerRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	registered, err := s.app.RegisterWorker(request.Context(), service.RegisterWorkerInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusCreated, body: toWorkerDTO(registered)})
}

func (s *Server) heartbeatWorker(writer http.ResponseWriter, request *http.Request) {
	updated, err := s.app.HeartbeatWorker(request.Context(), request.PathValue("workerID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: toWorkerDTO(updated)})
}

func (s *Server) listWorkers(writer http.ResponseWriter, request *http.Request) {
	workers, err := s.app.ListWorkers(request.Context())
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]workerDTO, 0, len(workers))
	for _, registered := range workers {
		result = append(result, toWorkerDTO(registered))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func toWorkerDTO(worker domain.Worker) workerDTO {
	return workerDTO{
		ID: string(worker.ID()), TargetID: string(worker.TargetID()), Name: worker.Name().String(),
		BaseURL: worker.BaseURL(), Version: worker.Version(), Enabled: worker.Enabled(),
		LastSeen: worker.LastSeen(), CreatedAt: worker.CreatedAt(),
	}
}
