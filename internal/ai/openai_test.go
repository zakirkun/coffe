package ai_test

import (
	"coffe/internal/ai"
	"coffe/internal/domain"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_OpenAIPlanner_Classify_parses_deploy_intent(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Authorization = %q", request.Header.Get("Authorization"))
		}
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("Decode() error = %v", err)
		}
		if body.Model != "local-model" {
			t.Errorf("model = %q, want local-model", body.Model)
		}
		writer.Header().Set("Content-Type", "application/json")
		if _, err := writer.Write([]byte(`{"choices":[{"message":{"content":"{\"action\":\"deploy\"}"}}]}`)); err != nil {
			t.Errorf("Write() error = %v", err)
		}
	}))
	t.Cleanup(server.Close)
	planner, err := ai.NewOpenAIPlanner(ai.OpenAIConfig{
		BaseURL: server.URL + "/v1", Model: "local-model", APIKey: "test-key",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewOpenAIPlanner() error = %v", err)
	}

	// When
	intent, err := planner.Classify(context.Background(), "deploy payments")
	// Then
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}
	if intent != domain.IntentDeploy {
		t.Fatalf("Classify() = %q, want deploy", intent)
	}
}

func Test_OpenAIPlanner_Classify_rejects_non_allowlisted_action(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		if _, err := writer.Write([]byte(`{"choices":[{"message":{"content":"{\"action\":\"shell\"}"}}]}`)); err != nil {
			t.Errorf("Write() error = %v", err)
		}
	}))
	t.Cleanup(server.Close)
	planner, err := ai.NewOpenAIPlanner(ai.OpenAIConfig{
		BaseURL: server.URL, Model: "local-model", Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewOpenAIPlanner() error = %v", err)
	}

	// When
	_, err = planner.Classify(context.Background(), "run whoami")

	// Then
	if err == nil {
		t.Fatal("Classify() error = nil, want unsupported action")
	}
}
