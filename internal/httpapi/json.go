package httpapi

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const maxRequestBytes = 1 << 20

type response struct {
	body   any
	status int
}

type errorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func decodeJSON[T any](writer http.ResponseWriter, request *http.Request) (T, error) {
	var value T
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode json: %w", service.ErrInvalidInput)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return value, fmt.Errorf("trailing json: %w", service.ErrInvalidInput)
	}
	return value, nil
}

func (s *Server) respond(ctx context.Context, writer http.ResponseWriter, value response) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(value.status)
	if err := json.NewEncoder(writer).Encode(value.body); err != nil {
		s.logger.ErrorContext(ctx, "http response encode failed", slog.Any("err", err))
	}
}

func (s *Server) fail(ctx context.Context, writer http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	body := errorDTO{Code: "internal", Message: "internal error"}
	switch {
	case errors.Is(err, service.ErrNotFound):
		status = http.StatusNotFound
		body = errorDTO{Code: "not_found", Message: "resource not found"}
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrReadOnlyQuery), isDomainInputError(err):
		status = http.StatusBadRequest
		body = errorDTO{Code: "invalid_input", Message: err.Error()}
	case errors.Is(err, domain.ErrPlanTransition), errors.Is(err, service.ErrNotRunning):
		status = http.StatusConflict
		body = errorDTO{Code: "invalid_transition", Message: err.Error()}
	case errors.Is(err, domain.ErrGitTransition):
		status = http.StatusConflict
		body = errorDTO{Code: "invalid_transition", Message: err.Error()}
	case errors.Is(err, service.ErrPolicyDenied):
		status = http.StatusForbidden
		body = errorDTO{Code: "policy_denied", Message: err.Error()}
	default:
		s.logger.ErrorContext(ctx, "http request failed", slog.Any("err", err))
	}
	s.respond(ctx, writer, response{status: status, body: body})
}

func isDomainInputError(err error) bool {
	return errors.Is(err, domain.ErrInvalidName) ||
		errors.Is(err, domain.ErrInvalidProject) ||
		errors.Is(err, domain.ErrInvalidTarget) ||
		errors.Is(err, domain.ErrInvalidHost) ||
		errors.Is(err, domain.ErrInvalidPort) ||
		errors.Is(err, domain.ErrInvalidSSHUser) ||
		errors.Is(err, domain.ErrInvalidRemotePath) ||
		errors.Is(err, domain.ErrInvalidBranch) ||
		errors.Is(err, domain.ErrInvalidGitOperation) ||
		errors.Is(err, domain.ErrInvalidGitRemote) ||
		errors.Is(err, domain.ErrInvalidAIProvider) ||
		errors.Is(err, domain.ErrInvalidPolicy) ||
		errors.Is(err, domain.ErrInvalidWorker) ||
		errors.Is(err, domain.ErrInvalidPlugin)
}
