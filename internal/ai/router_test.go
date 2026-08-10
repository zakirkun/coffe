package ai_test

import (
	"coffe/internal/ai"
	"coffe/internal/domain"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	coffesqlite "coffe/internal/store/sqlite"
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

func Test_Router_records_usage_with_real_SQLite_provider_constraint(t *testing.T) {
	// Given
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"choices":[{"message":{"content":"{\"action\":\"deploy\"}"}}],"usage":{"prompt_tokens":4,"completion_tokens":2,"total_tokens":6}}`))
	}))
	t.Cleanup(server.Close)
	store, err := coffesqlite.Open(ctx, filepath.Join(t.TempDir(), "coffe.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := store.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})
	configured := provider(t, "configured-fallback", server.URL, 1)
	if err := store.SaveAIProvider(ctx, configured); err != nil {
		t.Fatalf("SaveAIProvider() error = %v", err)
	}
	router := ai.NewRouter(ai.RouterDependencies{
		Store: store, Secrets: emptySecretReader{}, IDs: &providerIDs{}, Clock: providerClock{},
	})

	// When
	intent, err := router.Classify(ctx, "deploy")
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}
	usage, err := store.ListAIUsage(ctx, 10)

	// Then
	if err != nil {
		t.Fatalf("ListAIUsage() error = %v", err)
	}
	if intent != domain.IntentDeploy || len(usage) != 1 || usage[0].ProviderID() != configured.ID() {
		t.Fatalf("intent/usage = %q/%v", intent, usage)
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
