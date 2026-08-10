package backup

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestReplaceDatabasePreservesExistingPredecessorName(t *testing.T) {
	// Given
	directory := t.TempDir()
	databasePath := filepath.Join(directory, "coffe.db")
	createTestDatabase(t, databasePath)
	replacementPath := filepath.Join(directory, "replacement.db")
	createTestDatabase(t, replacementPath)
	predecessorPath := databasePath + ".before-import-" + strconv.FormatInt(time.Now().Unix(), 10)
	if err := os.WriteFile(predecessorPath, []byte("existing predecessor"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// When
	err := replaceDatabase(context.Background(), databasePath, replacementPath)

	// Then
	if err != nil {
		t.Fatalf("replaceDatabase() error = %v", err)
	}
	content, err := os.ReadFile(predecessorPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "existing predecessor" {
		t.Fatalf("existing predecessor = %q", content)
	}
	preserved, err := filepath.Glob(databasePath + ".before-import-*")
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(preserved) != 2 {
		t.Fatalf("preserved databases = %v", preserved)
	}
}

func createTestDatabase(t *testing.T, path string) {
	t.Helper()
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	if _, err := database.Exec(`CREATE TABLE state (value TEXT)`); err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
