package ai_test

import (
	"coffe/internal/ai"
	"coffe/internal/domain"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_Router_falls_back_and_records_each_attempt(t *testing.T) {
	// Given
	failing := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(failing.Close)
	succeeding := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		if _, err := writer.Write([]byte(`{"choices":[{"message":{"content":"{\"action\":\"deploy\"}"}}],"usage":{"prompt_tokens":4,"completion_tokens":2,"total_tokens":6}}`)); err != nil {
			t.Errorf("Write() error = %v", err)
		}
	}))
	t.Cleanup(succeeding.Close)
	store := &providerStore{providers: []domain.AIProvider{
		provider(t, "provider-1", failing.URL, 1), provider(t, "provider-2", succeeding.URL, 2),
	}}
	router := ai.NewRouter(ai.RouterDependencies{
		Store: store, Secrets: emptySecretReader{}, IDs: &providerIDs{}, Clock: providerClock{},
	})

	// When
	intent, err := router.Classify(context.Background(), "deploy")
	// Then
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}
	if intent != domain.IntentDeploy || len(store.usage) != 2 || !store.usage[1].Success() {
		t.Fatalf("intent/usage = %q/%d/%v", intent, len(store.usage), store.usage)
	}
}

type providerStore struct {
	providers []domain.AIProvider
	usage     []domain.AIUsage
}

func (s *providerStore) ListAIProviders(context.Context) ([]domain.AIProvider, error) {
	return s.providers, nil
}

func (s *providerStore) SaveAIUsage(_ context.Context, usage domain.AIUsage) error {
	s.usage = append(s.usage, usage)
	return nil
}

type emptySecretReader struct{}

func (emptySecretReader) Get(context.Context, string) (string, error) { return "", nil }

type providerIDs struct{ next int }

func (s *providerIDs) New() (string, error) {
	s.next++
	return fmt.Sprintf("usage-%d", s.next), nil
}

type providerClock struct{}

func (providerClock) Now() time.Time {
	return time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
}

func provider(t *testing.T, id, baseURL string, priority int) domain.AIProvider {
	t.Helper()
	name, err := domain.NewName(id)
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}
	provider, err := domain.NewAIProvider(domain.AIProviderValues{
		ID: domain.AIProviderID(id), Name: name, Kind: domain.AIProviderVLLM,
		BaseURL: baseURL, Model: "local-model", Priority: priority, Enabled: true,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewAIProvider() error = %v", err)
	}
	return provider
}
