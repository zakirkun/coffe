package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	listPluginsInput struct{}
	pluginOutput     struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Methods []string `json:"methods"`
		Enabled bool     `json:"enabled"`
	}
)

type pluginsOutput struct {
	Plugins []pluginOutput `json:"plugins"`
}

func registerPluginTools(server *mcp.Server, app Application) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_plugins", Description: "List hash-pinned plugin manifests; invocation remains operator-only",
	}, listPluginsHandler(app))
}

func listPluginsHandler(app Application) mcp.ToolHandlerFor[listPluginsInput, pluginsOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ listPluginsInput,
	) (*mcp.CallToolResult, pluginsOutput, error) {
		plugins, err := app.ListPlugins(ctx)
		if err != nil {
			return nil, pluginsOutput{}, fmt.Errorf("list plugins: %w", err)
		}
		output := make([]pluginOutput, 0, len(plugins))
		for _, plugin := range plugins {
			output = append(output, pluginOutput{
				ID: string(plugin.ID()), Name: plugin.Name().String(),
				Methods: plugin.Methods(), Enabled: plugin.Enabled(),
			})
		}
		return nil, pluginsOutput{Plugins: output}, nil
	}
}
