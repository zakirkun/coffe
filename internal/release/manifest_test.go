package release_test

import (
	"strings"
	"testing"

	cofferelease "coffe/internal/release"
)

func Test_ParseChecksums_selects_raw_platform_binaries(t *testing.T) {
	// Given
	checksums := strings.NewReader("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  coffe_update_v1.2.3_linux_amd64\n" +
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb  coffe_v1.2.3_linux_amd64.tar.gz\n")

	// When
	manifest, err := cofferelease.ParseChecksums(checksums, cofferelease.ManifestOptions{
		Version: "v1.2.3", BaseURL: "https://github.com/example/coffe/releases/download/v1.2.3",
	})
	// Then
	if err != nil {
		t.Fatalf("ParseChecksums() error = %v", err)
	}
	asset, ok := manifest.Assets["linux-amd64"]
	if !ok || asset.SHA256[0] != 'a' {
		t.Fatalf("manifest assets = %v", manifest.Assets)
	}
}
