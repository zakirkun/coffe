package command

import (
	"coffe/internal/bootstrap"
	"coffe/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

func newAICommand(state *rootState) *cobra.Command {
	ai := &cobra.Command{Use: "ai", Short: "Manage AI providers and usage"}
	ai.AddCommand(newAIProviderAddCommand(state), newAIProviderListCommand(state), newAIModelsCommand(state))
	return ai
}

func newAIProviderAddCommand(state *rootState) *cobra.Command {
	var input service.CreateAIProviderInput
	command := &cobra.Command{
		Use: "add", Args: cobra.NoArgs, Short: "Register an OpenAI-compatible provider",
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				provider, err := runtime.App.CreateAIProvider(state.ctx, input)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(command.OutOrStdout(), provider.ID()); err != nil {
					return fmt.Errorf("write provider id: %w", err)
				}
				return nil
			})
		},
	}
	flags := command.Flags()
	flags.StringVar(&input.Name, "name", "", "provider name")
	flags.StringVar(&input.Kind, "kind", "openai", "openai, ollama, vllm, or lmstudio")
	flags.StringVar(&input.BaseURL, "url", "", "OpenAI-compatible base URL ending in /v1")
	flags.StringVar(&input.Model, "model", "", "default model")
	flags.StringVar(&input.SecretName, "secret", "", "API key name in the OS keyring")
	flags.IntVar(&input.Priority, "priority", 100, "lower values are attempted first")
	flags.BoolVar(&input.Enabled, "enabled", true, "enable provider routing")
	return command
}

func newAIProviderListCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "list", Args: cobra.NoArgs, Short: "List provider routing order",
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				providers, err := runtime.App.ListAIProviders(state.ctx)
				if err != nil {
					return err
				}
				for _, provider := range providers {
					if _, err := fmt.Fprintf(command.OutOrStdout(), "%s %d %s %s %s\n",
						provider.ID(), provider.Priority(), provider.Kind(), provider.Name().String(), provider.Model(),
					); err != nil {
						return fmt.Errorf("write provider: %w", err)
					}
				}
				return nil
			})
		},
	}
}

func newAIModelsCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "models PROVIDER_ID", Args: cobra.ExactArgs(1), Short: "Discover provider models",
		RunE: func(command *cobra.Command, args []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				models, err := runtime.App.ListAIModels(state.ctx, args[0])
				if err != nil {
					return err
				}
				for _, model := range models {
					if _, err := fmt.Fprintln(command.OutOrStdout(), model); err != nil {
						return fmt.Errorf("write model: %w", err)
					}
				}
				return nil
			})
		},
	}
}
