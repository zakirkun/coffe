package bootstrap_test

import (
	"coffe/internal/bootstrap"
	"coffe/internal/service"
	"context"
	"io"
	"log/slog"
	"math"
	"path/filepath"
	"testing"
)

func Test_Open_avoids_fallback_name_collision_and_reuses_stable_provider(t *testing.T) {
	// Given
	ctx := context.Background()
	options := bootstrap.Options{
		DatabasePath: filepath.Join(t.TempDir(), "coffe.db"),
		AIBaseURL:    "http://127.0.0.1:11434/v1",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	first, err := bootstrap.Open(ctx, options, logger)
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	userProvider, err := first.App.CreateAIProvider(ctx, service.CreateAIProviderInput{
		Name: "configured-fallback", Kind: "openai", BaseURL: "http://127.0.0.1:11435/v1",
		Model: "ordinary-model", Priority: 10, Enabled: true,
	})
	if err != nil {
		t.Fatalf("CreateAIProvider() error = %v", err)
	}
	internalNameProvider, err := first.App.CreateAIProvider(ctx, service.CreateAIProviderInput{
		Name: "coffe-internal-fallback", Kind: "openai", BaseURL: "http://127.0.0.1:11436/v1",
		Model: "second-model", Priority: 10, Enabled: true,
	})
	if err != nil {
		t.Fatalf("CreateAIProvider() for internal name error = %v", err)
	}
	if err := first.Close(ctx); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	options.AIModel = "configured-model"

	// When
	second, err := bootstrap.Open(ctx, options, logger)
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	providers, err := second.App.ListAIProviders(ctx)
	if err != nil {
		t.Fatalf("ListAIProviders() error = %v", err)
	}
	if len(providers) != 3 {
		t.Fatalf("provider count = %d, want 3", len(providers))
	}
	fallback := providers[2]
	if fallback.ID() != "configured-fallback" || fallback.Name().String() == userProvider.Name().String() ||
		fallback.Name().String() == internalNameProvider.Name().String() {
		t.Fatalf("providers = %#v", providers)
	}
	if err := second.Close(ctx); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	options.AIModel = "updated-model"
	third, err := bootstrap.Open(ctx, options, logger)
	if err != nil {
		t.Fatalf("third Open() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := third.Close(ctx); closeErr != nil {
			t.Errorf("third Close() error = %v", closeErr)
		}
	})
	providers, err = third.App.ListAIProviders(ctx)

	// Then
	if err != nil {
		t.Fatalf("ListAIProviders() after restart error = %v", err)
	}
	if len(providers) != 3 {
		t.Fatalf("provider count after restart = %d, want 3", len(providers))
	}
	updated := providers[2]
	if updated.ID() != fallback.ID() || updated.Name() != fallback.Name() ||
		!updated.CreatedAt().Equal(fallback.CreatedAt()) || updated.Model() != options.AIModel ||
		updated.Priority() != math.MaxInt || !updated.Enabled() {
		t.Fatalf("updated fallback = %#v, previous = %#v", updated, fallback)
	}
}

func Test_Open_disables_configured_fallback_when_model_is_removed(t *testing.T) {
	// Given
	ctx := context.Background()
	options := bootstrap.Options{
		DatabasePath: filepath.Join(t.TempDir(), "coffe.db"),
		AIBaseURL:    "http://127.0.0.1:11434/v1",
		AIModel:      "configured-model",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	first, err := bootstrap.Open(ctx, options, logger)
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	providers, err := first.App.ListAIProviders(ctx)
	if err != nil || len(providers) != 1 {
		t.Fatalf("initial providers/error = %#v/%v", providers, err)
	}
	configured := providers[0]
	if err := first.Close(ctx); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	options.AIModel = ""

	// When
	second, err := bootstrap.Open(ctx, options, logger)
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := second.Close(ctx); closeErr != nil {
			t.Errorf("second Close() error = %v", closeErr)
		}
	})
	providers, err = second.App.ListAIProviders(ctx)

	// Then
	if err != nil {
		t.Fatalf("ListAIProviders() error = %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("provider count = %d, want 1", len(providers))
	}
	disabled := providers[0]
	if disabled.ID() != configured.ID() || disabled.Name() != configured.Name() ||
		!disabled.CreatedAt().Equal(configured.CreatedAt()) || disabled.Enabled() {
		t.Fatalf("disabled fallback = %#v, configured = %#v", disabled, configured)
	}
}
