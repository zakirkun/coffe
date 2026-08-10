//go:build wailsapp

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteExclusive_whenPathIsNew_thenWritesContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.csv")

	if err := writeExclusive(path, []byte("category,outcome\n")); err != nil {
		t.Fatalf("write export: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if string(content) != "category,outcome\n" {
		t.Errorf("content = %q", content)
	}
}

func TestWriteExclusive_whenPathExists_thenPreservesContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.csv")
	if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
		t.Fatalf("seed export: %v", err)
	}

	if err := writeExclusive(path, []byte("replacement")); err == nil {
		t.Fatal("write export succeeded, want exclusive-create error")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if string(content) != "original" {
		t.Errorf("content = %q, want original", content)
	}
}
