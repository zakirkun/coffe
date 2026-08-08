package command

import (
	"coffe/internal/bootstrap"
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

type rootState struct {
	ctx     context.Context
	options bootstrap.Options
	logger  *slog.Logger
}

func NewRoot(ctx context.Context) (*cobra.Command, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	options, err := bootstrap.DefaultOptions()
	if err != nil {
		return nil, err
	}
	state := &rootState{
		ctx:     ctx,
		options: options,
		logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
	root := &cobra.Command{
		Use: "coffe", Short: "Local-first deployment control plane",
		SilenceUsage: true, SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&state.options.DatabasePath, "db", options.DatabasePath, "SQLite database path")
	root.PersistentFlags().StringVar(&state.options.AIBaseURL, "ai-url", options.AIBaseURL, "OpenAI-compatible base URL")
	root.PersistentFlags().StringVar(&state.options.AIModel, "ai-model", options.AIModel, "OpenAI-compatible model name")
	root.PersistentFlags().StringVar(&state.options.AISecretName, "ai-secret", options.AISecretName, "API key name in OS keyring")
	root.AddCommand(
		newTUICommand(state), newServeCommand(state), newMCPCommand(state),
		newProjectCommand(state), newTargetCommand(state), newDatabaseCommand(state), newSecretCommand(),
		newGitCommand(state),
		newAICommand(state),
		newBackupCommand(state),
		newUpdateCommand(state), newVersionCommand(),
		newPolicyCommand(state), newAuditCommand(state),
		newWorkerCommand(state),
		newPluginCommand(state),
	)
	return root, nil
}

func Execute(ctx context.Context) error {
	root, err := NewRoot(ctx)
	if err != nil {
		return err
	}
	return root.ExecuteContext(ctx)
}

func withRuntime(state *rootState, action func(*bootstrap.Runtime) error) (err error) {
	runtime, err := bootstrap.Open(state.ctx, state.options, state.logger)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, runtime.Close(state.ctx)) }()
	return action(runtime)
}
