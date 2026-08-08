package update_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	coffeupdate "coffe/internal/update"
)

func Test_Client_checks_and_stages_verified_asset(t *testing.T) {
	// Given
	asset := []byte("coffe-binary")
	checksum := sha256.Sum256(asset)
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/manifest.json" {
			manifest := coffeupdate.Manifest{
				Version: "v1.2.0",
				Assets: map[string]coffeupdate.Asset{
					runtime.GOOS + "-" + runtime.GOARCH: {
						URL: server.URL + "/coffe", SHA256: hex.EncodeToString(checksum[:]),
					},
				},
			}
			if err := json.NewEncoder(writer).Encode(manifest); err != nil {
				t.Errorf("Encode() error = %v", err)
			}
			return
		}
		if _, err := writer.Write(asset); err != nil {
			t.Errorf("Write() error = %v", err)
		}
	}))
	t.Cleanup(server.Close)
	client, err := coffeupdate.NewClient(coffeupdate.ClientConfig{
		ManifestURL: server.URL + "/manifest.json", HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	destination := filepath.Join(t.TempDir(), "coffe.new")

	// When
	release, err := client.Check(context.Background(), "v1.1.0")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	err = client.Download(context.Background(), release, destination)
	// Then
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	got, err := os.ReadFile(destination) //nolint:gosec // destination is created inside t.TempDir.
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !release.Available || string(got) != string(asset) {
		t.Fatalf("release/content = %#v/%q", release, got)
	}
}
