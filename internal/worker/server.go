package worker

import (
	"bytes"
	"coffe/internal/deploy"
	"coffe/internal/domain"
	"coffe/internal/redact"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

const (
	maxRequestBytes = 64 * 1024
	maxOutputBytes  = 256 * 1024
)

type ServerConfig struct {
	Token   string
	Version string
}

type Server struct {
	token   string
	version string
	mux     *http.ServeMux
}

func NewServer(config ServerConfig) (*Server, error) {
	if len(config.Token) < 24 || strings.TrimSpace(config.Version) == "" {
		return nil, fmt.Errorf("worker token or version is invalid")
	}
	server := &Server{token: config.Token, version: config.Version, mux: http.NewServeMux()}
	server.mux.HandleFunc("GET /v1/health", server.health)
	server.mux.HandleFunc("POST /v1/tasks", server.task)
	return server, nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if !s.authorized(request) {
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) authorized(request *http.Request) bool {
	value := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	if len(value) != len(s.token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(value), []byte(s.token)) == 1
}

func (s *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, HealthResponse{Status: "ready", Version: s.version})
}

func (s *Server) task(writer http.ResponseWriter, request *http.Request) {
	if runtime.GOOS == "windows" {
		http.Error(writer, "worker execution requires a POSIX host", http.StatusNotImplemented)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var input TaskRequest
	if err := decoder.Decode(&input); err != nil {
		http.Error(writer, "invalid task", http.StatusBadRequest)
		return
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(writer, "invalid task", http.StatusBadRequest)
		return
	}
	execution, err := parseExecution(input)
	if err != nil {
		http.Error(writer, "invalid task", http.StatusBadRequest)
		return
	}
	command, err := deploy.Render(execution)
	if err != nil {
		http.Error(writer, "unsupported task", http.StatusBadRequest)
		return
	}
	output := &cappedWriter{limit: maxOutputBytes}
	process := exec.CommandContext(request.Context(), "sh", "-c", command) //nolint:gosec // command is rendered from closed StepKind and parsed domain values.
	process.Stdout = output
	process.Stderr = output
	if err := process.Run(); err != nil {
		writeJSON(writer, http.StatusUnprocessableEntity, TaskResponse{Logs: output.Lines()})
		return
	}
	logs := output.Lines()
	response := TaskResponse{Logs: logs}
	if input.Step == domain.StepInspectRepository || input.Step == domain.StepCaptureRevision {
		if len(logs) == 0 {
			writeJSON(writer, http.StatusUnprocessableEntity, TaskResponse{Logs: logs})
			return
		}
		revision, err := domain.NewGitRevision(logs[len(logs)-1])
		if err != nil {
			writeJSON(writer, http.StatusUnprocessableEntity, TaskResponse{Logs: logs})
			return
		}
		response.Revision = revision.String()
	}
	writeJSON(writer, http.StatusOK, response)
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		return
	}
}

type cappedWriter struct {
	mu     sync.Mutex
	buffer bytes.Buffer
	limit  int
}

func (w *cappedWriter) Write(content []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	length := len(content)
	remaining := w.limit - w.buffer.Len()
	if remaining <= 0 {
		return length, nil
	}
	if len(content) > remaining {
		content = content[:remaining]
	}
	_, err := w.buffer.Write(content)
	return length, err
}

func (w *cappedWriter) Lines() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	var lines []string
	for line := range strings.SplitSeq(w.buffer.String(), "\n") {
		if clean := redact.Line(line); clean != "" {
			lines = append(lines, clean)
		}
	}
	return lines
}
