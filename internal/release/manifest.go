package release

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	coffeupdate "coffe/internal/update"
)

var ErrInvalidChecksums = errors.New("release: invalid checksums")

type ManifestOptions struct {
	Version string
	BaseURL string
}

func ParseChecksums(input io.Reader, options ManifestOptions) (coffeupdate.Manifest, error) {
	pattern, err := regexp.Compile(`^coffe_update_` + regexp.QuoteMeta(options.Version) +
		`_(linux|darwin|windows)_(amd64|arm64)(\.exe)?$`)
	if err != nil {
		return coffeupdate.Manifest{}, fmt.Errorf("compile asset pattern: %w", err)
	}
	baseURL := strings.TrimRight(options.BaseURL, "/")
	if options.Version == "" || !strings.HasPrefix(baseURL, "https://") {
		return coffeupdate.Manifest{}, ErrInvalidChecksums
	}
	manifest := coffeupdate.Manifest{Version: options.Version, Assets: make(map[string]coffeupdate.Asset)}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || len(fields[0]) != 64 {
			return coffeupdate.Manifest{}, ErrInvalidChecksums
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return coffeupdate.Manifest{}, ErrInvalidChecksums
		}
		matches := pattern.FindStringSubmatch(fields[1])
		if len(matches) == 0 {
			continue
		}
		manifest.Assets[matches[1]+"-"+matches[2]] = coffeupdate.Asset{
			URL: baseURL + "/" + fields[1], SHA256: fields[0],
		}
	}
	if err := scanner.Err(); err != nil {
		return coffeupdate.Manifest{}, fmt.Errorf("scan checksums: %w", err)
	}
	if len(manifest.Assets) == 0 {
		return coffeupdate.Manifest{}, ErrInvalidChecksums
	}
	return manifest, nil
}

func ValidatePlatforms(manifest coffeupdate.Manifest) error {
	for _, platform := range []string{
		"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64", "windows-arm64",
	} {
		if _, ok := manifest.Assets[platform]; !ok {
			return fmt.Errorf("platform %s: %w", platform, ErrInvalidChecksums)
		}
	}
	return nil
}

type GenerateOptions struct {
	ChecksumsPath string
	OutputPath    string
	Version       string
	BaseURL       string
}

func Generate(options GenerateOptions) (err error) {
	checksums, err := os.Open(options.ChecksumsPath)
	if err != nil {
		return fmt.Errorf("open checksums: %w", err)
	}
	defer func() { err = errors.Join(err, checksums.Close()) }()
	manifest, err := ParseChecksums(checksums, ManifestOptions{Version: options.Version, BaseURL: options.BaseURL})
	if err != nil {
		return err
	}
	if err := ValidatePlatforms(manifest); err != nil {
		return err
	}
	output, err := os.OpenFile(options.OutputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create manifest: %w", err)
	}
	defer func() { err = errors.Join(err, output.Close()) }()
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	return nil
}
