package agentexec_test

import (
	"coffe/internal/agentexec"
	"coffe/internal/domain"
	"testing"
)

func Test_CommandSpec_maps_only_allowlisted_agent_tools(t *testing.T) {
	// Given
	tests := []struct {
		tool       domain.AgentTool
		wantBinary string
		wantPrefix string
	}{
		{tool: domain.AgentOpenCode, wantBinary: "opencode", wantPrefix: "run"},
		{tool: domain.AgentClaude, wantBinary: "claude", wantPrefix: "-p"},
		{tool: domain.AgentCodex, wantBinary: "codex", wantPrefix: "exec"},
		{tool: domain.AgentGemini, wantBinary: "gemini", wantPrefix: "-p"},
		{tool: domain.AgentAider, wantBinary: "aider", wantPrefix: "--message"},
	}

	for _, tt := range tests {
		t.Run(string(tt.tool), func(t *testing.T) {
			// When
			spec, err := agentexec.CommandSpec(tt.tool, "inspect project")
			// Then
			if err != nil {
				t.Fatalf("CommandSpec() error = %v", err)
			}
			if spec.Binary != tt.wantBinary || len(spec.Args) != 2 || spec.Args[0] != tt.wantPrefix {
				t.Fatalf("CommandSpec() = %#v", spec)
			}
		})
	}
}
