package backup_test

import (
	"coffe/internal/backup"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func Test_ExportImport_roundtrips_encrypted_SQLite_snapshot(t *testing.T) {
	// Given
	root := t.TempDir()
	source := filepath.Join(root, "source.db")
	destination := filepath.Join(root, "restored.db")
	archive := filepath.Join(root, "backup.age")
	database, err := sql.Open("sqlite", source)
	requireNoError(t, "sql.Open", err)
	_, err = database.Exec(`CREATE TABLE state (value TEXT); INSERT INTO state VALUES ('secret-state')`)
	requireNoError(t, "Exec", err)
	requireNoError(t, "Close", database.Close())
	oldDatabase, err := sql.Open("sqlite", destination)
	requireNoError(t, "sql.Open(old)", err)
	_, err = oldDatabase.Exec(`CREATE TABLE old_state (value TEXT)`)
	requireNoError(t, "Exec(old)", err)
	requireNoError(t, "Close(old)", oldDatabase.Close())
	passphrase := "correct horse battery staple"

	// When
	err = backup.Export(context.Background(), backup.ExportOptions{
		DatabasePath: source, OutputPath: archive, Passphrase: passphrase,
	})
	requireNoError(t, "Export", err)
	err = backup.Import(context.Background(), backup.ImportOptions{
		DatabasePath: destination, InputPath: archive, Passphrase: passphrase,
	})
	requireNoError(t, "Import", err)
	err = backup.Import(context.Background(), backup.ImportOptions{
		DatabasePath: destination, InputPath: archive, Passphrase: passphrase,
	})
	requireNoError(t, "Import(repeated)", err)

	// Then
	encrypted, err := os.ReadFile(archive) //nolint:gosec // archive is created inside t.TempDir.
	requireNoError(t, "ReadFile", err)
	if strings.Contains(string(encrypted), "secret-state") {
		t.Fatal("encrypted backup contains plaintext")
	}
	restored, err := sql.Open("sqlite", destination)
	requireNoError(t, "sql.Open(restored)", err)
	t.Cleanup(func() {
		if closeErr := restored.Close(); closeErr != nil {
			t.Errorf("Close(restored) error = %v", closeErr)
		}
	})
	var value string
	requireNoError(t, "QueryRow", restored.QueryRow(`SELECT value FROM state`).Scan(&value))
	if value != "secret-state" {
		t.Fatalf("restored value = %q", value)
	}
	preserved, err := filepath.Glob(destination + ".before-import-*")
	requireNoError(t, "Glob", err)
	if len(preserved) != 2 {
		t.Fatalf("preserved databases = %v", preserved)
	}
}

func requireNoError(t *testing.T, operation string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s() error = %v", operation, err)
	}
}
