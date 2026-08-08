package command

import (
	"coffe/internal/bootstrap"
	"coffe/internal/service"
	"coffe/internal/version"
	"coffe/internal/worker"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newWorkerCommand(state *rootState) *cobra.Command {
	workerCommand := &cobra.Command{Use: "worker", Short: "Run the controlled remote execution worker"}
	var config worker.RunConfig
	var tokenEnv string
	serve := &cobra.Command{
		Use: "serve", Args: cobra.NoArgs, Short: "Serve authenticated worker health and task APIs",
		RunE: func(_ *cobra.Command, _ []string) error {
			config.Token = os.Getenv(tokenEnv)
			if config.Token == "" {
				return fmt.Errorf("worker token environment variable %s is empty", tokenEnv)
			}
			return worker.Run(state.ctx, config)
		},
	}
	flags := serve.Flags()
	flags.StringVar(&config.Addr, "addr", "127.0.0.1:4444", "worker listen address")
	flags.StringVar(&tokenEnv, "token-env", "COFFE_WORKER_TOKEN", "environment variable containing bearer token")
	flags.StringVar(&config.TLSCert, "tls-cert", "", "TLS certificate path")
	flags.StringVar(&config.TLSKey, "tls-key", "", "TLS private key path")
	flags.BoolVar(&config.InsecureLoopback, "insecure-loopback", false, "allow HTTP only on loopback")
	config.Version = version.Version
	workerCommand.AddCommand(
		serve, newWorkerRegisterCommand(state), newWorkerListCommand(state), newWorkerHeartbeatCommand(state),
	)
	return workerCommand
}

func newWorkerRegisterCommand(state *rootState) *cobra.Command {
	var input service.RegisterWorkerInput
	command := &cobra.Command{
		Use: "register", Args: cobra.NoArgs, Short: "Map a target to an HTTPS worker",
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				registered, err := runtime.App.RegisterWorker(state.ctx, input)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(command.OutOrStdout(), registered.ID()); err != nil {
					return fmt.Errorf("write worker id: %w", err)
				}
				return nil
			})
		},
	}
	flags := command.Flags()
	flags.StringVar(&input.TargetID, "target", "", "registered target ID")
	flags.StringVar(&input.Name, "name", "", "worker name")
	flags.StringVar(&input.BaseURL, "url", "", "HTTPS worker base URL")
	flags.StringVar(&input.SecretName, "secret", "", "worker token name in OS keyring")
	flags.BoolVar(&input.Enabled, "enabled", true, "route target execution through worker")
	return command
}

func newWorkerListCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "list", Args: cobra.NoArgs, Short: "List registered workers",
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				workers, err := runtime.App.ListWorkers(state.ctx)
				if err != nil {
					return err
				}
				for _, registered := range workers {
					if _, err := fmt.Fprintf(command.OutOrStdout(), "%s %s %s %s %v\n",
						registered.ID(), registered.TargetID(), registered.Name().String(),
						registered.Version(), registered.Enabled(),
					); err != nil {
						return fmt.Errorf("write worker: %w", err)
					}
				}
				return nil
			})
		},
	}
}

func newWorkerHeartbeatCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "heartbeat WORKER_ID", Args: cobra.ExactArgs(1), Short: "Probe worker health",
		RunE: func(command *cobra.Command, args []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				updated, err := runtime.App.HeartbeatWorker(state.ctx, args[0])
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(command.OutOrStdout(), updated.Version()); err != nil {
					return fmt.Errorf("write worker version: %w", err)
				}
				return nil
			})
		},
	}
}
