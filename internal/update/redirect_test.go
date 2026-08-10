package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

func TestClientCheckRejectsHTTPRedirect(t *testing.T) {
	// Given
	var targetRequests atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		targetRequests.Add(1)
	}))
	t.Cleanup(target.Close)
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL, http.StatusFound)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(ClientConfig{ManifestURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// When
	_, err = client.Check(context.Background(), "v1.0.0")

	// Then
	if !errors.Is(err, ErrInvalidManifest) {
		t.Fatalf("Check() error = %v, want ErrInvalidManifest", err)
	}
	if targetRequests.Load() != 0 {
		t.Fatal("HTTP redirect target was requested")
	}
}

func TestClientDownloadRejectsHTTPRedirect(t *testing.T) {
	// Given
	asset := []byte("coffe-binary")
	checksum := sha256.Sum256(asset)
	var targetRequests atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		targetRequests.Add(1)
	}))
	t.Cleanup(target.Close)
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/manifest.json" {
			manifest := Manifest{
				Version: "v1.2.0",
				Assets: map[string]Asset{
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
		http.Redirect(writer, request, target.URL, http.StatusFound)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(ClientConfig{ManifestURL: server.URL + "/manifest.json", HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	release, err := client.Check(context.Background(), "v1.1.0")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	// When
	err = client.Download(context.Background(), release, filepath.Join(t.TempDir(), "coffe.new"))

	// Then
	if !errors.Is(err, ErrInvalidManifest) {
		t.Fatalf("Download() error = %v, want ErrInvalidManifest", err)
	}
	if targetRequests.Load() != 0 {
		t.Fatal("HTTP redirect target was requested")
	}
}

func TestNewClientComposesRedirectPolicyWithoutMutatingCaller(t *testing.T) {
	// Given
	redirectPolicyErr := errors.New("redirect policy")
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/manifest.json", http.StatusFound)
	}))
	t.Cleanup(server.Close)
	caller := server.Client()
	caller.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		return redirectPolicyErr
	}
	client, err := NewClient(ClientConfig{ManifestURL: server.URL, HTTPClient: caller})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client.httpClient == caller {
		t.Fatal("NewClient() retained the caller HTTP client")
	}

	// When
	_, err = client.Check(context.Background(), "v1.0.0")

	// Then
	if !errors.Is(err, redirectPolicyErr) {
		t.Fatalf("Check() error = %v, want redirect policy error", err)
	}
}

func TestClientCheckStopsHTTPSRedirectLoopAtDefaultLimit(t *testing.T) {
	// Given
	requests := 0
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		if requests > 10 {
			writer.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(writer, request, server.URL, http.StatusFound)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(ClientConfig{ManifestURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// When
	_, err = client.Check(context.Background(), "v1.0.0")

	// Then
	if err == nil || !strings.Contains(err.Error(), "stopped after 10 redirects") {
		t.Fatalf("Check() error = %v, want default redirect limit", err)
	}
	if requests != 10 {
		t.Fatalf("redirect requests = %d, want 10", requests)
	}
}
