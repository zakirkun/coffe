package update

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"golang.org/x/mod/semver"
)

const maxManifestBytes = 1024 * 1024

var (
	ErrInvalidManifest = errors.New("update: invalid manifest")
	ErrInvalidVersion  = errors.New("update: invalid version")
)

type Asset struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type Manifest struct {
	Version string           `json:"version"`
	Assets  map[string]Asset `json:"assets"`
}

type Release struct {
	Version   string
	Asset     Asset
	Available bool
}

type ClientConfig struct {
	ManifestURL string
	HTTPClient  *http.Client
}

type Client struct {
	manifestURL string
	httpClient  *http.Client
}

func NewClient(config ClientConfig) (*Client, error) {
	manifestURL, err := parseHTTPSURL(config.ManifestURL)
	if err != nil {
		return nil, err
	}
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	clonedClient := *client
	callerRedirect := clonedClient.CheckRedirect
	clonedClient.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if _, err := parseHTTPSURL(request.URL.String()); err != nil {
			return err
		}
		if callerRedirect != nil {
			return callerRedirect(request, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
	return &Client{manifestURL: manifestURL, httpClient: &clonedClient}, nil
}

func (c *Client) Check(ctx context.Context, currentVersion string) (release Release, err error) {
	if !semver.IsValid(currentVersion) {
		return Release{}, fmt.Errorf("current version %q: %w", currentVersion, ErrInvalidVersion)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.manifestURL, nil)
	if err != nil {
		return Release{}, fmt.Errorf("create manifest request: %w", err)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return Release{}, fmt.Errorf("fetch update manifest: %w", err)
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	if response.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("manifest response status %d", response.StatusCode)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, maxManifestBytes+1))
	if err != nil {
		return Release{}, fmt.Errorf("read update manifest: %w", err)
	}
	if len(content) > maxManifestBytes {
		return Release{}, fmt.Errorf("manifest exceeds size limit: %w", ErrInvalidManifest)
	}
	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Release{}, fmt.Errorf("decode update manifest: %w", err)
	}
	if !semver.IsValid(manifest.Version) {
		return Release{}, fmt.Errorf("manifest version %q: %w", manifest.Version, ErrInvalidManifest)
	}
	asset, ok := manifest.Assets[runtime.GOOS+"-"+runtime.GOARCH]
	if !ok {
		return Release{}, fmt.Errorf("platform asset missing: %w", ErrInvalidManifest)
	}
	if _, err := parseHTTPSURL(asset.URL); err != nil || len(asset.SHA256) != 64 {
		return Release{}, fmt.Errorf("platform asset: %w", ErrInvalidManifest)
	}
	return Release{
		Version: manifest.Version, Asset: asset,
		Available: semver.Compare(manifest.Version, currentVersion) > 0,
	}, nil
}

func parseHTTPSURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return "", fmt.Errorf("URL %q must use HTTPS: %w", raw, ErrInvalidManifest)
	}
	return parsed.String(), nil
}
