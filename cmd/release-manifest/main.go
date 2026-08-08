package main

import (
	"flag"
	"log/slog"
	"os"

	cofferelease "coffe/internal/release"
)

func main() {
	checksums := flag.String("checksums", "dist/checksums.txt", "GoReleaser checksums file")
	output := flag.String("out", "dist/manifest.json", "output manifest")
	version := flag.String("version", "", "release tag, for example v1.2.3")
	baseURL := flag.String("base-url", "", "HTTPS release asset base URL")
	flag.Parse()
	if err := cofferelease.Generate(cofferelease.GenerateOptions{
		ChecksumsPath: *checksums, OutputPath: *output, Version: *version, BaseURL: *baseURL,
	}); err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("manifest generation failed", slog.Any("err", err))
		os.Exit(1)
	}
}
