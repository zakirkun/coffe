package domain_test

import (
	"coffe/internal/domain"
	"path/filepath"
	"testing"
	"time"
)

func Test_NewPlugin_accepts_hash_pinned_method_manifest(t *testing.T) {
	// Given
	name, err := domain.NewName("incident-tools")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}

	// When
	plugin, err := domain.NewPlugin(domain.PluginValues{
		ID: "plugin-1", Name: name, ExecutablePath: filepath.Join(t.TempDir(), "plugin.exe"),
		SHA256:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Methods: []string{"incident.lookup", "incident.timeline"}, Enabled: true,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	// Then
	if err != nil {
		t.Fatalf("NewPlugin() error = %v", err)
	}
	if !plugin.Allows("incident.lookup") || plugin.Allows("shell") {
		t.Fatalf("plugin method allowlist = %v", plugin.Methods())
	}
}
