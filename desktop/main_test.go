//go:build wailsapp

package main

import (
	"path/filepath"
	"testing"
)

func TestDesktopOptions_whenDatabaseEnvironmentSet_thenUsesAbsolutePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "coffe.db")
	t.Setenv("COFFE_DB", path)

	options, err := desktopOptions()
	if err != nil {
		t.Fatalf("desktop options: %v", err)
	}
	if options.DatabasePath != path {
		t.Errorf("database path = %q, want %q", options.DatabasePath, path)
	}
}

func TestDesktopOptions_whenDatabaseEnvironmentRelative_thenRejects(t *testing.T) {
	t.Setenv("COFFE_DB", "relative/coffe.db")

	_, err := desktopOptions()

	if err == nil {
		t.Fatal("desktop options succeeded, want absolute-path error")
	}
}
