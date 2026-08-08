package mcpserver

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listAIProvidersInput struct{}

type aiProviderOutput struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	Kind     domain.AIProviderKind `json:"kind"`
	BaseURL  string                `json:"base_url"`
	Model    string                `json:"model"`
	Priority int                   `json:"priority"`
	Enabled  bool                  `json:"enabled"`
}

type aiProvidersOutput struct {
	Providers []aiProviderOutput `json:"providers"`
}

type aiProviderInput struct {
	ProviderID string `json:"provider_id" jsonschema:"registered AI provider identifier"`
}

type (
	aiModelsOutput struct {
		Models []string `json:"models"`
	}
	aiUsageInput  struct{}
	aiUsageOutput struct {
		Usage []aiUsageRecord `json:"usage"`
	}
)

type aiUsageRecord struct {
	ProviderID       string    `json:"provider_id"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	LatencyMillis    int64     `json:"latency_millis"`
	Success          bool      `json:"success"`
	CreatedAt        time.Time `json:"created_at"`
}

func registerAIProviderTools(server *mcp.Server, app Application) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_ai_providers", Description: "List AI provider routing metadata without credentials",
	}, listAIProvidersHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_ai_models", Description: "Discover models from one OpenAI-compatible provider",
	}, listAIModelsHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_ai_usage", Description: "List provider token and latency usage records",
	}, listAIUsageHandler(app))
}

func listAIProvidersHandler(app Application) mcp.ToolHandlerFor[listAIProvidersInput, aiProvidersOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ listAIProvidersInput,
	) (*mcp.CallToolResult, aiProvidersOutput, error) {
		providers, err := app.ListAIProviders(ctx)
		if err != nil {
			return nil, aiProvidersOutput{}, fmt.Errorf("list AI providers: %w", err)
		}
		output := make([]aiProviderOutput, 0, len(providers))
		for _, provider := range providers {
			output = append(output, aiProviderOutput{
				ID: string(provider.ID()), Name: provider.Name().String(), Kind: provider.Kind(),
				BaseURL: provider.BaseURL(), Model: provider.Model(),
				Priority: provider.Priority(), Enabled: provider.Enabled(),
			})
		}
		return nil, aiProvidersOutput{Providers: output}, nil
	}
}

func listAIModelsHandler(app Application) mcp.ToolHandlerFor[aiProviderInput, aiModelsOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input aiProviderInput,
	) (*mcp.CallToolResult, aiModelsOutput, error) {
		models, err := app.ListAIModels(ctx, input.ProviderID)
		if err != nil {
			return nil, aiModelsOutput{}, fmt.Errorf("list AI models: %w", err)
		}
		return nil, aiModelsOutput{Models: models}, nil
	}
}

func listAIUsageHandler(app Application) mcp.ToolHandlerFor[aiUsageInput, aiUsageOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ aiUsageInput,
	) (*mcp.CallToolResult, aiUsageOutput, error) {
		usage, err := app.ListAIUsage(ctx, 100)
		if err != nil {
			return nil, aiUsageOutput{}, fmt.Errorf("list AI usage: %w", err)
		}
		output := make([]aiUsageRecord, 0, len(usage))
		for _, record := range usage {
			output = append(output, aiUsageRecord{
				ProviderID: string(record.ProviderID()), Model: record.Model(),
				PromptTokens: record.PromptTokens(), CompletionTokens: record.CompletionTokens(),
				TotalTokens: record.TotalTokens(), LatencyMillis: record.LatencyMillis(),
				Success: record.Success(), CreatedAt: record.CreatedAt(),
			})
		}
		return nil, aiUsageOutput{Usage: output}, nil
	}
}
