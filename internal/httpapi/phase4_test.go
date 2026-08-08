package httpapi_test

import (
	"bytes"
	"coffe/internal/httpapi"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_HTTP_policy_denial_is_recorded_in_audit(t *testing.T) {
	// Given
	app, closeApp, root := newApplication(t)
	t.Cleanup(closeApp)
	server := httptest.NewServer(httpapi.New(app, slog.New(slog.DiscardHandler)))
	t.Cleanup(server.Close)
	project := postJSON[projectResponse](t, postCall{
		client: server.Client(), url: server.URL + "/api/v1/projects",
		payload: map[string]any{"name": "policy-test", "root_path": root}, wantStatus: http.StatusCreated,
	})
	policyBody, err := json.Marshal(map[string]any{
		"allow_git_push": false, "allow_swarm_deploy": true,
		"agent_tools": []string{"opencode"}, "max_concurrent_agents": 1, "audit_retention_days": 30,
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	request, err := http.NewRequest(http.MethodPut, server.URL+"/api/v1/policy", bytes.NewReader(policyBody))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatalf("Do(policy) error = %v", err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("Close(policy) error = %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("policy status = %d", response.StatusCode)
	}

	// When
	pushBody, err := json.Marshal(map[string]any{
		"project_id": project.ID, "kind": "push", "remote": "origin", "branch": "main",
	})
	if err != nil {
		t.Fatalf("Marshal(push) error = %v", err)
	}
	pushResponse, err := server.Client().Post(
		server.URL+"/api/v1/git/operations/preview", "application/json", bytes.NewReader(pushBody),
	)
	if err != nil {
		t.Fatalf("Post(push) error = %v", err)
	}
	if err := pushResponse.Body.Close(); err != nil {
		t.Fatalf("Close(push) error = %v", err)
	}
	audit := getJSON[[]struct {
		Outcome string `json:"outcome"`
	}](t, server.Client(), server.URL+"/api/v1/audit")

	// Then
	if pushResponse.StatusCode != http.StatusForbidden || len(audit) != 2 || audit[0].Outcome != "denied" {
		t.Fatalf("push status/audit = %d/%v", pushResponse.StatusCode, audit)
	}
}
