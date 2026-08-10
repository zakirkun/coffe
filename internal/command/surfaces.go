package command

import (
	"coffe/internal/bootstrap"
	"coffe/internal/httpapi"
	"coffe/internal/mcpserver"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func newServeCommand(state *rootState) *cobra.Command {
	command := &cobra.Command{
		Use: "serve", Short: "Serve the local HTTP API and MCP endpoint",
		RunE: func(_ *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				api := httpapi.New(runtime.App, state.logger)
				mux := http.NewServeMux()
				mux.Handle("/mcp", mcpserver.HTTPHandler(runtime.App))
				mux.Handle("/", api)
				return httpapi.Run(state.ctx, httpapi.RunConfig{
					Addr: state.options.HTTPAddr, Handler: mux, Logger: state.logger,
					ShutdownTimeout: 10 * time.Second,
				})
			})
		},
	}
	command.Flags().StringVar(&state.options.HTTPAddr, "addr", state.options.HTTPAddr, "HTTP listen address")
	return command
}

func newMCPCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "mcp", Short: "Run MCP over standard input and output",
		RunE: func(_ *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				return mcpserver.RunStdio(state.ctx, runtime.App)
			})
		},
	}
}
