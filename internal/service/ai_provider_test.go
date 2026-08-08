package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"testing"
)

func Test_App_CreateAIProvider_persists_key_reference_not_key(t *testing.T) {
	// Given
	fixture := newFixture(t)

	// When
	provider, err := fixture.app.CreateAIProvider(context.Background(), service.CreateAIProviderInput{
		Name: "local-ollama", Kind: string(domain.AIProviderOllama),
		BaseURL: "http://127.0.0.1:11434/v1", Model: "qwen3", Priority: 1, Enabled: true,
	})
	// Then
	if err != nil {
		t.Fatalf("CreateAIProvider() error = %v", err)
	}
	if provider.SecretName() != "" || provider.Kind() != domain.AIProviderOllama {
		t.Fatalf("provider secret/kind = %q/%q", provider.SecretName(), provider.Kind())
	}
}
