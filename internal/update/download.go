package update

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxAssetBytes int64 = 512 * 1024 * 1024

func (c *Client) Download(ctx context.Context, release Release, destination string) (err error) {
	plan, err := prepareDownload(release, destination)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, release.Asset.URL, nil)
	if err != nil {
		return fmt.Errorf("create asset request: %w", err)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("download update asset: %w", err)
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("asset response status %d", response.StatusCode)
	}
	return stageAsset(response.Body, plan)
}

type downloadPlan struct {
	destination string
	expected    []byte
}

func prepareDownload(release Release, destination string) (downloadPlan, error) {
	if !release.Available {
		return downloadPlan{}, fmt.Errorf("release %s is not newer: %w", release.Version, ErrInvalidVersion)
	}
	if _, err := parseHTTPSURL(release.Asset.URL); err != nil {
		return downloadPlan{}, err
	}
	clean := filepath.Clean(destination)
	if !filepath.IsAbs(clean) {
		return downloadPlan{}, fmt.Errorf("update destination must be absolute: %w", ErrInvalidManifest)
	}
	expected, err := hex.DecodeString(release.Asset.SHA256)
	if err != nil || len(expected) != sha256.Size {
		return downloadPlan{}, fmt.Errorf("asset checksum: %w", ErrInvalidManifest)
	}
	return downloadPlan{destination: clean, expected: expected}, nil
}

func stageAsset(source io.Reader, plan downloadPlan) (err error) {
	clean := plan.destination
	if _, err := os.Stat(clean); err == nil {
		return fmt.Errorf("update destination already exists: %w", ErrInvalidManifest)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect update destination: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(clean), ".coffe-update-*")
	if err != nil {
		return fmt.Errorf("create staged update: %w", err)
	}
	temporaryPath := temporary.Name()
	temporaryOpen := true
	succeeded := false
	defer func() {
		if temporaryOpen {
			err = errors.Join(err, temporary.Close())
		}
		if !succeeded {
			err = errors.Join(err, removeIfExists(temporaryPath))
		}
	}()
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(temporary, hash), io.LimitReader(source, maxAssetBytes+1))
	if err != nil {
		return fmt.Errorf("write staged update: %w", err)
	}
	if written > maxAssetBytes {
		return fmt.Errorf("asset exceeds size limit: %w", ErrInvalidManifest)
	}
	if !bytes.Equal(hash.Sum(nil), plan.expected) {
		return fmt.Errorf("asset checksum mismatch: %w", ErrInvalidManifest)
	}
	if err := temporary.Chmod(0o755); err != nil {
		return fmt.Errorf("mark staged update executable: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close staged update: %w", err)
	}
	temporaryOpen = false
	if err := os.Rename(temporaryPath, clean); err != nil {
		return fmt.Errorf("install staged update: %w", err)
	}
	succeeded = true
	return nil
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
