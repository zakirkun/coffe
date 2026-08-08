package pluginexec_test

import (
	"coffe/internal/domain"
	"coffe/internal/pluginexec"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if os.Getenv("COFFE_PLUGIN_PROTOCOL") == "1" {
		var request struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
			os.Exit(2)
		}
		if err := json.NewEncoder(os.Stdout).Encode(map[string]any{
			"result": map[string]string{"method": request.Method},
		}); err != nil {
			os.Exit(3)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func Test_HashFile_changes_when_plugin_binary_changes(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "plugin.bin")
	if err := os.WriteFile(path, []byte("version-1"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	first, err := pluginexec.HashFile(path)
	if err != nil {
		t.Fatalf("HashFile(first) error = %v", err)
	}
	if err := os.WriteFile(path, []byte("version-2"), 0o600); err != nil {
		t.Fatalf("WriteFile(second) error = %v", err)
	}

	// When
	second, err := pluginexec.HashFile(path)
	// Then
	if err != nil {
		t.Fatalf("HashFile(second) error = %v", err)
	}
	if first == second {
		t.Fatal("plugin hash did not change")
	}
}

func Test_Runner_invokes_hash_pinned_JSON_protocol(t *testing.T) {
	// Given
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("Executable() error = %v", err)
	}
	hash, err := pluginexec.HashFile(executable)
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}
	name, err := domain.NewName("test-plugin")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}
	plugin, err := domain.NewPlugin(domain.PluginValues{
		ID: "plugin-1", Name: name, ExecutablePath: executable, SHA256: hash,
		Methods: []string{"test.echo"}, Enabled: true,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewPlugin() error = %v", err)
	}

	// When
	result, err := pluginexec.NewRunner(5*time.Second).Invoke(
		context.Background(), plugin, "test.echo", json.RawMessage(`{"value":1}`),
	)
	// Then
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if string(result) != `{"method":"test.echo"}` {
		t.Fatalf("Invoke() result = %s", result)
	}
}
