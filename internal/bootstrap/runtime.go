package bootstrap

import (
	"coffe/internal/agentexec"
	"coffe/internal/ai"
	"coffe/internal/config"
	"coffe/internal/dbtool"
	"coffe/internal/deploy"
	"coffe/internal/gittool"
	"coffe/internal/pluginexec"
	"coffe/internal/secrets"
	"coffe/internal/service"
	"coffe/internal/worker"
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	fallback, err := buildFallbackPlanner(ctx, options, keyring)
	if err != nil {
		return nil, errors.Join(err, store.Close())
	}
	router := ai.NewRouter(ai.RouterDependencies{
		Store: store, Secrets: keyring, IDs: uuidSource{}, Clock: realClock{}, Fallback: fallback.planner,
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

type fallbackPlanner struct{ planner *ai.OpenAIPlanner }

func buildFallbackPlanner(
	ctx context.Context,
	options Options,
	keyring *secrets.Keyring,
) (fallbackPlanner, error) {
	if options.AIModel == "" {
		return fallbackPlanner{}, nil
	}
	apiKey := ""
	if options.AISecretName != "" {
		var err error
		apiKey, err = keyring.Get(ctx, options.AISecretName)
		if err != nil && !errors.Is(err, secrets.ErrNotFound) {
			return fallbackPlanner{}, err
		}
	}
	planner, err := ai.NewOpenAIPlanner(ai.OpenAIConfig{
		BaseURL: options.AIBaseURL, Model: options.AIModel, APIKey: apiKey, Timeout: 45 * time.Second,
	})
	if err != nil {
		return fallbackPlanner{}, err
	}
	return fallbackPlanner{planner: planner}, nil
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
