package httpapi_test

import (
	"coffe/internal/domain"
	"coffe/internal/httpapi"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
)

func Test_HTTP_registers_AI_provider_and_previews_Git(t *testing.T) {
	// Given
	app, closeApp, root := newApplication(t)
	t.Cleanup(closeApp)
	command := exec.CommandContext(context.Background(), "git", "-C", root, "init", "-b", "main") //nolint:gosec // test invokes fixed git against a temporary directory.
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git init error = %v: %s", err, output)
	}
	server := httptest.NewServer(httpapi.New(app, slog.New(slog.DiscardHandler)))
	t.Cleanup(server.Close)
	project := postJSON[projectResponse](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/projects",
		payload: map[string]any{"name": "phase3", "root_path": root}, wantStatus: http.StatusCreated,
	})

	// When
	status := getJSON[domain.GitStatus](t, server.Client(), server.URL+"/api/v1/projects/"+project.ID+"/git/status")
	gitPreview := postJSON[gitOperationResponse](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/git/operations/preview",
		payload: map[string]any{
			"project_id": project.ID, "kind": "pull", "remote": "origin", "branch": "main",
		},
		wantStatus: http.StatusCreated,
	})
	provider := postJSON[providerResponse](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/ai/providers",
		payload: map[string]any{
			"name": "local-vllm", "kind": "vllm", "base_url": "http://127.0.0.1:8000/v1",
			"model": "qwen3", "priority": 1, "enabled": true,
		},
		wantStatus: http.StatusCreated,
	})

	// Then
	if status.Branch != "main" || gitPreview.Command == "" || provider.Kind != "vllm" {
		t.Fatalf("status/git/provider = %#v/%#v/%#v", status, gitPreview, provider)
	}
}

type gitOperationResponse struct {
	ID      string `json:"id"`
	Command string `json:"command"`
}

type providerResponse struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

func getJSON[T any](t *testing.T, client *http.Client, url string) T {
	t.Helper()
	response, err := client.Get(url)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Errorf("Body.Close() error = %v", closeErr)
		}
	}()
	if response.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			t.Fatalf("ReadAll() error = %v", readErr)
		}
		t.Fatalf("GET %s status = %d: %s", url, response.StatusCode, body)
	}
	var result T
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return result
}
