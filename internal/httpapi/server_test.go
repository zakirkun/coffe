package httpapi_test

import (
	"bytes"
	"coffe/internal/domain"
	"coffe/internal/httpapi"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func Test_HTTP_user_can_create_target_preview_approve_and_execute(t *testing.T) {
	// Given
	app, closeApp, root := newApplication(t)
	t.Cleanup(closeApp)
	server := httptest.NewServer(httpapi.New(app, slog.New(slog.DiscardHandler)))
	t.Cleanup(server.Close)
	client := server.Client()

	// When
	project := postJSON[projectResponse](t, postCall{
		client: client, url: server.URL + "/api/v1/projects",
		payload: map[string]any{"name": "payments", "root_path": root}, wantStatus: http.StatusCreated,
	})
	target := postJSON[targetResponse](t, postCall{
		client: client, url: server.URL + "/api/v1/targets", wantStatus: http.StatusCreated,
		payload: map[string]any{
			"project_id": project.ID, "name": "office", "host": "server.example.com", "port": 22,
			"user": "deploy", "remote_path": "/srv/payments", "branch": "main",
		},
	})
	preview := postJSON[planResponse](t, postCall{
		client: client, url: server.URL + "/api/v1/plans/preview", wantStatus: http.StatusCreated,
		payload: map[string]any{
			"project_id": project.ID, "target_id": target.ID, "prompt": "deploy payments",
		},
	})
	approved := postJSON[planResponse](t, postCall{
		client: client, url: server.URL + "/api/v1/plans/" + preview.ID + "/approve",
		payload: struct{}{}, wantStatus: http.StatusOK,
	})
	finished := postJSON[planResponse](t, postCall{
		client: client, url: server.URL + "/api/v1/plans/" + preview.ID + "/execute",
		payload: struct{}{}, wantStatus: http.StatusAccepted,
	})
	waitCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := app.Wait(waitCtx); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	stored, err := app.GetPlan(context.Background(), preview.ID)
	if err != nil {
		t.Fatalf("GetPlan() error = %v", err)
	}

	// Then
	if len(preview.Commands) != 7 {
		t.Fatalf("preview commands = %d, want 7", len(preview.Commands))
	}
	if approved.Status != string(domain.PlanApproved) {
		t.Fatalf("approved status = %q, want approved", approved.Status)
	}
	if finished.Status != string(domain.PlanRunning) || stored.Status() != domain.PlanSucceeded {
		t.Fatalf("execute response/stored status = %q/%q", finished.Status, stored.Status())
	}
}

func Test_HTTP_healthz_reports_ready(t *testing.T) {
	// Given
	app, closeApp, _ := newApplication(t)
	t.Cleanup(closeApp)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	// When
	httpapi.New(app, slog.New(slog.DiscardHandler)).ServeHTTP(recorder, request)

	// Then
	if recorder.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", recorder.Code)
	}
}

func Test_HTTP_user_can_register_and_query_SQLite_read_only(t *testing.T) {
	// Given
	app, closeApp, _ := newApplication(t)
	t.Cleanup(closeApp)
	server := httptest.NewServer(httpapi.New(app, slog.New(slog.DiscardHandler)))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "analytics.db")
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	if _, err := database.Exec(`CREATE TABLE metrics (name TEXT); INSERT INTO metrics VALUES ('jobs')`); err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// When
	connection := postJSON[databaseResponse](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/databases", wantStatus: http.StatusCreated,
		payload: map[string]any{"name": "analytics", "driver": "sqlite", "sqlite_path": path},
	})
	result := postJSON[domain.QueryResult](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/databases/" + connection.ID + "/query",
		wantStatus: http.StatusOK, payload: map[string]any{"sql": "SELECT name FROM metrics"},
	})
	history := getJSON[[]struct {
		Statement string `json:"statement"`
	}](t, server.Client(), server.URL+"/api/v1/databases/"+connection.ID+"/history")

	// Then
	if len(result.Rows) != 1 || result.Rows[0][0] != "jobs" ||
		len(history) != 1 || history[0].Statement != "SELECT name FROM metrics" {
		t.Fatalf("query rows/history = %v/%v", result.Rows, history)
	}
}

type projectResponse struct {
	ID string `json:"id"`
}
type targetResponse struct {
	ID string `json:"id"`
}
type databaseResponse struct {
	ID string `json:"id"`
}

type planResponse struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"`
	Commands []string `json:"commands"`
}

type postCall struct {
	client     *http.Client
	url        string
	payload    any
	wantStatus int
}

func postJSON[T any](t *testing.T, call postCall) T {
	t.Helper()
	body, err := json.Marshal(call.payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	response, err := call.client.Post(call.url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Errorf("Body.Close() error = %v", closeErr)
		}
	})
	if response.StatusCode != call.wantStatus {
		raw, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			t.Fatalf("ReadAll() error = %v", readErr)
		}
		t.Fatalf("POST %s status = %d, want %d: %s", call.url, response.StatusCode, call.wantStatus, raw)
	}
	var result T
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return result
}
