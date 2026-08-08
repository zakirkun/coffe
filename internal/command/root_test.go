package command_test

import (
	"bytes"
	"coffe/internal/command"
	"context"
	"strings"
	"testing"
)

func Test_Root_help_lists_runtime_surfaces(t *testing.T) {
	// Given
	root, err := command.NewRoot(context.Background())
	if err != nil {
		t.Fatalf("NewRoot() error = %v", err)
	}
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)
	root.SetArgs([]string{"--help"})

	// When
	err = root.Execute()
	// Then
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, commandName := range []string{
		"tui", "serve", "mcp", "project", "target", "database", "git", "ai", "secret",
		"backup", "update", "version", "policy", "audit",
		"worker", "plugin",
	} {
		if !strings.Contains(output.String(), commandName) {
			t.Fatalf("help missing %q:\n%s", commandName, output.String())
		}
	}
}
