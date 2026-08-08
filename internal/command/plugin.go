package command

import (
	"coffe/internal/bootstrap"
	"coffe/internal/pluginexec"
	"coffe/internal/service"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newPluginCommand(state *rootState) *cobra.Command {
	command := &cobra.Command{Use: "plugin", Short: "Manage hash-pinned JSON-stdio plugins"}
	command.AddCommand(newPluginRegisterCommand(state), newPluginListCommand(state), newPluginInvokeCommand(state))
	return command
}

func newPluginRegisterCommand(state *rootState) *cobra.Command {
	var input service.RegisterPluginInput
	var methods string
	command := &cobra.Command{
		Use: "register", Args: cobra.NoArgs, Short: "Hash and register a plugin executable",
		RunE: func(command *cobra.Command, _ []string) error {
			path, err := filepath.Abs(input.ExecutablePath)
			if err != nil {
				return fmt.Errorf("resolve plugin path: %w", err)
			}
			hash, err := pluginexec.HashFile(path)
			if err != nil {
				return err
			}
			input.ExecutablePath = path
			input.SHA256 = hash
			input.Methods = strings.Split(methods, ",")
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				plugin, err := runtime.App.RegisterPlugin(state.ctx, input)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(command.OutOrStdout(), plugin.ID()); err != nil {
					return fmt.Errorf("write plugin id: %w", err)
				}
				return nil
			})
		},
	}
	flags := command.Flags()
	flags.StringVar(&input.Name, "name", "", "plugin name")
	flags.StringVar(&input.ExecutablePath, "path", "", "plugin executable path")
	flags.StringVar(&methods, "methods", "", "comma-separated declared methods")
	flags.BoolVar(&input.Enabled, "enabled", true, "enable plugin invocation")
	return command
}

func newPluginListCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "list", Args: cobra.NoArgs, Short: "List registered plugin manifests",
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				plugins, err := runtime.App.ListPlugins(state.ctx)
				if err != nil {
					return err
				}
				for _, plugin := range plugins {
					if _, err := fmt.Fprintf(command.OutOrStdout(), "%s %s %v %v\n",
						plugin.ID(), plugin.Name().String(), plugin.Methods(), plugin.Enabled(),
					); err != nil {
						return fmt.Errorf("write plugin: %w", err)
					}
				}
				return nil
			})
		},
	}
}

func newPluginInvokeCommand(state *rootState) *cobra.Command {
	var method, input string
	command := &cobra.Command{
		Use: "invoke PLUGIN_ID", Args: cobra.ExactArgs(1), Short: "Invoke one declared plugin method",
		RunE: func(command *cobra.Command, args []string) error {
			payload := json.RawMessage(input)
			if !json.Valid(payload) {
				return fmt.Errorf("plugin input must be valid JSON")
			}
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				result, err := runtime.App.InvokePlugin(state.ctx, service.PluginInvocationInput{
					PluginID: args[0], Method: method, Input: payload,
				})
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(command.OutOrStdout(), string(result)); err != nil {
					return fmt.Errorf("write plugin result: %w", err)
				}
				return nil
			})
		},
	}
	command.Flags().StringVar(&method, "method", "", "declared plugin method")
	command.Flags().StringVar(&input, "input", "{}", "JSON input object")
	return command
}
