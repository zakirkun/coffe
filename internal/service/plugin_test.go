package service_test

import (
	"coffe/internal/service"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
)

func Test_App_registers_and_invokes_declared_plugin_method(t *testing.T) {
	// Given
	fixture := newFixture(t)
	plugin, err := fixture.app.RegisterPlugin(context.Background(), service.RegisterPluginInput{
		Name: "incident-tools", ExecutablePath: filepath.Join(t.TempDir(), "plugin.exe"),
		SHA256:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Methods: []string{"incident.lookup"}, Enabled: true,
	})
	if err != nil {
		t.Fatalf("RegisterPlugin() error = %v", err)
	}

	// When
	result, err := fixture.app.InvokePlugin(context.Background(), service.PluginInvocationInput{
		PluginID: string(plugin.ID()), Method: "incident.lookup", Input: json.RawMessage(`{"id":"INC-1"}`),
	})
	// Then
	if err != nil {
		t.Fatalf("InvokePlugin() error = %v", err)
	}
	if string(result) != `{"ok":true}` {
		t.Fatalf("plugin result = %s", result)
	}
}
