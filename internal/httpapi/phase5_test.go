package httpapi_test

import (
	"coffe/internal/httpapi"
	"coffe/internal/pluginexec"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_HTTP_registers_worker_and_hash_pinned_plugin(t *testing.T) {
	// Given
	app, closeApp, root := newApplication(t)
	t.Cleanup(closeApp)
	server := httptest.NewServer(httpapi.New(app, slog.New(slog.DiscardHandler)))
	t.Cleanup(server.Close)
	project := postJSON[projectResponse](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/projects",
		payload: map[string]any{"name": "phase5", "root_path": root}, wantStatus: http.StatusCreated,
	})
	target := postJSON[targetResponse](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/targets", wantStatus: http.StatusCreated,
		payload: map[string]any{
			"project_id": project.ID, "name": "office", "host": "server.example.com", "port": 22,
			"user": "deploy", "remote_path": "/srv/app", "branch": "main",
		},
	})
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("Executable() error = %v", err)
	}
	hash, err := pluginexec.HashFile(executable)
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}

	// When
	registeredWorker := postJSON[workerResponse](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/workers", wantStatus: http.StatusCreated,
		payload: map[string]any{
			"target_id": target.ID, "name": "office-worker", "base_url": "https://worker.example.com",
			"secret_name": "worker-token", "enabled": true,
		},
	})
	registeredPlugin := postJSON[pluginResponse](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/plugins", wantStatus: http.StatusCreated,
		payload: map[string]any{
			"name": "incident-tools", "executable_path": executable, "sha256": hash,
			"methods": []string{"incident.lookup"}, "enabled": true,
		},
	})

	// Then
	if registeredWorker.ID == "" || registeredPlugin.ID == "" {
		t.Fatalf("worker/plugin IDs = %q/%q", registeredWorker.ID, registeredPlugin.ID)
	}
}

type (
	workerResponse struct {
		ID string `json:"id"`
	}
	pluginResponse struct {
		ID string `json:"id"`
	}
)
