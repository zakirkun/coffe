package domain_test

import (
	"coffe/internal/domain"
	"testing"
	"time"
)

func Test_NewAIProvider_accepts_local_OpenAI_compatible_provider(t *testing.T) {
	// Given
	name, err := domain.NewName("local-vllm")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}

	// When
	provider, err := domain.NewAIProvider(domain.AIProviderValues{
		ID: "provider-1", Name: name, Kind: domain.AIProviderVLLM,
		BaseURL: "http://127.0.0.1:8000/v1", Model: "Qwen/Qwen3-8B",
		Priority: 10, Enabled: true, CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	// Then
	if err != nil {
		t.Fatalf("NewAIProvider() error = %v", err)
	}
	if provider.BaseURL() != "http://127.0.0.1:8000/v1" || !provider.Enabled() {
		t.Fatalf("provider base/enabled = %q/%v", provider.BaseURL(), provider.Enabled())
	}
}
