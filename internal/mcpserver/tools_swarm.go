package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type swarmTargetInput struct {
	TargetID string `json:"target_id" jsonschema:"registered deployment target identifier"`
}

type swarmNodeOutput struct {
	ID           string `json:"id"`
	Hostname     string `json:"hostname"`
	Status       string `json:"status"`
	Availability string `json:"availability"`
	Manager      bool   `json:"manager"`
}

type swarmNodesOutput struct {
	Nodes []swarmNodeOutput `json:"nodes"`
}

type swarmServiceOutput struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Mode     string `json:"mode"`
	Replicas string `json:"replicas"`
	Image    string `json:"image"`
}

type swarmServicesOutput struct {
	Services []swarmServiceOutput `json:"services"`
}

func registerSwarmTools(server *mcp.Server, app Application) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_swarm_nodes", Description: "List nodes from a target Swarm manager",
	}, listSwarmNodesHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_swarm_services", Description: "List services for the target's configured stack",
	}, listSwarmServicesHandler(app))
}

func listSwarmNodesHandler(app Application) mcp.ToolHandlerFor[swarmTargetInput, swarmNodesOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input swarmTargetInput,
	) (*mcp.CallToolResult, swarmNodesOutput, error) {
		nodes, err := app.ListSwarmNodes(ctx, input.TargetID)
		if err != nil {
			return nil, swarmNodesOutput{}, fmt.Errorf("list swarm nodes: %w", err)
		}
		output := make([]swarmNodeOutput, 0, len(nodes))
		for _, node := range nodes {
			output = append(output, swarmNodeOutput{
				ID: node.ID(), Hostname: node.Hostname(), Status: node.Status(),
				Availability: node.Availability(), Manager: node.Manager(),
			})
		}
		return nil, swarmNodesOutput{Nodes: output}, nil
	}
}

func listSwarmServicesHandler(app Application) mcp.ToolHandlerFor[swarmTargetInput, swarmServicesOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input swarmTargetInput,
	) (*mcp.CallToolResult, swarmServicesOutput, error) {
		services, err := app.ListSwarmServices(ctx, input.TargetID)
		if err != nil {
			return nil, swarmServicesOutput{}, fmt.Errorf("list swarm services: %w", err)
		}
		output := make([]swarmServiceOutput, 0, len(services))
		for _, service := range services {
			output = append(output, swarmServiceOutput{
				ID: service.ID(), Name: service.Name(), Mode: service.Mode(),
				Replicas: service.Replicas(), Image: service.Image(),
			})
		}
		return nil, swarmServicesOutput{Services: output}, nil
	}
}
