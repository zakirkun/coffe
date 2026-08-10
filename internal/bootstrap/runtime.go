package bootstrap

import (
	"coffe/internal/agentexec"
	"coffe/internal/ai"
	"coffe/internal/config"
	"coffe/internal/dbtool"
	"coffe/internal/deploy"
	"coffe/internal/domain"
	"coffe/internal/gittool"
	"coffe/internal/pluginexec"
	"coffe/internal/secrets"
	"coffe/internal/service"
	"coffe/internal/worker"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"time"

	coffesqlite "coffe/internal/store/sqlite"

	"github.com/google/uuid"
)

type Options struct {
	DatabasePath string
	AIBaseURL    string
	AIModel      string
	AISecretName string
	HTTPAddr     string
}

func DefaultOptions() (Options, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Options{}, fmt.Errorf("user config directory: %w", err)
	}
	return Options{
		DatabasePath: filepath.Join(configDir, "coffe", "coffe.db"),
		AIBaseURL:    "http://127.0.0.1:11434/v1",
		HTTPAddr:     "127.0.0.1:4422",
	}, nil
}

type Runtime struct {
	App   *service.App
	store *coffesqlite.Store
}

func Open(ctx context.Context, options Options, logger *slog.Logger) (*Runtime, error) {
	if options.DatabasePath == "" {
		return nil, fmt.Errorf("database path: empty")
	}
	if err := os.MkdirAll(filepath.Dir(options.DatabasePath), 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	store, err := coffesqlite.Open(ctx, options.DatabasePath)
	if err != nil {
		return nil, err
	}
	keyring, err := secrets.NewKeyring("coffe")
	if err != nil {
		return nil, errors.Join(err, store.Close())
	}
	providers, err := store.ListAIProviders(ctx)
	if err != nil {
		return nil, errors.Join(err, store.Close())
	}
	fallback, saveFallback, err := configuredFallbackProvider(options, providers, realClock{}.Now())
	if err != nil {
		return nil, errors.Join(err, store.Close())
	}
	if saveFallback {
		if err := store.SaveAIProvider(ctx, fallback); err != nil {
			return nil, errors.Join(err, store.Close())
		}
	}
	router := ai.NewRouter(ai.RouterDependencies{
		Store: store, Secrets: keyring, IDs: uuidSource{}, Clock: realClock{},
	})
	workerClient := worker.NewClient(worker.ClientConfig{})
	remoteExecutor := worker.NewHybridExecutor(worker.HybridDependencies{
		Registry: store, Secrets: keyring, Client: workerClient,
		Fallback: deploy.NewExecutor(deploy.BinaryRunner{}),
	})
	app := service.New(service.Dependencies{
		Logger:     logger,
		Repository: store,
		Workflows:  config.WorkflowLoader{},
		Planner:    router,
		Executor:   remoteExecutor,
		Renderer:   deploy.CommandRenderer{},
		Swarm:      deploy.NewSwarmInspector(deploy.BinaryRunner{}),
		Agents:     agentexec.Runner{},
		Databases:  dbtool.NewReader(15*time.Second, 200),
		Secrets:    keyring,
		Git:        gittool.Tool{},
		AIModels:   router,
		Workers:    workerClient,
		Plugins:    pluginexec.NewRunner(30 * time.Second),
		Clock:      realClock{},
		IDs:        uuidSource{},
	})
	if err := app.EnsureDefaultPolicy(ctx); err != nil {
		return nil, errors.Join(err, app.Shutdown(context.WithoutCancel(ctx)), store.Close())
	}
	if err := app.PruneAudit(ctx); err != nil {
		return nil, errors.Join(err, app.Shutdown(context.WithoutCancel(ctx)), store.Close())
	}
	if err := app.RecoverInterruptedPlans(ctx); err != nil {
		return nil, errors.Join(err, app.Shutdown(context.WithoutCancel(ctx)), store.Close())
	}
	if err := app.RecoverAgentSessions(ctx); err != nil {
		return nil, errors.Join(err, app.Shutdown(context.WithoutCancel(ctx)), store.Close())
	}
	return &Runtime{App: app, store: store}, nil
}

func (r *Runtime) Close(parent context.Context) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 10*time.Second)
	defer cancel()
	return errors.Join(r.App.Shutdown(ctx), r.store.Close())
}

const (
	configuredFallbackID   domain.AIProviderID = "configured-fallback"
	configuredFallbackName                     = "coffe-internal-fallback"
)

func configuredFallbackProvider(
	options Options,
	providers []domain.AIProvider,
	createdAt time.Time,
) (domain.AIProvider, bool, error) {
	names := make(map[string]struct{}, len(providers))
	var stored domain.AIProvider
	var found bool
	for _, provider := range providers {
		names[provider.Name().String()] = struct{}{}
		if provider.ID() == configuredFallbackID {
			stored = provider
			found = true
		}
	}
	if found && options.AIModel == "" {
		disabled, err := domain.NewAIProvider(domain.AIProviderValues{
			ID: stored.ID(), Name: stored.Name(), Kind: stored.Kind(), BaseURL: stored.BaseURL(),
			Model: stored.Model(), SecretName: stored.SecretName(), Priority: stored.Priority(),
			Enabled: false, CreatedAt: stored.CreatedAt(),
		})
		return disabled, true, err
	}
	if options.AIModel == "" {
		return domain.AIProvider{}, false, nil
	}
	name := stored.Name()
	if !found {
		rawName := configuredFallbackName
		for suffix := 2; ; suffix++ {
			if _, occupied := names[rawName]; !occupied {
				break
			}
			rawName = fmt.Sprintf("%s-%d", configuredFallbackName, suffix)
		}
		var err error
		name, err = domain.NewName(rawName)
		if err != nil {
			return domain.AIProvider{}, false, err
		}
	} else {
		createdAt = stored.CreatedAt()
	}
	provider, err := domain.NewAIProvider(domain.AIProviderValues{
		ID: configuredFallbackID, Name: name, Kind: domain.AIProviderOpenAI,
		BaseURL: options.AIBaseURL, Model: options.AIModel, SecretName: options.AISecretName,
		Priority: math.MaxInt, Enabled: true, CreatedAt: createdAt,
	})
	if err != nil {
		return domain.AIProvider{}, false, err
	}
	return provider, true, nil
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type uuidSource struct{}

func (uuidSource) New() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate UUIDv7: %w", err)
	}
	return id.String(), nil
}
